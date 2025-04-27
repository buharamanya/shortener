package main

import (
	"net/http"

	"github.com/buharamanya/shortener/internal/app/config"
	"github.com/buharamanya/shortener/internal/app/handlers"
	"github.com/buharamanya/shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

// const baseURL string = "http://localhost:8080/"

func main() {
	config.InitConfiguration()

	repo := storage.NewInMemoryStorage()

	r := chi.NewRouter()
	r.Post("/", handlers.NewShortenHandler(repo, config.Config.RedirectBaseURL).ShortenURL)
	r.Get("/{shortCode}", handlers.NewRedirectHandler(repo).RedirectByShortURL)

	err := http.ListenAndServe(config.Config.ServerBaseURL, r)

	if err != nil {
		panic(err)
	}
}
