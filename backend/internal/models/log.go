package models

type LogEntry struct {
	Source     string `json:"source"`
	LogMessage string `json:"log_message"`
}

type ClassificationResult struct {
	LabelID    string  `json:"label_id"`
	Label      string  `json:"label"`
	Classifier string  `json:"classifier"`
	LogSource  string  `json:"log_source"`
	Confidence float64 `json:"confidence"`
}
