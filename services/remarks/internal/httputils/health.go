package httputils

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Method not allowed",
			"allowed_methods": "GET",
		})
		return
	}

	// Устанавливаем заголовки
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)

	// Формируем ответ
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "remarks",
		Version:   "1.0.0",
	}

	json.NewEncoder(w).Encode(response)
}
