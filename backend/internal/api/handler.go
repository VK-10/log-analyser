package api

import (
	"encoding/json"
	"net/http"

	"log-classifier/backend/internal/models"
	"log-classifier/backend/internal/worker"
)

func ClassifyHandler(w http.ResponseWriter, r *http.Request) {
	var logs []models.LogEntry
	json.NewDecoder(r.Body).Decode(&logs)

	results := worker.ProcessLogs(logs, 4)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
