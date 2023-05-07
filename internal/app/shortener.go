// "Сокращатель" ссылок, построен на снове алгоритма crc64,
// доступного в стандартной библиотеке.
// Удовлетворяет интерфейсу "Shortener"

package app

import (
	"hash/crc64"
)

type SimpleShortener struct {
	symTab  []byte       // симфолы для хеша
	hashLen int          // размер сокращения
	crcTab  *crc64.Table // для контрольной суммы
}

// Утверждение типа, ошибка компиляции
var _ Shortener = (*SimpleShortener)(nil)

// Функция-конструктор
func NewSimpleShortener() *SimpleShortener {
	// создаём таблицу символов которые будем использовать в хеше
	symTab := make([]byte, 0, 64)

	// заполняем символами которые учавствуют в хеше
	var c byte
	for c = '0'; c <= '9'; c++ {
		symTab = append(symTab, c)
	}
	for c = 'a'; c <= 'z'; c++ {
		symTab = append(symTab, c)
	}
	for c = 'A'; c <= 'Z'; c++ {
		symTab = append(symTab, c)
	}

	return &SimpleShortener{
		symTab,
		10, // сокращаем до 10 символов
		crc64.MakeTable(crc64.ISO),
	}
}

// Будем сохкращать строку до 10 символов
func (s *SimpleShortener) Short(addr string) string {
	sum := crc64.Checksum([]byte(addr), s.crcTab)
	return s.hash(sum)
}

// Функция хеширует число до нужной длинны
func (s *SimpleShortener) hash(val uint64) string {
	buff := make([]byte, s.hashLen) // можно вынести в структуру, чтоб каждый раз не пересоздавать.

	reminder := val
	size := len(s.symTab)

	for i := 0; i < s.hashLen; i++ {
		buff[i] = s.symTab[(reminder % uint64(size))]
		reminder /= uint64(size)
	}
	return string(buff)
}
