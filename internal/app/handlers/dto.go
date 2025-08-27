package handlers

// дто ответ.
type UserURLsDataResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// дто на запрос для массового сокращения.
type ShortenlURLBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// дто ответ на запрос для массового сокращения.
type ShortenlURLBatchResponce struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
