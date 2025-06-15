package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/buharamanya/shortener/internal/app/logger"
	"github.com/buharamanya/shortener/internal/app/storage"
	"go.uber.org/zap"
)

type URLSaver interface {
	Save(shortCode string, originalURL string) error
	SaveBatch(records []storage.ShortURLRecord) error
}

type ShortenHandler struct {
	storage URLSaver
	baseURL string
}

func NewShortenHandler(storage URLSaver, baseURL string) *ShortenHandler {
	return &ShortenHandler{
		storage: storage,
		baseURL: baseURL,
	}
}

func getHash(urlStr string) string {
	hash := sha256.Sum256([]byte(urlStr))
	shortCode := base64.URLEncoding.EncodeToString(hash[:6])
	return strings.TrimRight(shortCode, "=")
}

func (sh *ShortenHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// проверяем метод запроса
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// читаем тело запроса
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	urlStr := strings.TrimSpace(string(originalURL))

	// проверяем что URL не пустой
	if urlStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL cannot be empty"))
		return
	}

	// генерируем короткий код
	shortCode := getHash(urlStr)

	// сохраняем в хранилище
	shortURL := sh.baseURL + "/" + shortCode
	err = sh.storage.Save(shortCode, urlStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// возвращаем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

type ShortenlURLRequest struct {
	URL string `json:"url"`
}

type ShortenlURLResponce struct {
	Result string `json:"result"`
}

func isJSONContentType(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "application/json")
}

func (sh *ShortenHandler) JSONShortenURL(w http.ResponseWriter, r *http.Request) {
	// проверяем метод запроса
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// проверяем тип содержимого
	if !isJSONContentType(r) {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var reqDto ShortenlURLRequest
	var buf bytes.Buffer
	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// десериализуем JSON
	if err = json.Unmarshal(buf.Bytes(), &reqDto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// читаем тело запроса

	urlStr := strings.TrimSpace(string(reqDto.URL))

	// проверяем что URL не пустой
	if urlStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL cannot be empty"))
		return
	}

	// генерируем короткий код
	shortCode := getHash(urlStr)

	// создаем и сохраняем в хранилище короткую ссылку
	shortURL := sh.baseURL + "/" + shortCode
	err = sh.storage.Save(shortCode, urlStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	var respDto = ShortenlURLResponce{
		Result: shortURL,
	}
	resp, _ := json.Marshal(respDto)
	w.Write(resp)
}

type ShortenlURLBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenlURLBatchResponce struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (sh *ShortenHandler) JSONShortenBatchURL(w http.ResponseWriter, r *http.Request) {

	var req []ShortenlURLBatchRequest

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Log.Error("failed to read request body", zap.Error(err))
		return
	}

	var records []storage.ShortURLRecord

	for _, v := range req {
		records = append(
			records,
			storage.ShortURLRecord{
				OriginalURL:   v.OriginalURL,
				CorrelationID: v.CorrelationID,
				ShortCode:     getHash(v.OriginalURL),
			},
		)
	}

	err := sh.storage.SaveBatch(records)
	if err != nil {
		logger.Log.Error("Ошибка сохранения группы записей", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var resp []ShortenlURLBatchResponce

	for _, v := range records {
		resp = append(
			resp,
			ShortenlURLBatchResponce{
				CorrelationID: v.CorrelationID,
				ShortURL:      sh.baseURL + "/" + v.ShortCode,
			},
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("error encoding response", zap.Error(err))
		return
	}
}
