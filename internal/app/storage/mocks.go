package storage

import "github.com/stretchr/testify/mock"

// MockURLStorage реализует URLStorage для тестов.
type MockURLStorage struct {
	mock.Mock
}

func (m *MockURLStorage) Get(shortCode string) (string, error) {
	args := m.Called(shortCode)
	return args.String(0), args.Error(1)
}

func (m *MockURLStorage) Save(shortCode, originalURL string) error {
	args := m.Called(shortCode, originalURL)
	return args.Error(0)
}
