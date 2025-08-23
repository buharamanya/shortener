package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/buharamanya/shortener/internal/app/auth"
	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/handlers"
	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	logger.Initialize("info")

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

	if err := http.ListenAndServe(appConfig.ServerBaseURL, r); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
