package storage

import (
	"errors"
)

var ErrNotFound = errors.New("URL not found")
var ErrDeleted = errors.New("URL was deleted")

type ShortURLRecord struct {
	ShortCode     string `json:"short_code"`
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
	UserID        string `json:"user_id"`
	DeletedFlag   bool   `json:"is_deleted"`
}

type URLStorage interface {
	Get(shortCode string) (string, error)
	Save(record ShortURLRecord) error
	SaveBatch(records []ShortURLRecord) error
	GetURLsByUserID(userID string) ([]ShortURLRecord, error)
	DeleteURLs(shortCodes []string, userID string) error
}
