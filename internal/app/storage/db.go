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
		short_code 		VARCHAR(20) 	NOT NULL,
		url  			VARCHAR 		NOT NULL UNIQUE,
		correlation_id  VARCHAR(32)		
	)`

	if _, err = db.Exec(createTableQuery); err != nil {
		logger.Log.Fatal("Ошибка инициализации базы данных", zap.Error(err))
	}

	return &DBStorage{
		DB: db,
	}
}

func (db *DBStorage) Save(shortCode string, originalURL string) error {
	query := `INSERT INTO shorturl (short_code, url) VALUES ($1, $2)`
	_, err := db.Exec(query, shortCode, originalURL)
	return err
}

func (db *DBStorage) SaveBatch(records []ShortURLRecord) error {
	query := `INSERT INTO shorturl (short_code, url, correlation_id) VALUES ($1, $2, $3)`
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, v := range records {
		// все изменения записываются в транзакцию
		_, err := tx.Exec(query, v.ShortCode, v.OriginalURL, v.CorrelationID)
		if err != nil {
			// если ошибка, то откатываем изменения
			tx.Rollback()
			return err
		}
	}
	// завершаем транзакцию
	return tx.Commit()
}

func (db *DBStorage) Get(shortCode string) (string, error) {
	query := `SELECT url FROM shorturl WHERE short_code = $1 LIMIT 1`
	row := db.QueryRow(query, shortCode)
	var url string
	err := row.Scan(&url)
	if err != nil {
		logger.Log.Error("Не нашел записи по запросу", zap.Error(err))
		return "", err
	}
	return url, nil
}
