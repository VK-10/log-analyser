package models

type LogEntry struct {
	Source     string `json:"source"`
	LogMessage string `json:"log_message"`
}

type ClassificationResult struct {
	LabelID    string  `json:"label_id"`
	Label      string  `json:"label"`
	Source     string  `json:"source"`
	Confidence float64 `json:"confidence"`
}
