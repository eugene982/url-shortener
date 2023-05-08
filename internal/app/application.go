// Пакет приложения. на данном этапе объём кода небольшой,
// поэтому все компоненты сложены в один пакет.
// В дальнейшем их можно будет разложить по собственным пакетам (route, store...)
package app

import (
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
	shortener Shortener
	store     Storage
	baseAddr  string
}

// Функция конструктор приложения.
func NewApplication(shortener Shortener, store Storage, baseAddr string) *Application {
	if baseAddr != "" && !strings.HasSuffix(baseAddr, "/") {
		baseAddr += "/"
	}
	return &Application{shortener, store, baseAddr}
}
