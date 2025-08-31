package handlers

import (
	"net/http"

	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"go.uber.org/zap"
)

// пинг.
func PingHandler(s storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, ok := s.(*storage.DBStorage)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error("failed to connect to DB", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
