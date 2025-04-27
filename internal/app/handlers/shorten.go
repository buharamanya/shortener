package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/buharamanya/shortener/internal/app/storage"
)

type ShortenHandler struct {
	storage storage.URLStorage
	baseURL string
}

func NewShortenHandler(storage storage.URLStorage, baseURL string) *ShortenHandler {
	return &ShortenHandler{
		storage: storage,
		baseURL: baseURL,
	}
}

func (sh *ShortenHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
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
	shortURL := sh.baseURL + "/" + shortCode
	sh.storage.Save(shortCode, urlStr)

	// возвращаем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
