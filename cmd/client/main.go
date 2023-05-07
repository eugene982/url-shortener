// Консольный клиент
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var endpoint string

func main() {

	flag.StringVar(&endpoint, "a", "http://localhost:8080", "адрес сервера")
	flag.Parse()

	// 	контейнер данных для запроса
	data := url.Values{}

	// приглашение в консоли
	fmt.Println("Введите длинный URL")

	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	//читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSuffix(long, "\n")

	// 	заполняем контейнер данными
	data.Set("url", long)
	// 	добавляем клиент
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// Пишем запрос
	// ПОСТ запрос должен содержатьи  заголовок и тело
	request, err := http.NewRequest(http.MethodPost, endpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	// в заголовках указываем кодировку
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	// выводим код ответа
	fmt.Println("Статус-код", response.Status)
	defer response.Body.Close()

	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// печатем
	fmt.Println(string(body))
}
