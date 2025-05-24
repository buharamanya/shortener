package handlers

import (
	"bytes"
	"errors"
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
			name:   "Success:_Valid_URL",
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
			name:           "Fail:_Empty_URL",
			method:         http.MethodPost,
			body:           "",
			mockSetup:      func(m *storage.MockURLStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "URL cannot be empty",
		},
		{
			name:           "Fail:_Wrong_HTTP_method_(GET)",
			method:         http.MethodGet,
			body:           "https://example.com",
			mockSetup:      func(m *storage.MockURLStorage) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "Fail:_Wrong_HTTP_method_(PUT)",
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
			assert.Equal(t, tt.expectedStatus, rr.Code, "Ошибка: некорректный статуса ответа")
			if tt.expectedBody != "" {
				if tt.name == "Success:_Valid_URL" {
					// Для успешного случая проверяем только префикс URL
					assert.True(t, strings.HasPrefix(rr.Body.String(), tt.expectedBody), "Ошибка: проверь тело ответа")
				} else {
					assert.Equal(t, tt.expectedBody, strings.TrimSpace(rr.Body.String()), "Ошибка: проверь тело ответа")
				}
			}

			// Проверяем, что все ожидания по моку выполнены
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestJSONShortenURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		setupMock      func(*storage.MockURLStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Wrong_method_GET",
			method:         http.MethodGet,
			contentType:    "application/json",
			body:           `{"url":"https://example.com"}`,
			setupMock:      func(ms *storage.MockURLStorage) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "Wrong_content_type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           `{"url":"https://example.com"}`,
			setupMock:      func(ms *storage.MockURLStorage) {},
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   "",
		},
		{
			name:           "Empty_URL",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":""}`,
			setupMock:      func(ms *storage.MockURLStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "URL cannot be empty",
		},
		{
			name:           "Invalid_JSON",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":}`,
			setupMock:      func(ms *storage.MockURLStorage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:        "Valid_request",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"url":"https://example.com"}`,
			setupMock: func(ms *storage.MockURLStorage) {
				ms.On("Save", mock.AnythingOfType("string"), "https://example.com").Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"http://localhost/`,
		},
		{
			name:        "Storage_error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"url":"https://example.com"}`,
			setupMock: func(ms *storage.MockURLStorage) {
				ms.On("Save", mock.AnythingOfType("string"), "https://example.com").Return(errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый мок для каждого теста
			mockStorage := new(storage.MockURLStorage)
			sh := &ShortenHandler{
				baseURL: "http://localhost",
				storage: mockStorage,
			}

			// Настраиваем мок
			tt.setupMock(mockStorage)

			req := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			sh.JSONShortenURL(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body := new(bytes.Buffer)
				body.ReadFrom(resp.Body)
				bodyStr := body.String()

				if tt.expectedStatus == http.StatusCreated {
					if !strings.HasPrefix(bodyStr, tt.expectedBody) {
						t.Errorf("Expected body to start with %q, got %q", tt.expectedBody, bodyStr)
					}
				} else if !strings.Contains(bodyStr, tt.expectedBody) {
					t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, bodyStr)
				}
			}

			// Проверяем ожидания мока
			mockStorage.AssertExpectations(t)
		})
	}
}
