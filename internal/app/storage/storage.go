package storage

import (
	"errors"
)

var ErrNotFound = errors.New("URL not found")

type ShortURLRecord struct {
	ShortCode     string `json:"short_code"`
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type URLStorage interface {
	Get(shortCode string) (string, error)
	Save(shortCode string, originalURL string) error
	SaveBatch([]ShortURLRecord) error
}
