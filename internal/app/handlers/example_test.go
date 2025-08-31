package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/buharamanya/shortener/internal/app/auth"
	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/storage"
)

// ExampleAPIFetchUserURLsHandler демонстрирует использование APIFetchUserURLsHandler
func ExampleAPIFetchUserURLsHandler() {
	// Настройка базового URL для редиректов
	config.AppParams.RedirectBaseURL = "http://localhost:8080"

	// Создаем мок хранилища с тестовыми данными
	mockStorage := &ExampleStorage{
		Records: []storage.ShortURLRecord{
			{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com/page1",
				UserID:      "user-123",
			},
			{
				ShortCode:   "def456",
				OriginalURL: "https://example.com/page2",
				UserID:      "user-123",
			},
			{
				ShortCode:   "ghi789",
				OriginalURL: "https://example.com/page3",
				UserID:      "user-123",
			},
		},
	}

	// Создаем обработчик
	handler := APIFetchUserURLsHandler(mockStorage)

	// Создаем тестовый HTTP запрос
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	// Добавляем userID в контекст (как это делает middleware аутентификации)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, "user-123")
	req = req.WithContext(ctx)

	// Создаем ResponseWriter для захвата ответа
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	fmt.Printf("Status Code: %d\n", rr.Code)

	// Декодируем и выводим ответ
	if rr.Code == http.StatusOK {
		var response []UserURLsDataResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err == nil {
			fmt.Println("User URLs:")
			for _, url := range response {
				fmt.Printf("  Short: %s -> Original: %s\n", url.ShortURL, url.OriginalURL)
			}
		}
	}

	// Output:
	// Status Code: 200
	// User URLs:
	//   Short: http://localhost:8080/abc123 -> Original: https://example.com/page1
	//   Short: http://localhost:8080/def456 -> Original: https://example.com/page2
	//   Short: http://localhost:8080/ghi789 -> Original: https://example.com/page3
}

// ExampleStorage реализует URLGetterByUserID для примеров
type ExampleStorage struct {
	Records []storage.ShortURLRecord
}

func (es *ExampleStorage) GetURLsByUserID(userID string) ([]storage.ShortURLRecord, error) {
	// Фильтруем записи по userID
	var userRecords []storage.ShortURLRecord
	for _, record := range es.Records {
		if record.UserID == userID {
			userRecords = append(userRecords, record)
		}
	}
	return userRecords, nil
}
