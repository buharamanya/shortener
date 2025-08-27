package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/buharamanya/shortener/internal/app/auth"
	"github.com/buharamanya/shortener/internal/app/logger"
	"go.uber.org/zap"
)

type URLDeleter interface {
	DeleteURLs(shortCodes []string, userID string) error
}

// Удаление сохраненных пользователем урлов.
func APIDeleteUserURLsHandler(s URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req []string
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Error("Ошибка чтение запроса", zap.Error(err))
			return
		}
		go func() {
			err := s.DeleteURLs(req, r.Context().Value(auth.UserIDContextKey).(string))
			if err != nil {
				logger.Log.Error("Ошибка удаления url", zap.Error(err))
			}
		}()

		w.WriteHeader(http.StatusAccepted)
	}
}
