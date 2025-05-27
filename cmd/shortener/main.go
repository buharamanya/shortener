package main

import (
	"log"
	"net/http"
	"os"

	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/handlers"
	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	logger.Initialize("info")

	var appConfig = config.InitConfiguration()

	file, fileerr := os.OpenFile(appConfig.StorageFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if fileerr != nil {
		log.Fatal("Ошибка запуска файлового хранилища:", fileerr)
	}
	defer file.Close()

	repo := storage.NewInMemoryStorage(file)

	shortenHandler := handlers.NewShortenHandler(repo, appConfig.RedirectBaseURL)

	r := chi.NewRouter()
	r.Use(handlers.WithGzipMiddleware, logger.WithRequestLogging)
	r.Post("/", shortenHandler.ShortenURL)
	r.Get("/{shortCode}", handlers.NewRedirectHandler(repo).RedirectByShortURL)
	r.Post("/api/shorten", shortenHandler.JSONShortenURL)

	err := http.ListenAndServe(appConfig.ServerBaseURL, r)

	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
