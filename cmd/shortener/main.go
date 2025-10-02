package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

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
			logger.Log.Fatal("Ошибка запуска файлового хранилища:", zap.Error(err))
		}
		repo = storage.NewInMemoryStorage(file)
	} else {
		var err error
		repo, err = storage.NewDBStorage(appConfig.DataBaseDSN)
		if err != nil {
			logger.Log.Fatal("Ошибка подключения к базе данных:", zap.Error(err))
		}
	}
	defer repo.Close()

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

	// Создаем канал для сигналов ОС
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем сервер в отдельной горутине
	server := &http.Server{
		Addr:    appConfig.ServerBaseURL,
		Handler: r,
	}

	var serverErr error
	serverStopped := make(chan struct{})

	go func() {
		if appConfig.EnableHTTPS {
			logger.Log.Info("Запуск сервера с HTTPS (autocert)", zap.String("address", appConfig.ServerBaseURL))

			// Создаем менеджер TLS-сертификатов
			manager := &autocert.Manager{
				Cache:      autocert.DirCache("certs"),
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(),
			}

			// HTTP сервер для ACME challenge на порту 80
			acmeServer := &http.Server{
				Addr:    ":80",
				Handler: manager.HTTPHandler(nil),
			}

			go func() {
				logger.Log.Info("Запуск ACME challenge сервера на порту 80")
				if err := acmeServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Log.Error("Ошибка ACME сервера", zap.Error(err))
				}
			}()

			// HTTPS сервер с автоматическими сертификатами
			server.TLSConfig = manager.TLSConfig()
			serverErr = server.ListenAndServeTLS("", "")
		} else {
			// Обычный HTTP сервер
			logger.Log.Info("Запуск сервера с HTTP", zap.String("address", appConfig.ServerBaseURL))
			serverErr = server.ListenAndServe()
		}

		if serverErr != nil && serverErr != http.ErrServerClosed {
			logger.Log.Error("Ошибка сервера", zap.Error(serverErr))
		}
		close(serverStopped)
	}()

	// Ожидаем сигнал или ошибку сервера
	select {
	case sig := <-signalChan:
		logger.Log.Info("Получен сигнал завершения", zap.String("signal", sig.String()))
	case <-serverStopped:
		logger.Log.Info("Сервер остановился самостоятельно")
	}

	// Инициируем graceful shutdown
	logger.Log.Info("Начинаем graceful shutdown...")

	// Создаем контекст с таймаутом для shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Останавливаем сервер
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Ошибка при graceful shutdown", zap.Error(err))
	} else {
		logger.Log.Info("Сервер успешно остановлен")
	}

	// Закрываем хранилище
	if err := repo.Close(); err != nil {
		logger.Log.Error("Ошибка при закрытии хранилища", zap.Error(err))
	} else {
		logger.Log.Info("Хранилище успешно закрыто")
	}

	logger.Log.Info("Приложение завершено")
}
