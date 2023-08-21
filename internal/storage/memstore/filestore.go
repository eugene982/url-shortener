package memstore

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"

	"github.com/eugene982/url-shortener/internal/model"
)

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
	return fs.file.Close()
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
