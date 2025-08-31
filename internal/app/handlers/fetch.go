package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/buharamanya/shortener/internal/app/auth"
	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"go.uber.org/zap"
)

// получатель.
type URLGetterByUserID interface {
	GetURLsByUserID(userID string) ([]storage.ShortURLRecord, error)
}

// получить урлы.
func APIFetchUserURLsHandler(s URLGetterByUserID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		records, err := s.GetURLsByUserID(r.Context().Value(auth.UserIDContextKey).(string))

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error("failed to fetch URLs from storage", zap.Error(err))
			return
		}

		if len(records) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var resp []UserURLsDataResponse

		for _, v := range records {
			resp = append(resp, UserURLsDataResponse{
				ShortURL:    config.AppParams.RedirectBaseURL + "/" + v.ShortCode,
				OriginalURL: v.OriginalURL,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error("error encoding response", zap.Error(err))
			return
		}
	}
}
