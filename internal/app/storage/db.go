package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	*sql.DB
}

func NewDBStorage(dbDSN string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	return &DBStorage{
		DB: db,
	}, nil
}

func (s *DBStorage) Save(shortURL string, originalURL string) error {
	return nil
}

func (s *DBStorage) Get(shortURL string) (string, error) {
	return "", nil
}
