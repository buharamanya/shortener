package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestInitialize_Success(t *testing.T) {
	// Сохраняем оригинальный логер для восстановления после теста
	originalLog := Log

	t.Cleanup(func() {
		Log = originalLog
	})

	err := Initialize("debug")
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if Log == nil {
		t.Error("Logger should not be nil after initialization")
	}
}

func TestInitialize_InvalidLevel(t *testing.T) {
	// Сохраняем оригинальный логер для восстановления после теста
	originalLog := Log

	t.Cleanup(func() {
		Log = originalLog
	})

	err := Initialize("invalid_level")
	if err == nil {
		t.Error("Expected error for invalid log level, but got none")
	}

	// Проверяем, что логер остался no-op после ошибки
	if Log != originalLog {
		t.Error("Logger should remain unchanged after initialization error")
	}
}

func TestLoggingResponseWriter_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	responseData := &responseData{}
	lw := loggingResponseWriter{
		ResponseWriter: recorder,
		responseData:   responseData,
	}

	testData := []byte("test response")
	size, err := lw.Write(testData)

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if size != len(testData) {
		t.Errorf("Expected size %d, got %d", len(testData), size)
	}

	if responseData.size != len(testData) {
		t.Errorf("Expected responseData.size %d, got %d", len(testData), responseData.size)
	}

	if !bytes.Equal(recorder.Body.Bytes(), testData) {
		t.Error("Response body doesn't match expected data")
	}
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	responseData := &responseData{}
	lw := loggingResponseWriter{
		ResponseWriter: recorder,
		responseData:   responseData,
	}

	expectedStatus := http.StatusNotFound
	lw.WriteHeader(expectedStatus)

	if responseData.status != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, responseData.status)
	}

	if recorder.Code != expectedStatus {
		t.Errorf("Expected recorder code %d, got %d", expectedStatus, recorder.Code)
	}
}

func BenchmarkWithRequestLogging(b *testing.B) {
	// Инициализируем логер в no-op режиме чтобы избежать вывода в консоль
	originalLog := Log
	Log = zap.NewNop()
	defer func() { Log = originalLog }()

	// Создаем простой handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Оборачиваем middleware
	wrappedHandler := WithRequestLogging(handler)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)
	}
}

func BenchmarkLoggingResponseWriter_Write(b *testing.B) {
	recorder := httptest.NewRecorder()
	responseData := &responseData{}
	lw := loggingResponseWriter{
		ResponseWriter: recorder,
		responseData:   responseData,
	}

	testData := []byte("test data")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		lw.Write(testData)
		responseData.size = 0 // Сбрасываем для следующей итерации
	}
}

func BenchmarkInitialize(b *testing.B) {
	originalLog := Log
	defer func() { Log = originalLog }()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		Initialize("info")
		Log = zap.NewNop() // Сбрасываем чтобы избежать накопления логеров
	}
}

func BenchmarkLoggingResponseWriter_WriteHeader(b *testing.B) {
	recorder := httptest.NewRecorder()
	responseData := &responseData{}
	lw := loggingResponseWriter{
		ResponseWriter: recorder,
		responseData:   responseData,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		lw.WriteHeader(http.StatusOK)
		responseData.status = 0 // Сбрасываем для следующей итерации
	}
}

func BenchmarkConcurrentWithRequestLogging(b *testing.B) {
	originalLog := Log
	Log = zap.NewNop()
	defer func() { Log = originalLog }()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := WithRequestLogging(handler)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)
		}
	})
}
