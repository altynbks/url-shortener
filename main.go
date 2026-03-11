package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
)

type Url struct {
	Long  string `json:"long"`
	Short string `json:"short"`
}

type UrlStore struct {
	urls map[string]string
	mu   sync.RWMutex
}

type Server struct {
	store *UrlStore
}

func (s *Server) ShortenHandler(w http.ResponseWriter, r *http.Request) {
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

	key := s.store.Set(data.Long)
	data.Short = key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) < 4 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	shortKey := r.URL.Path[3:]
	longURL, ok := s.store.Get(shortKey)
	if !ok {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

func generateKey(n int) string {
	alphanumeric := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	res := make([]byte, n)
	for i := range res {
		randomIndex := rand.IntN(len(alphanumeric))
		selected := alphanumeric[randomIndex]
		res[i] = selected
	}
	return string(res)
}

func NewUrlStore() *UrlStore {
	return &UrlStore{
		urls: make(map[string]string),
	}
}

func (s *UrlStore) Set(longURL string) string {
	// 1. Заблокируй на запись
	s.mu.Lock()
	defer s.mu.Unlock()
	// 2. Сгенерируй ключ через свою функцию generateKey
	key := generateKey(6)
	// 3. Сохрани в s.urls
	s.urls[key] = longURL
	// 4. Верни ключ
	return key
}

func (s *UrlStore) Get(shortKey string) (string, bool) {
	// 1. Заблокируй на ЧТЕНИЕ (RLock)
	s.mu.RLock()
	defer s.mu.RUnlock()
	// 2. Достань ссылку из мапы
	val, ok := s.urls[shortKey]
	if ok {
		return val, true
	}
	// 3. Верни ссылку и флаг "найдено/не найдено"
	return "", false
}

func main() {
	store := NewUrlStore()

	srv := &Server{
		store: store,
	}
	http.HandleFunc("/shorten", srv.ShortenHandler)
	http.HandleFunc("/r/", srv.RedirectHandler)

	fmt.Println("Listening and serving on:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
