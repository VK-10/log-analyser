package classifier

import (
	"context"
	"log-classifier/internal/models"
	"strings"
	"time"
)

func Classify(entry models.LogEntry) *models.ClassificationResult {

	// fmt.Printf("DEBUG LogMessage = %#v\n", entry.LogMessage)
	//regex
	if result := ClassifyWithRegex(entry.LogMessage); result != nil {
		return result
	}

	// bert
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	bertResult, err := Retry(ctx, 2, func() (*models.ClassificationResult, error) {
		result, err := ClassifyWithBERT(ctx, entry.LogMessage)
		if err != nil && strings.Contains(err.Error(), "circuit open") {
			return nil, Permanent(err) // stops retry immediately
		}
		return result, err
	})

	if err == nil && bertResult != nil {
		if bertResult.LabelID != "UNCLASSIFIED" && bertResult.Confidence >= 0.2 {
			return bertResult
		}
	}

	//llm with timeout + retry
	llmCtx, cancelLLM := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelLLM()

	llmResult, err := Retry(llmCtx, 2, func() (*models.ClassificationResult, error) {
		return CallLLMWithTimeout(llmCtx, entry.LogMessage)
	})

	if err == nil && llmResult != nil {
		return llmResult
	}

	return &models.ClassificationResult{
		LabelID:    "UNCLASSIFIED",
		Label:      "Unclassified",
		Source:     "orchestrator",
		Confidence: 0.0,
	}
}
