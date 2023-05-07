// Хранилище раннее сгенерированных ссылок.
// построен на мапе.
// Удовлетворяет интерфейсу "Storage"

package app

// Объявление структуры-хранителя
type MemStore struct {
	addrList  map[string]string
	shortList map[string]string
}

// Утверждение типа, ошибка компиляции
var _ Storage = (*MemStore)(nil)

// Функция-конструктор нового хранилща
func NewMemstore() *MemStore {
	return &MemStore{
		make(map[string]string, 8),
		make(map[string]string, 8),
	}
}

// Получение короткой ссылки по полному адресу
func (m *MemStore) GetShort(addr string) (short string, ok bool) {
	short, ok = m.shortList[addr]
	return
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

	// если по данной короткой ссылке уже содержатся данные
	// необходимо их почистить, чтоб не занимали место
	if old, ok := m.addrList[short]; ok {
		delete(m.shortList, old)
	}

	m.addrList[short] = addr
	m.shortList[addr] = short
	return true
}
