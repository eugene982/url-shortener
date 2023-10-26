// Хранилище раннее сгенерированных ссылок.
// построен на мапе.
// Удовлетворяет интерфейсу "Storage"
package memstore

import (
	"context"
	"fmt"

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
func (m *MemStore) GetAddr(ctx context.Context, short string) (data model.StoreData, err error) {
	select {
	case <-ctx.Done():
		return model.StoreData{}, ctx.Err()
	default:
	}

	if data, ok := m.addrList[short]; ok {
		return data, nil
	}
	return model.StoreData{}, storage.ErrAddressNotFound
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
func (m *MemStore) GetUserURLs(ctx context.Context, userID string) ([]model.StoreData, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	res := make([]model.StoreData, 0)
	for _, v := range m.addrList {
		if v.UserID == userID {
			res = append(res, v)
		}
	}
	return res, nil
}

// Пометка на удаление
func (m *MemStore) DeleteShort(ctx context.Context, shortURLs []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, short := range shortURLs {
		data, ok := m.addrList[short]
		if ok {
			data.DeletedFlag = true
			m.addrList[short] = data
		}
	}

	return nil
}

// Пометка на удаление
func (m *MemStore) Stats(ctx context.Context) (URLs int, users int, err error) {
	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	default:
	}

	URLs = len(m.addrList)

	// пользователи могут повторяться, группируем
	usrGrp := make(map[string]bool, URLs)
	for _, d := range m.addrList {
		usrGrp[d.UserID] = true
	}
	users = len(usrGrp)
	return
}
