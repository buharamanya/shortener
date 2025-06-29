package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/google/uuid"
)

// InMemoryStorage - реализация хранилища в памяти
type InMemoryStorage struct {
	file os.File
	urls map[string]ShortURLRecord
}

func NewInMemoryStorage(file *os.File) *InMemoryStorage {

	urls := make(map[string]ShortURLRecord)

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

		urls[record.ShortCode] = record
	}

	return &InMemoryStorage{
		*file,
		urls,
	}
}

func (s *InMemoryStorage) Save(record ShortURLRecord) error {
	record.CorrelationID = uuid.New().String()
	s.urls[record.ShortCode] = record
	encoder := json.NewEncoder(&s.file)
	encoder.Encode(record)
	return nil
}

func (s *InMemoryStorage) SaveBatch(records []ShortURLRecord) error {
	for _, v := range records {
		s.urls[v.ShortCode] = v
		encoder := json.NewEncoder(&s.file)
		encoder.Encode(v)
	}
	return nil
}

func (s *InMemoryStorage) Get(shortCode string) (string, error) {
	url, exists := s.urls[shortCode]
	if !exists {
		return "", ErrNotFound
	}
	if url.DeletedFlag == true {
		return "", ErrDeleted
	}
	return url.OriginalURL, nil
}

func (s *InMemoryStorage) GetURLsByUserID(userID string) ([]ShortURLRecord, error) {
	var userURLs []ShortURLRecord
	for _, v := range s.urls {
		if v.UserID == userID {
			userURLs = append(userURLs, v)
		}
	}
	return userURLs, nil
}

func (s *InMemoryStorage) DeleteURLs(shortCodes []string, userID string) error {

	for _, v := range shortCodes {
		record, ok := s.urls[v]
		if ok && record.UserID == userID {
			record.DeletedFlag = true
			s.urls[v] = record
			encoder := json.NewEncoder(&s.file)
			encoder.Encode(record)
		}
	}

	return nil
}
