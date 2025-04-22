package main

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
)

var urlStorage = make(map[string]string)

const baseURL string = "http://localhost:8080/"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortenURL)
	mux.HandleFunc("/{shortCode}", redirectByShortURL)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

func shortenURL(w http.ResponseWriter, r *http.Request) {
	// проверяем метод запроса
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// читаем тело запроса
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	urlStr := strings.TrimSpace(string(originalURL))

	// проверяем что URL не пустой
	if urlStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL cannot be empty"))
		return
	}

	// генерируем короткий код
	hash := sha256.Sum256([]byte(urlStr))
	shortCode := base64.URLEncoding.EncodeToString(hash[:6])
	shortCode = strings.TrimRight(shortCode, "=")

	// сохраняем в хранилище
	shortURL := baseURL + shortCode
	urlStorage[shortCode] = urlStr

	// возвращаем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func redirectByShortURL(w http.ResponseWriter, r *http.Request) {
	// проверяем метод запроса
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/")

	if shortCode == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL, ok := urlStorage[shortCode]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
