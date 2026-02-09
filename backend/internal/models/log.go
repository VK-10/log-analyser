package models

type LogEntry struct {
	Source     string `json:"source"`
	LogMessage string `json:"log_message"`
}

type ClassificationResult struct {
	Source string `json:"source"`
	Label  string `json:"label"`
}
