package storage

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/buharamanya/shortener/internal/app/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// type DBStorage struct.
type DBStorage struct {
	*sql.DB
}

// NewDBStorage.
func NewDBStorage(dbDSN string) *DBStorage {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		logger.Log.Fatal("Ошибка инициализации базы данных", zap.Error(err))
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS shorturl (
		short_code 		VARCHAR(20) 	NOT NULL,
		url  			VARCHAR 		NOT NULL UNIQUE,
		correlation_id  VARCHAR(200),
		user_id			VARCHAR(100),
		is_deleted		BOOLEAN 		NOT NULL DEFAULT FALSE		
	)`

	if _, err = db.Exec(createTableQuery); err != nil {
		logger.Log.Fatal("Ошибка инициализации базы данных", zap.Error(err))
	}

	return &DBStorage{
		DB: db,
	}
}

// сохранить.
func (db *DBStorage) Save(record ShortURLRecord) error {
	query := `INSERT INTO shorturl (short_code, url, user_id, correlation_id) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, record.ShortCode, record.OriginalURL, record.UserID, record.CorrelationID)
	return err
}

// много сохранить.
func (db *DBStorage) SaveBatch(records []ShortURLRecord) error {
	query := `INSERT INTO shorturl (short_code, url, correlation_id, user_id) VALUES ($1, $2, $3, $4)`
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, v := range records {
		// все изменения записываются в транзакцию
		_, err := tx.Exec(query, v.ShortCode, v.OriginalURL, v.CorrelationID, v.UserID)
		if err != nil {
			// если ошибка, то откатываем изменения
			tx.Rollback()
			return err
		}
	}
	// завершаем транзакцию
	return tx.Commit()
}

// получить.
func (db *DBStorage) Get(shortCode string) (string, error) {
	query := `SELECT url, is_deleted FROM shorturl WHERE short_code = $1 LIMIT 1`
	row := db.QueryRow(query, shortCode)
	var url string
	var isDeleted bool
	err := row.Scan(&url, &isDeleted)
	if err != nil {
		logger.Log.Error("Не нашел записи по запросу", zap.Error(err))
		return "", err
	}
	if isDeleted {
		return "", ErrDeleted
	}
	return url, nil
}

// получить по пользаку.
func (db *DBStorage) GetURLsByUserID(userID string) ([]ShortURLRecord, error) {
	query := `SELECT short_code, url, correlation_id, user_id
		FROM shorturl
		WHERE user_id = $1`

	urls := []ShortURLRecord{}
	rows, err := db.Query(query, userID)
	if err != nil {
		return []ShortURLRecord{}, fmt.Errorf("failed to execute query: %w", err)
	}

	for rows.Next() {
		var u ShortURLRecord
		err = rows.Scan(&u.ShortCode, &u.OriginalURL, &u.CorrelationID, &u.UserID)
		if err != nil {
			return []ShortURLRecord{}, fmt.Errorf("failed to scan query: %w", err)
		}
		urls = append(urls, u)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return []ShortURLRecord{}, fmt.Errorf("failed to read query: %w", err)
	}

	return urls, nil
}

// удалить.
func (db *DBStorage) DeleteURLs(shortCodes []string, userID string) error {

	if len(shortCodes) == 0 {
		return nil
	}

	query := `UPDATE shorturl SET is_deleted = true WHERE user_id = '` + userID + `' and short_code IN (` + placeholders(len(shortCodes)) + `)`
	args := make([]interface{}, len(shortCodes))
	for i, sc := range shortCodes {
		args[i] = sc
	}
	_, err := db.Exec(query, args...)
	return err
}

func placeholders(n int) string {
	ph := make([]string, n)
	for i := range ph {
		ph[i] = "$" + strconv.Itoa(i+1)
	}
	return strings.Join(ph, ",")
}
