package main

import (
	"net/http"

	"github.com/buharamanya/shortener/internal/app/handlers"
	"github.com/buharamanya/shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

const baseURL string = "http://localhost:8080/"

func main() {
	repo := storage.NewInMemoryStorage()
	r := chi.NewRouter()
	r.Post("/", handlers.NewShortenHandler(repo, baseURL).ShortenURL)
	r.Get("/{shortCode}", handlers.NewRedirectHandler(repo).RedirectByShortURL)
	// r передаётся как http.Handler
	http.ListenAndServe(":8080", r)
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", handlers.NewShortenHandler(repo, baseURL).ShortenURL)
	// mux.HandleFunc("/{shortCode}", handlers.NewRedirectHandler(repo).RedirectByShortURL)
	// err := http.ListenAndServe(`:8080`, mux)
	// if err != nil {
	// 	panic(err)
	// }
}
