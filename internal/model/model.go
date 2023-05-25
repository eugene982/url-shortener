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
