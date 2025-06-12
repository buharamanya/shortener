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

	file, err := os.OpenFile(appConfig.StorageFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Ошибка запуска файлового хранилища:", err)
	}
	defer file.Close()

	repo := storage.NewInMemoryStorage(file)

	var dbrepo *storage.DBStorage
	if dbrepo, err = storage.NewDBStorage(appConfig.DataBaseDSN); err != nil {
		log.Fatal("Ошибка инициализвции базы данных")
	}

	shortenHandler := handlers.NewShortenHandler(repo, appConfig.RedirectBaseURL)

	r := chi.NewRouter()
	r.Use(handlers.WithGzipMiddleware, logger.WithRequestLogging)
	r.Post("/", shortenHandler.ShortenURL)
	r.Get("/{shortCode}", handlers.NewRedirectHandler(repo).RedirectByShortURL)
	r.Post("/api/shorten", shortenHandler.JSONShortenURL)
	r.Get("/ping", handlers.PingHandler(dbrepo))

	if err := http.ListenAndServe(appConfig.ServerBaseURL, r); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
