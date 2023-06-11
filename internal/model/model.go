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
type StoreData struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// запрос на добавление POST /api/shorten/batch
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// Валидация входящей стуктуры
func (br BatchRequest) IsValid() (bool, error) {
	if strings.TrimSpace(br.CorrelationID) == "" {
		return false, fmt.Errorf("correlation ID is empty")
	}
	if strings.TrimSpace(br.OriginalURL) == "" {
		return false, fmt.Errorf("original URL is empty")
	}
	return true, nil
}

// ответ на добавление POST /api/shorten/batch
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
