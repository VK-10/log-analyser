package classifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BERTRequest struct {
	Message string `json:"message"`
}

type BERTResponse struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
}

var bertClient = &http.Client{
	Timeout: 1 * time.Second,
}

const bertServiceURL = "http://bert-service:5000/classify"

func ClassifyWithBERT(ctx context.Context, msg string) (string, error) {
	reqBody := BERTRequest{Message: msg}
	jsonData, err := json.Marshal(reqBody)

	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", bertServiceURL, bytes.NewBuffer((jsonData)))

	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := bertClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call BERT service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("BERT service returned status %d: %s", resp.StatusCode, string(body))

	}

	var bertResp BERTResponse
	if err := json.NewDecoder(resp.Body).Decode(&bertResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return bertResp.Label, nil

}
