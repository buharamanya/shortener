package storage

import (
	"database/sql"

	"github.com/buharamanya/shortener/internal/app/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type DBStorage struct {
	*sql.DB
}

func NewDBStorage(dbDSN string) *DBStorage {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		logger.Log.Fatal("Ошибка инициализации базы данных", zap.Error(err))
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS shorturl (
		hash VARCHAR(20) PRIMARY KEY,
		url VARCHAR NOT NULL
	);`

	if _, err = db.Exec(createTableQuery); err != nil {
		logger.Log.Fatal("Ошибка инициализации базы данных", zap.Error(err))
	}

	return &DBStorage{
		DB: db,
	}
}

func (db *DBStorage) Save(shortCode string, originalURL string) error {
	query := `
        INSERT INTO shorturl (hash, url)
        VALUES ($1, $2)
        ON CONFLICT (hash)
        DO UPDATE SET
            url = EXCLUDED.url`

	_, err := db.Exec(query, shortCode, originalURL)
	return err
}

func (db *DBStorage) Get(shortCode string) (string, error) {
	query := `SELECT url FROM shorturl WHERE hash = $1`
	row := db.QueryRow(query, shortCode)
	var url string
	err := row.Scan(&url)
	if err != nil {
		logger.Log.Error("Не нашел записи по запросу", zap.Error(err))
		return "", err
	}
	return url, nil
}
