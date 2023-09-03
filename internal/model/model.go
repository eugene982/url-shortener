package model

import (
	"fmt"
	"strings"
)

// Структура запроса /api/shorten
type RequestShorten struct {
	URL string `json:"url"`
}

// ResponseShorten Структура ответа /api/shorten
type ResponseShorten struct {
	Result string `json:"result"`
}

// IsValid валидация полей структуры RequestShorten
func (req RequestShorten) IsValid() (bool, error) {
	if strings.TrimSpace(req.URL) == "" {
		return false, fmt.Errorf("url is empty")
	}
	return true, nil
}

// StoreData данные для хранения в файловом хранилище
type StoreData struct {
	ID          string `json:"uuid"`
	UserID      string `json:"user_id" db:"user_id"`
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"origin_url"`
	DeletedFlag bool   `json:"is_deleted" db:"is_deleted"`
}

// IsValid валидация полей структуры StoreData
func (st StoreData) IsValid() (bool, error) {
	if strings.TrimSpace(st.ShortURL) == "" {
		return false, fmt.Errorf("short URL is empty")
	}
	if strings.TrimSpace(st.OriginalURL) == "" {
		return false, fmt.Errorf("original URL is empty")
	}
	return true, nil
}

// BatchRequest запрос на добавление POST /api/shorten/batch
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// IsValid валидация полей входящей стуктуры BatchRequest
func (br BatchRequest) IsValid() (bool, error) {
	if strings.TrimSpace(br.CorrelationID) == "" {
		return false, fmt.Errorf("correlation ID is empty")
	}
	if strings.TrimSpace(br.OriginalURL) == "" {
		return false, fmt.Errorf("original URL is empty")
	}
	return true, nil
}

// BatchResponse ответ на добавление POST /api/shorten/batch
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLResponse ответ возвращает ссылки пользователя.
type UserURLResponse struct {
	UserID      int64
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}
