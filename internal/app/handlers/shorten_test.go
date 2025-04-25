package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buharamanya/shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortenURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		mockSetup      func(*storage.MockURLStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success: Valid URL",
			method: http.MethodPost,
			body:   "https://example.com",
			mockSetup: func(m *storage.MockURLStorage) {
				// Ожидаем вызов Save с любым shortCode и URL
				m.On("Save", mock.Anything, "https://example.com").Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost/", // Без кода, так как он рандомный
		},
		{
			name:           "Fail: Empty URL",
			method:         http.MethodPost,
			body:           "",
			mockSetup:      func(m *storage.MockURLStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "URL cannot be empty",
		},
		{
			name:           "Fail: Wrong HTTP method (GET)",
			method:         http.MethodGet,
			body:           "https://example.com",
			mockSetup:      func(m *storage.MockURLStorage) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "Fail: Wrong HTTP method (PUT)",
			method:         http.MethodPut,
			body:           "https://example.com",
			mockSetup:      func(m *storage.MockURLStorage) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.MockURLStorage)
			handler := NewShortenHandler(mockStorage, "http://localhost/")

			tt.mockSetup(mockStorage)

			req := httptest.NewRequest(tt.method, "/shorten", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ShortenURL(rr, req)

			// Проверяем статус и тело ответа
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				if tt.name == "Success: Valid URL" {
					// Для успешного случая проверяем только префикс URL
					assert.True(t, strings.HasPrefix(rr.Body.String(), tt.expectedBody))
				} else {
					assert.Equal(t, tt.expectedBody, strings.TrimSpace(rr.Body.String()))
				}
			}

			// Проверяем, что все ожидания по моку выполнены
			mockStorage.AssertExpectations(t)
		})
	}
}
