package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"golang.org/x/crypto/acme/autocert"

	"github.com/buharamanya/shortener/internal/app/auth"
	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/handlers"
	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {

	logger.Initialize("info")

	logger.Log.Info("Build info: ", zap.String("version", buildVersion))
	logger.Log.Info("Build info: ", zap.String("date", buildDate))
	logger.Log.Info("Build info: ", zap.String("commit", buildCommit))

	var appConfig = config.InitConfiguration()

	var repo storage.URLStorage

	if appConfig.DataBaseDSN == "" {
		file, err := os.OpenFile(appConfig.StorageFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Ошибка запуска файлового хранилища:", err)
		}
		defer file.Close()
		repo = storage.NewInMemoryStorage(file)
	} else {
		repo = storage.NewDBStorage(appConfig.DataBaseDSN)
	}

	shortenHandler := handlers.NewShortenHandler(repo, appConfig.RedirectBaseURL)

	r := chi.NewRouter()

	r.Use(logger.WithRequestLogging)

	r.Get("/ping", handlers.PingHandler(repo))

	r.Group(func(r chi.Router) {
		r.Mount("/debug/pprof", http.DefaultServeMux)
	})

	r.Group(func(r chi.Router) {
		r.Use(handlers.WithGzipMiddleware, auth.WithAuthMiddleware())
		r.Post("/", shortenHandler.ShortenURL)
		r.Get("/{shortCode}", handlers.NewRedirectHandler(repo).RedirectByShortURL)
		r.Post("/api/shorten", shortenHandler.JSONShortenURL)
		r.Post("/api/shorten/batch", shortenHandler.JSONShortenBatchURL)
		r.Delete("/api/user/urls", handlers.APIDeleteUserURLsHandler(repo))
	})

	r.Group(func(r chi.Router) {
		r.Use(handlers.WithGzipMiddleware, auth.WithCheckAuthMiddleware())
		r.Get("/api/user/urls", handlers.APIFetchUserURLsHandler(repo))
	})

	// Запуск сервера с поддержкой HTTPS через autocert
	if appConfig.EnableHTTPS {
		logger.Log.Info("Запуск сервера с HTTPS (autocert)", zap.String("address", appConfig.ServerBaseURL))

		// Создаем менеджер TLS-сертификатов
		manager := &autocert.Manager{
			// Директория для хранения сертификатов
			Cache: autocert.DirCache("certs"),
			// Принимаем Terms of Service Let's Encrypt
			Prompt: autocert.AcceptTOS,
			// Для тестирования можно указать домен, в продакшене нужно указать реальные домены
			// HostPolicy: autocert.HostWhitelist("your-domain.com", "www.your-domain.com"),
			HostPolicy: autocert.HostWhitelist(),
		}

		// HTTP сервер для ACME challenge (проверка домена) на порту 80
		go func() {
			httpServer := &http.Server{
				Addr:    ":80",
				Handler: manager.HTTPHandler(nil),
			}
			logger.Log.Info("Запуск ACME challenge сервера на порту 80")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Log.Error("Ошибка ACME сервера", zap.Error(err))
			}
		}()

		// HTTPS сервер с автоматическими сертификатами
		server := &http.Server{
			Addr:      appConfig.ServerBaseURL,
			Handler:   r,
			TLSConfig: manager.TLSConfig(),
		}

		logger.Log.Info("Запуск HTTPS сервера")
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatal("Ошибка запуска HTTPS сервера:", err)
		}

	} else {
		// Обычный HTTP сервер
		logger.Log.Info("Запуск сервера с HTTP", zap.String("address", appConfig.ServerBaseURL))

		server := &http.Server{
			Addr:    appConfig.ServerBaseURL,
			Handler: r,
		}

		if err := server.ListenAndServe(); err != nil {
			log.Fatal("Ошибка запуска сервера:", err)
		}
	}
}
