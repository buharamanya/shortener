package storage

import "errors"

var ErrNotFound = errors.New("URL not found")

// InMemoryStorage - реализация хранилища в памяти
type InMemoryStorage struct {
	urls map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		urls: make(map[string]string),
	}
}

func (s *InMemoryStorage) Save(shortCode string, originalURL string) error {
	s.urls[shortCode] = originalURL
	return nil
}

func (s *InMemoryStorage) Get(shortCode string) (string, error) {
	url, exists := s.urls[shortCode]
	if !exists {
		return "", ErrNotFound
	}
	return url, nil
}
