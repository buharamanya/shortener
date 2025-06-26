package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type contextKey string

const UserIDContextKey contextKey = "userID"

func WithAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCookie, cookieErr := r.Cookie("AUTH_TOKEN")

			if cookieErr != nil {
				if errors.Is(cookieErr, http.ErrNoCookie) {
					var err error
					authCookie, err = setAuthCookie(w)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						logger.Log.Error("failed to build auth token", zap.Error(err))
						return
					}
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Log.Error("failed to fetch cookie", zap.Error(cookieErr))
					return
				}
			}

			userID := getUserID(authCookie.Value)

			if userID == "" {
				authCookie, err := setAuthCookie(w)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Log.Error("failed to build auth token", zap.Error(err))
					return
				}
				userID = getUserID(authCookie.Value)
			}

			newContext := context.WithValue(r.Context(), UserIDContextKey, userID)
			newRequest := r.WithContext(newContext)
			next.ServeHTTP(w, newRequest)
		})
	}
}

func WithCheckAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCookie, cookieErr := r.Cookie("AUTH_TOKEN")

			if cookieErr != nil {
				w.WriteHeader(http.StatusUnauthorized)
				logger.Log.Error("failed to fetch auth token", zap.Error(cookieErr))
				return
			}

			userID := getUserID(authCookie.Value)

			if userID == "" {
				w.WriteHeader(http.StatusUnauthorized)
				logger.Log.Error("failed to parse auth token")
				return
			}

			newContext := context.WithValue(r.Context(), UserIDContextKey, userID)
			newRequest := r.WithContext(newContext)
			next.ServeHTTP(w, newRequest)
		})
	}
}

func setAuthCookie(w http.ResponseWriter) (*http.Cookie, error) {
	authToken, err := buildJWTString()
	if err != nil {
		return nil, fmt.Errorf("failed to build auth token: %w", err)
	}

	cookie := &http.Cookie{
		Name:     "AUTH_TOKEN",
		Value:    authToken,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)

	return cookie, nil
}

func buildJWTString() (string, error) {
	userID := uuid.New().String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(config.AppParams.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to signed token: %w", err)
	}

	return tokenString, nil
}

func getUserID(tokenString string) string {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.AppParams.SecretKey), nil
	})

	if err != nil {
		return ""
	}

	return claims.UserID
}
