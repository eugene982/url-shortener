// Роутер - на основе внешней библиотеки "chi"

package app

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Возвращает роутер
func (a *Application) NewRouter() http.Handler {

	r := chi.NewRouter()

	r.Get("/{short}", a.getAddr)
	r.Post("/", a.postAddr)

	// во всех остальных случаях 404
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		errorNotFound(false, w, r)
	})

	return r
}

// Получение полного адреса по короткой ссылке
func (a *Application) getAddr(w http.ResponseWriter, r *http.Request) {

	short := chi.URLParam(r, "short")
	addr, ok := a.store.GetAddr(short)
	if errorNotFound(ok, w, r) {
		return
	}

	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
func (a *Application) postAddr(w http.ResponseWriter, r *http.Request) {

	// чтобы длиинно не писать...
	notOk := func(ok bool) bool { return errorNotFound(ok, w, r) }

	if notOk(strings.Contains(r.Header.Get("Content-Type"), "text/plain")) {
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close() // Вроде как надо закрывать если что-то там есть...
	if notOk(err == nil) {
		return
	}

	// Добавляем в хранилилище если ранее не было добавлено
	addr := string(body)
	short, ok := a.store.GetShort(addr)
	if !ok {
		short = a.shortener.Short(addr)
		if notOk(a.store.Set(addr, short)) { // не удалось поместить, например пустые строки.
			return
		}
	}

	short = `http://` + r.Host + `/` + short
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
}

// при ошибке всегда возвращаем 404
func errorNotFound(ok bool, w http.ResponseWriter, r *http.Request) bool {
	if ok {
		return false
	}
	http.NotFound(w, r)
	return true
}
