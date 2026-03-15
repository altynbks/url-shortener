package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type URL struct {
	Long  string `json:"long"`
	Short string `json:"short"`
}

type URLStore struct {
	conn *pgx.Conn
}

type Server struct {
	store *URLStore
}

func (s *Server) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var data URL
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if data.Long == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	key := s.store.Set(data.Long)
	if key == "" {
		http.Error(w, "failed to create short URL", http.StatusInternalServerError)
		return
	}

	data.Short = key

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) <= 3 {
		http.Error(w, "invalid short URL", http.StatusBadRequest)
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
		res[i] = alphanumeric[randomIndex]
	}

	return string(res)
}

func NewURLStore(conn *pgx.Conn) *URLStore {
	return &URLStore{
		conn: conn,
	}
}

func (s *URLStore) Set(longURL string) string {
	key := generateKey(6)

	query := `INSERT INTO urls (short_key, long_url) VALUES ($1, $2)`

	_, err := s.conn.Exec(context.Background(), query, key, longURL)
	if err != nil {
		fmt.Printf("database insert error: %v\n", err)
		return ""
	}

	return key
}

func (s *URLStore) Get(shortKey string) (string, bool) {
	var longURL string

	query := `SELECT long_url FROM urls WHERE short_key = $1`

	err := s.conn.QueryRow(context.Background(), query, shortKey).Scan(&longURL)
	if err != nil {
		return "", false
	}

	return longURL, true
}

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		fmt.Println("DATABASE_URL environment variable is not set")
		return
	}
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		fmt.Printf("failed to connect to the database: %v\n", err)
		return
	}
	defer conn.Close(ctx)

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS urls (
        id SERIAL PRIMARY KEY,
        short_key TEXT UNIQUE NOT NULL,
        long_url TEXT NOT NULL
    );`

	_, err = conn.Exec(ctx, createTableSQL)
	if err != nil {
		fmt.Printf("Error creating the database: %v\n", err)
		return
	}

	store := NewURLStore(conn)
	srv := &Server{store: store}

	http.HandleFunc("/shorten", srv.ShortenHandler)
	http.HandleFunc("/r/", srv.RedirectHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	port := os.Getenv("PORT")
	if port == "" {
	    port = "8080" 
	}
	
	fmt.Printf("Server is running on :%s\n", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
	    fmt.Println("server error:", err)
	}
}
