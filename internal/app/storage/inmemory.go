package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/buharamanya/shortener/internal/app/logger"
)

// InMemoryStorage - реализация хранилища в памяти
type InMemoryStorage struct {
	file os.File
	urls map[string]string
}

func NewInMemoryStorage(file *os.File) *InMemoryStorage {

	urls := make(map[string]string)

	if _, err := file.Seek(0, 0); err != nil {
		logger.Log.Fatal("не удалось перейти в начало файла")
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record ShortURLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			logger.Log.Info(fmt.Sprintf("Ошибка декодирования строки '%s': %v", line, err))
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
