// Хранилище раннее сгенерированных ссылок.
// построен на мапе.
// Удовлетворяет интерфейсу "Storage"
package memstore

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
)

// Объявление структуры-хранителя
type MemStore struct {
	addrList   map[string]model.StoreData
	savingAddr map[string]bool
	fs         *fileStorage // запись во временный файл
}

// Утверждение типа, ошибка компиляции
var _ storage.Storage = (*MemStore)(nil)

// Функция-конструктор нового хранилща
func New(fname string) (*MemStore, error) {

	var (
		err        error
		fs         *fileStorage
		addrList   = make(map[string]model.StoreData)
		savingAddr = make(map[string]bool)
	)

	// хранение ранее созданных сокращений в файле
	// для восстановления после перезапуска.
	if fname != "" {
		if fs, err = newFileSorage(fname); err != nil {
			return nil, fmt.Errorf("error open file storage: %w", err)
		}

		urls, err := fs.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("error read from file storage: %w", err)
		}
		// переносим все ранее сохранённые значения из файла
		for _, v := range urls {
			addrList[v.ShortURL] = model.StoreData{
				ShortURL:    v.ShortURL,
				OriginalURL: v.OriginalURL,
			}
			savingAddr[v.OriginalURL] = true
		}
	}

	ms := &MemStore{
		addrList:   addrList,
		savingAddr: savingAddr,
		fs:         fs,
	}
	return ms, nil
}

func (m *MemStore) Close() error {
	if err := m.fs.Close(); err != nil {
		return fmt.Errorf("error close file storage: %w", err)
	}
	return nil
}

// Проверка соединения до хранилища
func (m *MemStore) Ping(ctx context.Context) (err error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Получение полного адреса по короткой ссылке
func (m *MemStore) GetAddr(ctx context.Context, short string) (addr string, err error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if addr, ok := m.addrList[short]; ok {
		return addr.OriginalURL, nil
	}
	return "", storage.ErrAddressNotFound
}

// Установка уникального соответствия
func (m *MemStore) Set(ctx context.Context, data model.StoreData) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Проверка на налицие сохранённого полного адреса
	if m.savingAddr[data.OriginalURL] {
		return storage.ErrAddressConflict
	}

	if ok, err := data.IsValid(); !ok {
		return err
	}

	list := []model.StoreData{
		data,
	}

	return m.Update(ctx, list)
}

// Установка/обновление соответствиq между адресом и короткой ссылкой
func (m *MemStore) Update(ctx context.Context, list []model.StoreData) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := m.fs.Append(list); err != nil {
		return err
	}

	for _, d := range list {
		m.addrList[d.ShortURL] = d
		m.savingAddr[d.OriginalURL] = true
	}
	return nil
}

// Получение данных пользователя
func (m *MemStore) GetUserURLs(ctx context.Context, userID int64) ([]model.StoreData, error) {
	res := make([]model.StoreData, 0)
	for _, v := range m.addrList {
		if v.UserID == userID {
			res = append(res, v)
		}
	}
	return res, nil
}

// Временное хранилище адресов на диске
type fileStorage struct {
	file    *os.File
	writer  *bufio.Writer // ожидается, что записывать будем чаще чем записывать.
	counter int
}

// Создание нового файла хранилища
func newFileSorage(fname string) (*fileStorage, error) {
	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	writter := bufio.NewWriter(file)

	return &fileStorage{
		file:   file,
		writer: writter,
	}, nil
}

// закрытие  файла
func (fs *fileStorage) Close() error {
	if fs == nil {
		return nil
	}
	return fs.Close()
}

// чтение всех ранее сохраненных данных
func (fs *fileStorage) ReadAll() ([]model.StoreData, error) {
	if fs == nil {
		return nil, nil
	}

	res := make([]model.StoreData, 0, 8)
	scanner := bufio.NewScanner(fs.file)

	var data model.StoreData
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			return nil, err
		}
		res = append(res, data)
	}

	fs.counter = len(res)
	return res, nil
}

// Добавление новых данных
func (fs *fileStorage) Append(data []model.StoreData) error {
	if fs == nil {
		return nil
	}

	for _, d := range data {
		fs.counter++

		if d.ID == "" {
			d.ID = strconv.Itoa(fs.counter)
		}
		err := json.NewEncoder(fs.writer).Encode(&d)
		if err != nil {
			return err
		}
	}
	return fs.writer.Flush()
}
