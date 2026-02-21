package classifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log-classifier/internal/models"
	"net/http"
	"time"
)

var llmClient = &http.Client{
	Timeout: 3 * time.Second,
}

const llmServiceURL = "http://llm-service:5001/classify"

func callLLMInternal(msg string) (*models.ClassificationResult, error) {
	fmt.Println("DEBUG: BERT CALLED with:", msg)
	return CallWithBreaker(llmBreaker, func() (*models.ClassificationResult, error) {

		reqBody := map[string]string{
			"message": msg,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal LLM request: %w", err)
		}

		resp, err := llmClient.Post(llmServiceURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to call LLM service: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("LLM service returned status %d: %s", resp.StatusCode, string(body))
		}

		var result models.ClassificationResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode LLM response: %w", err)
		}

		return &result, nil
	})
}

// Public API with timeout
func CallLLMWithTimeout(ctx context.Context, msg string) (*models.ClassificationResult, error) {
	resultCh := make(chan *models.ClassificationResult, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := callLLMInternal(msg)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	select {
	case label := <-resultCh:
		return label, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
