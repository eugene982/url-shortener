// Роутер - на основе внешней библиотеки "chi"
package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Возвращает роутер
func (a *Application) NewRouter() http.Handler {

	r := chi.NewRouter()

	r.Get("/{short}", a.getAddr)

	r.Post("/", a.postAddr)
	r.Post("/api/shorten", a.postAddrJson)

	// во всех остальных случаях 404
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		errorNotFound(fmt.Errorf("method not allowed"), w, r)
	})

	return r
}

// Получение полного адреса по короткой ссылке
func (a *Application) getAddr(w http.ResponseWriter, r *http.Request) {

	short := chi.URLParam(r, "short")
	addr, ok := a.store.GetAddr(short)
	if !ok {
		errorNotFound(fmt.Errorf("short address %s not found", short), w, r)
		return
	}

	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
func (a *Application) postAddr(w http.ResponseWriter, r *http.Request) {

	// чтобы длиинно не писать...
	hasErr := func(e error) bool { return errorNotFound(e, w, r) }

	err := checkContentType("text/plain", r)
	if hasErr(err) {
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close() // Вроде как надо закрывать если что-то там есть...
	if hasErr(err) {
		return
	}

	short, err := a.findOrGetShort(string(body), r)
	if hasErr(err) {
		return
	}

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
}

// Откуда-то в тестах взялся JSON?? Не видел я задания с ним.
// работаем по ситуации...
func (a *Application) postAddrJson(w http.ResponseWriter, r *http.Request) {

	hasErr := func(e error) bool { return errorNotFound(e, w, r) }

	err := checkContentType("application/json", r)
	if hasErr(err) {
		return
	}

	// запрос
	request := struct {
		URL string `json:"url"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&request)
	defer r.Body.Close()
	if hasErr(err) {
		return
	}

	short, err := a.findOrGetShort(string(request.URL), r)
	if hasErr(err) {
		return
	}

	// ответ
	result := struct {
		Result string `json:"result"`
	}{short}

	// не двигать!!!
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(result)
	hasErr(err)
}

// при ошибке всегда возвращаем 404
func errorNotFound(err error, w http.ResponseWriter, r *http.Request) bool {
	if err == nil {
		return false
	}
	log.Println("error:", err)
	http.NotFound(w, r)
	return true
}

// проверка заголовка на формат
func checkContentType(value string, r *http.Request) error {
	if strings.Contains(r.Header.Get("Content-Type"), value) {
		return nil
	}
	return fmt.Errorf("Content-Type: %s not found", value)
}

// ищем или пытаемся создать короткую ссылку
func (a *Application) findOrGetShort(addr string, r *http.Request) (string, error) {

	if addr == "" {
		return "", fmt.Errorf("address is empty")
	}

	// Добавляем в хранилилище если ранее не было добавлено
	short, ok := a.store.GetShort(addr)
	if !ok {
		short = a.shortener.Short(addr)
		if !a.store.Set(addr, short) { // не удалось поместить, например пустые строки.
			return "", fmt.Errorf("error short address %s", addr)
		}
	}

	if a.baseUrl == "" {
		return "http://" + r.Host + "/" + short, nil
	} else {
		return a.baseUrl + short, nil
	}
}
