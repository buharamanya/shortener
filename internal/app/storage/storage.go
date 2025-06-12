package storage

import (
	"errors"
)

var ErrNotFound = errors.New("URL not found")

type ShortURLRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
