package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildJWTString(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	token, err := buildJWTString()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGetUserID_ValidToken(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	// Создаем валидный токен
	tokenString, err := buildJWTString()
	require.NoError(t, err)

	userID := getUserID(tokenString)
	assert.NotEmpty(t, userID)
}

func TestGetUserID_InvalidToken(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	// Невалидный токен
	userID := getUserID("invalid.token.here")
	assert.Empty(t, userID)

	// Токен с неправильной подписью
	config.AppParams.SecretKey = "different-secret"
	tokenString, err := buildJWTString()
	require.NoError(t, err)

	config.AppParams.SecretKey = "test-secret-key" // Возвращаем оригинальный секрет
	userID = getUserID(tokenString)
	assert.Empty(t, userID)
}

func TestSetAuthCookie(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	w := httptest.NewRecorder()
	cookie, err := setAuthCookie(w)

	require.NoError(t, err)
	assert.NotNil(t, cookie)
	assert.Equal(t, "AUTH_TOKEN", cookie.Name)
	assert.NotEmpty(t, cookie.Value)
	assert.True(t, cookie.HttpOnly)

	// Проверяем, что cookie установлена в response writer
	response := w.Result()
	defer response.Body.Close()

	cookies := response.Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "AUTH_TOKEN", cookies[0].Name)
	assert.Equal(t, cookie.Value, cookies[0].Value)
}

func TestWithAuthMiddleware_NoCookie(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	middleware := WithAuthMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDContextKey)
		assert.NotNil(t, userID)
		assert.NotEmpty(t, userID.(string))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что cookie была установлена
	response := w.Result()
	defer response.Body.Close()

	cookies := response.Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "AUTH_TOKEN", cookies[0].Name)
}

func TestWithAuthMiddleware_ValidCookie(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	// Сначала создаем валидный токен
	tokenString, err := buildJWTString()
	require.NoError(t, err)

	middleware := WithAuthMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDContextKey)
		assert.NotNil(t, userID)
		assert.NotEmpty(t, userID.(string))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "AUTH_TOKEN", Value: tokenString})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Закрываем body
	response := w.Result()
	defer response.Body.Close()
	io.Copy(io.Discard, response.Body) // Читаем body чтобы избежать утечек
}

func TestWithAuthMiddleware_InvalidCookie(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	middleware := WithAuthMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDContextKey)
		assert.NotNil(t, userID)
		assert.NotEmpty(t, userID.(string))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "AUTH_TOKEN", Value: "invalid-token"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что была установлена новая cookie
	response := w.Result()
	defer response.Body.Close()

	cookies := response.Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "AUTH_TOKEN", cookies[0].Name)
	assert.NotEqual(t, "invalid-token", cookies[0].Value)
}

func TestWithCheckAuthMiddleware_NoCookie(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	middleware := WithCheckAuthMiddleware()
	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID := r.Context().Value(UserIDContextKey)
		assert.NotNil(t, userID)
		assert.NotEmpty(t, userID.(string))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Хендлер ДОЛЖЕН быть вызван, т.к. middleware устанавливает новую cookie при отсутствии
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что cookie была установлена
	response := w.Result()
	defer response.Body.Close()

	cookies := response.Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "AUTH_TOKEN", cookies[0].Name)
}

func TestWithCheckAuthMiddleware_ValidCookie(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	tokenString, err := buildJWTString()
	require.NoError(t, err)

	middleware := WithCheckAuthMiddleware()
	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID := r.Context().Value(UserIDContextKey)
		assert.NotNil(t, userID)
		assert.NotEmpty(t, userID.(string))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "AUTH_TOKEN", Value: tokenString})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)

	// Закрываем body
	response := w.Result()
	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)
}

func TestWithCheckAuthMiddleware_InvalidCookie(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	middleware := WithCheckAuthMiddleware()
	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		t.Error("Handler should not be called when auth fails with invalid token")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "AUTH_TOKEN", Value: "invalid-token"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Хендлер НЕ должен быть вызван при невалидном токене
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Закрываем body
	response := w.Result()
	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)
}

func TestContextKeyType(t *testing.T) {
	// Проверяем, что контекстный ключ имеет правильный тип
	var key contextKey = "test"
	assert.IsType(t, UserIDContextKey, key)
}

func TestMiddlewareChaining(t *testing.T) {
	config.AppParams.SecretKey = "test-secret-key"

	// Тестируем цепочку middleware
	authMiddleware := WithAuthMiddleware()
	checkAuthMiddleware := WithCheckAuthMiddleware()

	handler := authMiddleware(checkAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDContextKey)
		assert.NotNil(t, userID)
		assert.NotEmpty(t, userID.(string))
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Закрываем body
	response := w.Result()
	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)
}
