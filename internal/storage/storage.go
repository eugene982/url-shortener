// Хранилище раннее сгенерированных ссылок.
// построен на мапе.
// Удовлетворяет интерфейсу "Storage"
package storage

import "github.com/eugene982/url-shortener/internal/app"

// Объявление структуры-хранителя
type MemStore struct {
	addrList map[string]string
}

// Утверждение типа, ошибка компиляции
var _ app.Storage = (*MemStore)(nil)

// Функция-конструктор нового хранилща
func NewMemstore() *MemStore {
	return &MemStore{
		make(map[string]string),
	}
}

// Получение полного адреса по короткой ссылке
func (m *MemStore) GetAddr(short string) (addr string, ok bool) {
	addr, ok = m.addrList[short]
	return
}

// Установка соответствия между адресом и короткой ссылкой
func (m *MemStore) Set(addr string, short string) bool {
	if addr == "" || short == "" {
		return false
	}

	m.addrList[short] = addr
	return true
}
