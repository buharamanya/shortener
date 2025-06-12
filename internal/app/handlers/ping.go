package handlers

import (
	"net/http"

	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"go.uber.org/zap"
)

func PingHandler(db *storage.DBStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error("failed to connect to DB", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
