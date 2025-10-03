package handlers

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"go.uber.org/zap"
)

// StatsResponse - структура ответа для статистики
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// APIStatsHandler - обработчик для получения статистики
func APIStatsHandler(repo storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем доверенную подсеть
		if !isTrustedIP(r, config.AppParams.TrustedSubnet) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Получаем статистику из хранилища
		stats, err := getStats(repo)
		if err != nil {
			logger.Log.Error("Ошибка получения статистики", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(stats)
	}
}

// isTrustedIP проверяет, находится ли IP в доверенной подсети
func isTrustedIP(r *http.Request, trustedSubnet string) bool {
	if trustedSubnet == "" {
		return false
	}

	// Получаем IP из заголовка X-Real-IP
	ipStr := r.Header.Get("X-Real-IP")
	if ipStr == "" {
		return false
	}

	// Парсим IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Парсим CIDR
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		logger.Log.Error("Ошибка парсинга trusted_subnet", zap.String("subnet", trustedSubnet), zap.Error(err))
		return false
	}

	// Проверяем вхождение IP в подсеть
	return subnet.Contains(ip)
}

// getStats получает статистику из хранилища через единый интерфейс
func getStats(repo storage.URLStorage) (*StatsResponse, error) {
	urlsCount, usersCount, err := repo.GetStats()
	if err != nil {
		return nil, err
	}

	return &StatsResponse{
		URLs:  urlsCount,
		Users: usersCount,
	}, nil
}
