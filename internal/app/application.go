package app

import (
	"bufio"
	"io"
	"net/http"
	"strings"
)

// Сокращатель ссылок
type Shortener interface {
	Short(string) string
}

// Хранитель ссылок
type Storage interface {
	GetAddr(string) (string, bool)
	GetShort(string) (string, bool)
	Set(string, string) bool
}

// Управлятель ссылок
type Application struct {
	handler   http.Handler
	shortener Shortener
	store     Storage
}

// Функция конструктор приложения.
func NewApplication(sh Shortener, st Storage) *Application {

	mux := http.NewServeMux()
	app := &Application{mux, sh, st}

	mux.HandleFunc(`/`, app.rootHandler)
	return app
}

// Возвращает мультиплексор
func (a *Application) GetMux() http.Handler {
	return a.handler
}

// Доступ к корневому пути сервиса
func (a *Application) rootHandler(w http.ResponseWriter, r *http.Request) {

	var ok bool

	switch r.Method {
	case http.MethodGet:
		ok = a.getAddr(w, r)
	case http.MethodPost:
		ok = a.postAddr(w, r)
	}

	if !ok {
		http.NotFound(w, r)
	}
}

// Получение полного адреса по короткой ссылке
func (a *Application) getAddr(w http.ResponseWriter, r *http.Request) bool {

	short := strings.TrimLeft(r.RequestURI, `/`)
	addr, ok := a.store.GetAddr(short)
	if !ok {
		return false
	}

	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
	return true
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
func (a *Application) postAddr(w http.ResponseWriter, r *http.Request) bool {
	defer r.Body.Close() // Вроде как надо закрывать если что-то там есть...

	if r.RequestURI != `/` || !strings.Contains(r.Header.Get("Content-Type"), "text/plain") {
		return false
	}

	addr, err := bufio.NewReader(r.Body).ReadString('\n')
	if err != nil && err != io.EOF {
		return false
	}

	// Добавляем в хранилилище если ранее не было добавлено.
	short, ok := a.store.GetShort(addr)
	if !ok {
		short = a.shortener.Short(addr)
		if !a.store.Set(addr, short) { // не удалось поместить, например пустые строки.
			return false
		}
	}

	short = `http://` + r.Host + `/` + short
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
	return true
}
