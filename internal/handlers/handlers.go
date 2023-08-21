package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/eugene982/url-shortener/internal/middleware"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/shortener"
)

type BaseURLGetter interface {
	GetBaseURL() string
}

type Pinger interface {
	Ping(context.Context) error
}

type Setter interface {
	Set(ctx context.Context, data model.StoreData) error
}

type AddrGetter interface {
	GetAddr(context.Context, string) (model.StoreData, error)
}

type UserURLGetter interface {
	GetUserURLs(context.Context, string) ([]model.StoreData, error)
}

type UserShortAsyncDeleter interface {
	DeleteUserShortAsync(userID string, shorts []string)
}

type Updater interface {
	Update(ctx context.Context, list []model.StoreData) error
}

// проверка заголовка на формат
func CheckContentType(value string, r *http.Request) (bool, error) {
	if strings.Contains(r.Header.Get("Content-Type"), value) {
		return true, nil
	}
	return false, fmt.Errorf("Content-Type: %s not found", value)
}

// ищем или пытаемся создать короткую ссылку
func GetAndWriteShort(sh shortener.Shortener, setter Setter, addr string, r *http.Request) (string, error) {

	short, err := sh.Short(addr)
	if err != nil {
		return "", err
	}

	userID, err := middleware.GetUserID(r)
	if err != nil {
		return "", err
	}

	data := model.StoreData{
		UserID:      userID,
		ShortURL:    short,
		OriginalURL: addr,
	}

	if ok, err := data.IsValid(); !ok {
		return "", err
	}

	// запись в файловое хранилище
	return short, setter.Set(r.Context(), data)
}
