package main

import (
	"fmt"
	"net/http"
)

type Url struct {
	Long  string `json:"long"`
	Short string `json:"short"`
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World!")
}

func main() {
	http.HandleFunc("/", HelloHandler)

	fmt.Println("Listening and serving on:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}

//что бы получить короткую строку из числа и строки нам необходимо использовать base62 encoding.
//что бы из числа получить строку нам сначала надо вычислить остаток от деления каждой цифры в числе. затем мы возводим в степень 62 на остаток. получаем сумму
//а с строкой мы делаем следующее : мы сначала получаем байтовое значение, потом переводим это в hex, затем его в base62
