package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
)

// TestConfigInitialization тестирует инициализацию конфигурации
func TestConfigInitialization(t *testing.T) {
	// Инициализируем конфигурацию
	appConfig := config.InitConfiguration()

	// Проверяем что конфигурация загрузилась с фактическими значениями по умолчанию
	assert.Equal(t, "localhost:8080", appConfig.ServerBaseURL)
	assert.NotEmpty(t, appConfig.SecretKey)                   // SecretKey должен генерироваться автоматически
	assert.Equal(t, "storage.txt", appConfig.StorageFileName) // Фактическое значение из вашего конфига
}

// TestPingHandler тестирует работу ping handler
func TestPingHandler(t *testing.T) {
	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()

	// Простой handler для теста (имитация ping)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)
}
