// Пакет приложения. на данном этапе объём кода небольшой,
// поэтому все компоненты сложены в один пакет.
// В дальнейшем их можно будет разложить по собственным пакетам (route, store...)
package app

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
}

// Функция конструктор приложения.
func NewApplication(shortener Shortener, store Storage) *Application {
	return &Application{shortener, store}
}
