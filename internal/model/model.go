package model

import (
	"fmt"
	"strings"
)

// Структура запроса /api/shorten
type RequestShorten struct {
	URL string `json:"url"`
}

// Структура ответа /api/shorten
type ResponseShorten struct {
	Result string `json:"result"`
}

// Валидация
func (req RequestShorten) IsValid() (bool, error) {
	if strings.TrimSpace(req.URL) == "" {
		return false, fmt.Errorf("url is empty")
	}
	return true, nil
}

// Данные для хранения в файловом хранилище
type FileStoreData struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
