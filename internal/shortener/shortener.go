// "Сокращатель" ссылок, построен на снове алгоритма crc64,
// доступного в стандартной библиотеке.
// Удовлетворяет интерфейсу "Shortener"
package shortener

import (
	"fmt"
	"hash/crc64"
)

const (
	hashLen = 10 // размер сокращения
)

// Сокращатель ссылок
type Shortener interface {
	Short(string) (string, error)
}

type SimpleShortener struct {
	symTab []byte       // символы для хеша
	crcTab *crc64.Table // для контрольной суммы

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
		crc64.MakeTable(crc64.ISO),
	}
}

// Будем сохкращать строку до 10 символов
func (s *SimpleShortener) Short(addr string) (short string, err error) {
	sum := crc64.Checksum([]byte(addr), s.crcTab)
	short = s.toString(sum)
	if short == "" {
		err = fmt.Errorf("cannot generate short url %s", addr)
	}
	return
}

// Функция преобразует хэш в строку нужной длинны
// Пробовал на основе base64 но случаются коллизии т.к. начало строк часто совпадают.
func (s *SimpleShortener) toString(val uint64) string {
	var buff [hashLen]byte

	reminder := val
	size := len(s.symTab)

	for i := 0; i < hashLen; i++ {
		buff[i] = s.symTab[(reminder % uint64(size))]
		reminder /= uint64(size)
	}
	return string(buff[:])
}
