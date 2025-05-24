package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type URLSaver interface {
	Save(shortCode string, originalURL string) error
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
	sh.storage.Save(shortCode, urlStr)

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

	// сохраняем в хранилище
	shortURL := sh.baseURL + "/" + shortCode
	sh.storage.Save(shortCode, urlStr)

	// возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	var respDto = ShortenlURLResponce{
		Result: shortURL,
	}
	resp, _ := json.Marshal(respDto)
	w.Write(resp)
}
