package storage

import (
	"github.com/stretchr/testify/mock"
)

// MockURLStorage реализует URLStorage для тестов.
type MockURLStorage struct {
	mock.Mock
}

// получить.
func (m *MockURLStorage) Get(shortCode string) (string, error) {
	args := m.Called(shortCode)
	return args.String(0), args.Error(1)
}

// прихранить.
func (m *MockURLStorage) Save(record ShortURLRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

// много прихранить.
func (m *MockURLStorage) SaveBatch(records []ShortURLRecord) error {
	args := m.Called(records)
	return args.Error(0)
}
