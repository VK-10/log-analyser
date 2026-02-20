package api

import (
	"encoding/json"
	"net/http"

	"log-classifier/internal/models"
	"log-classifier/internal/worker"
)

func ClassifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var logs []models.LogEntry
	if err := json.NewDecoder(r.Body).Decode(&logs); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(logs) == 0 {
		http.Error(w, "no log entries provided", http.StatusBadRequest)
		return
	}

	results := worker.ProcessLogs(logs, 4)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
