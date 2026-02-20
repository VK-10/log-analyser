package classifier

import (
	"context"
	"log-classifier/internal/models"
	"strings"
	"time"
)

func Classify(entry models.LogEntry) string {
	label := ClassifyWithRegex(entry.LogMessage)
	if label != "" {
		return label
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// marking circuit-open errors as permanent
	label, err := Retry(ctx, 2, func() (string, error) {
		result, err := ClassifyWithBERT(ctx, entry.LogMessage)
		if err != nil && strings.Contains(err.Error(), "circuit open") {
			return "", Permanent(err) // stops retry immediately
		}
		return result, err
	})

	if err == nil && label != "" {
		return label
	}

	//llm with timeout + retry
	llmCtx, cancelLLM := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelLLM()

	label, err = Retry(llmCtx, 2, func() (string, error) {
		return CallLLMWithTimeout(llmCtx, entry.LogMessage)
	})

	if err == nil && label != "" {
		return label
	}

	return "unknown"
}
