package classifier

import (
	"context"
	"log-classifier/backend/internal/models"
	"time"
)

func Classify(entry models.LogEntry) string {
	label := ClassifyWithRegex(entry.LogMessage)
	if label != "" {
		return label
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	label, err := Retry(2, func() (string, error) {
		return ClassifyWithBERT(ctx, entry.LogMessage)
	})

	if err == nil && label != "" {
		return label
	}

	//llm with timeout + retry
	llmCtx, cancelLLM := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelLLM()

	label, err = Retry(2, func() (string, error) {
		return CallLLMWithTimeout(llmCtx, entry.LogMessage)
	})

	if err == nil && label != "" {
		return label
	}

	return "unknown"
}
