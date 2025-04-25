package handlers

import (
	"net/http"
	"strings"

	"github.com/buharamanya/shortener/internal/app/storage"
)

type RedirectHandler struct {
	storage storage.URLStorage
}

func NewRedirectHandler(storage storage.URLStorage) *RedirectHandler {
	return &RedirectHandler{
		storage: storage,
	}
}

func (rh *RedirectHandler) RedirectByShortURL(w http.ResponseWriter, r *http.Request) {
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

	originalURL, err := rh.storage.Get(shortCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
