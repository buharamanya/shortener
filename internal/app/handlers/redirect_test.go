package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buharamanya/shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestRedirectByShortURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		mockSetup      func(*storage.MockURLStorage)
		expectedStatus int
		expectedHeader string
	}{
		{
			name:   "Success:Valid_short_URL",
			method: http.MethodGet,
			path:   "/abc123",
			mockSetup: func(m *storage.MockURLStorage) {
				m.On("Get", "abc123").Return("https://example.com", nil)
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "https://example.com",
		},
		{
			name:           "Fail:_Empty_short_URL",
			method:         http.MethodGet,
			path:           "/",
			mockSetup:      func(m *storage.MockURLStorage) {}, // Мок не должен вызываться
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "",
		},
		{
			name:   "Fail:_Short_URL_not_found",
			method: http.MethodGet,
			path:   "/invalid",
			mockSetup: func(m *storage.MockURLStorage) {
				m.On("Get", "invalid").Return("", storage.ErrNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "",
		},
		{
			name:           "Fail:_Wrong_HTTP_method_(POST)",
			method:         http.MethodPost,
			path:           "/abc123",
			mockSetup:      func(m *storage.MockURLStorage) {}, // Мок не должен вызываться
			expectedStatus: http.StatusMethodNotAllowed,
			expectedHeader: "",
		},
		{
			name:           "Fail:_Wrong_HTTP_method_(PUT)",
			method:         http.MethodPut,
			path:           "/abc123",
			mockSetup:      func(m *storage.MockURLStorage) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.MockURLStorage)
			handler := NewRedirectHandler(mockStorage)

			// Настраиваем мок, если требуется
			tt.mockSetup(mockStorage)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			handler.RedirectByShortURL(rr, req)

			// Проверяем статус и заголовок
			assert.Equal(t, tt.expectedStatus, rr.Code, "Ошибка: некорректный статуса ответа")
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, rr.Header().Get("Location"), "Ошибка: хедер Location отсутствует")
			}

			// Проверяем, что все ожидания по моку выполнены
			mockStorage.AssertExpectations(t)
		})
	}
}
