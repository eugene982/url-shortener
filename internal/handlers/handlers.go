package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/eugene982/url-shortener/gen/go/proto"
	"github.com/eugene982/url-shortener/internal/middleware"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/shortener"
)

// Pinger интерфейс проверки связи с сервисом.
type Pinger interface {
	Ping(context.Context) error
}

// Setter интерфейс записи данных в хранилище.
type Setter interface {
	Set(ctx context.Context, data model.StoreData) error
}

// AddrGetter интерфейс получения полного адреса по короткой ссылке.
type AddrGetter interface {
	GetAddr(context.Context, string) (model.StoreData, error)
}

// UserURLGetter интерфейс получения сохраннённых ссылок пользователя.
type UserURLGetter interface {
	GetUserURLs(context.Context, string) ([]model.StoreData, error)
}

// UserShortAsyncDeleter интерфейс асинхронного удаления ссылок пользователя.
type UserShortAsyncDeleter interface {
	DeleteUserShortAsync(userID string, shorts []string)
}

// Updater интерфейс обновления данных в хранилище.
type Updater interface {
	Update(ctx context.Context, list []model.StoreData) error
}

// StatsGetter интерфейс получения сведений о статистике
type StatsGetter interface {
	Stats(ctx context.Context) (URLs int, users int, err error)
}

// CheckContentType проверка заголовка запроса на формат.
func CheckContentType(value string, r *http.Request) (bool, error) {
	if strings.Contains(r.Header.Get("Content-Type"), value) {
		return true, nil
	}
	return false, fmt.Errorf("Content-Type: %s not found", value)
}

// GetAndWriteShort ищем или пытаемся создать короткую ссылку.
func GetAndWriteShort(sh shortener.Shortener, setter Setter, addr string, r *http.Request) (string, error) {

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		return "", err
	}

	return GetAndWriteUserShort(r.Context(), sh, setter, userID, addr)
}

// GetAndWriteUserShort - запись пользовательской ссылки
func GetAndWriteUserShort(ctx context.Context, sh shortener.Shortener, setter Setter, userID, addr string) (string, error) {

	short, err := sh.Short(addr)
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
	return short, setter.Set(ctx, data)
}

// gRPC

type PingHandler func(context.Context, *empty.Empty) (*proto.PingResponse, error)
type FindAddrHandler func(context.Context, *proto.FindAddrRequest) (*proto.FindAddrResponse, error)
type CreateShortHandler func(context.Context, *proto.CreateShortRequest) (*proto.CreateShortResponse, error)
type BatchShortHandler func(context.Context, *proto.BatchRequest) (*proto.BatchResponse, error)
type GetUserURLsHandler func(context.Context, *proto.UserURLsRequest) (*proto.UserURLsResponse, error)
type DelUserURLsHandler func(context.Context, *proto.DelUserURLsRequest) (*empty.Empty, error)
