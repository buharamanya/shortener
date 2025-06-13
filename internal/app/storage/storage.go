package storage

import (
	"errors"
)

var ErrNotFound = errors.New("URL not found")

type ShortURLRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLStorage interface {
	Get(shortCode string) (string, error)
	Save(shortCode string, originalURL string) error
}
