package core

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/buharamanya/shortener/internal/app/storage"
)

type ShortenerService struct {
	Storage storage.URLStorage // экспортируемое поле
	BaseURL string             // экспортируемое поле
}

func NewShortenerService(storage storage.URLStorage, baseURL string) *ShortenerService {
	return &ShortenerService{
		Storage: storage,
		BaseURL: baseURL,
	}
}

func (s *ShortenerService) GetHash(urlStr string) string {
	hash := sha256.Sum256([]byte(urlStr))
	shortCode := base64.URLEncoding.EncodeToString(hash[:6])
	return strings.TrimRight(shortCode, "=")
}

func (s *ShortenerService) ShortenURL(ctx context.Context, originalURL, userID string) (string, error) {
	shortCode := s.GetHash(originalURL)

	record := storage.ShortURLRecord{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	err := s.Storage.Save(record)
	if err != nil {
		return "", err
	}

	return s.BaseURL + "/" + shortCode, nil
}

func (s *ShortenerService) ShortenURLBatch(ctx context.Context, items []BatchRequestItem, userID string) ([]BatchResponseItem, error) {
	records := make([]storage.ShortURLRecord, len(items))

	for i, item := range items {
		records[i] = storage.ShortURLRecord{
			OriginalURL:   item.OriginalURL,
			CorrelationID: item.CorrelationID,
			ShortCode:     s.GetHash(item.OriginalURL),
			UserID:        userID,
		}
	}

	err := s.Storage.SaveBatch(records)
	if err != nil {
		return nil, err
	}

	response := make([]BatchResponseItem, len(records))
	for i, record := range records {
		response[i] = BatchResponseItem{
			CorrelationID: record.CorrelationID,
			ShortURL:      s.BaseURL + "/" + record.ShortCode,
		}
	}

	return response, nil
}

func (s *ShortenerService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	return s.Storage.Get(shortCode)
}

func (s *ShortenerService) GetUserURLs(ctx context.Context, userID string) ([]storage.ShortURLRecord, error) {
	return s.Storage.GetURLsByUserID(userID)
}

func (s *ShortenerService) DeleteUserURLs(ctx context.Context, shortCodes []string, userID string) error {
	return s.Storage.DeleteURLs(shortCodes, userID)
}

func (s *ShortenerService) Ping(ctx context.Context) error {
	return nil
}

func (s *ShortenerService) GetStats(ctx context.Context) (urlsCount, usersCount int, err error) {
	return s.Storage.GetStats()
}

type BatchRequestItem struct {
	CorrelationID string
	OriginalURL   string
}

type BatchResponseItem struct {
	CorrelationID string
	ShortURL      string
}
