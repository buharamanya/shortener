package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
)

var ErrNotFound = errors.New("URL not found")

// InMemoryStorage - реализация хранилища в памяти
type InMemoryStorage struct {
	file os.File
	urls map[string]string
}

type ShortURLRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewInMemoryStorage(file *os.File) *InMemoryStorage {
	// file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	return nil, err
	// }

	urls := make(map[string]string)

	if _, err := file.Seek(0, 0); err != nil {
		log.Fatal("не удалось перейти в начало файла")
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record ShortURLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			log.Printf("Ошибка декодирования строки '%s': %v", line, err)
			continue
		}

		urls[record.ShortURL] = record.OriginalURL
	}

	return &InMemoryStorage{
		*file,
		urls,
	}
}

func (s *InMemoryStorage) Save(shortCode string, originalURL string) error {
	s.urls[shortCode] = originalURL
	record := ShortURLRecord{
		shortCode,
		originalURL,
	}
	encoder := json.NewEncoder(&s.file)
	encoder.Encode(record)
	return nil
}

func (s *InMemoryStorage) Get(shortCode string) (string, error) {
	url, exists := s.urls[shortCode]
	if !exists {
		return "", ErrNotFound
	}
	return url, nil
}
