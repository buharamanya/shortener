package main

import (
	"log"
	"net/http"
	"yandex-go/shortener/internal/app/logger"

	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/handlers"
	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	logger.Initialize("info")

	var appConfig = config.InitConfiguration()

	repo := storage.NewInMemoryStorage()

	r := chi.NewRouter()
	r.Post(
		"/",
		logger.WithLogging(handlers.NewShortenHandler(repo, appConfig.RedirectBaseURL).ShortenURL),
	)
	r.Get(
		"/{shortCode}",
		logger.WithLogging(handlers.NewRedirectHandler(repo).RedirectByShortURL),
	)

	err := http.ListenAndServe(appConfig.ServerBaseURL, r)

	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
