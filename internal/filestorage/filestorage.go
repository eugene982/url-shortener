package filestorage

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"

	"github.com/eugene982/url-shortener/internal/model"
)

// Структура для хранилища ссылок
// Может быть не создан!
type FileStorage struct {
	file    *os.File
	writer  *bufio.Writer // ожидается, что записывать будем чаще чем записывать.
	counter int
}

// Создание нового файла хранилища
func New(fname string) (*FileStorage, error) {
	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	writter := bufio.NewWriter(file)

	return &FileStorage{
		file:   file,
		writer: writter,
	}, nil
}

// закрытие  файла
func (fs *FileStorage) Close() error {
	if fs == nil {
		return nil
	}
	return fs.Close()
}

// чтение всех ранее сохраненных данных
func (fs *FileStorage) ReadAll() ([]model.FileStoreData, error) {
	if fs == nil {
		return nil, nil
	}

	res := make([]model.FileStoreData, 0, 8)
	scanner := bufio.NewScanner(fs.file)

	var data model.FileStoreData
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
func (fs *FileStorage) Append(originalURL, shortURL string) error {
	if fs == nil {
		return nil
	}
	fs.counter++

	data := model.FileStoreData{
		ID:          strconv.Itoa(fs.counter),
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}

	writeBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	writeBytes = append(writeBytes, '\n')

	_, err = fs.writer.Write(writeBytes)
	if err != nil {
		return err
	}
	return fs.writer.Flush()
}
