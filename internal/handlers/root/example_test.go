package root

import (
	"fmt"
	"net/http/httptest"

	"github.com/eugene982/url-shortener/internal/model"
)

func ExampleNewFindAddrHandler() {

	getter := addGetterFunc(func() (model.StoreData, error) {
		return model.StoreData{
			OriginalURL: "ya.ru"}, nil
	})

	handler := NewFindAddrHandler(getter)

	r := httptest.NewRequest("GET", "/ya.ru", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Location"))

	// Output:
	// 307
	// ya.ru

}
