package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Url struct {
	Long  string `json:"long"`
	Short string `json:"short"`
}

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	var data Url
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if data.Long == "" {
		http.Error(w, "Url is required", http.StatusBadRequest)
		return
	}
	result := generateKey(67)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func generateKey(n int) string {
	const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	return "abc123"
}

func main() {
	http.HandleFunc("/shorten", ShortenHandler)

	fmt.Println("Listening and serving on:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}

//что бы получить короткую строку из числа и строки нам необходимо использовать base62 encoding.
//что бы из числа получить строку нам сначала надо вычислить остаток от деления каждой цифры в числе. затем мы возводим в степень 62 на остаток. получаем сумму
//а с строкой мы делаем следующее : мы сначала получаем байтовое значение, потом переводим это в hex, затем его в base62
