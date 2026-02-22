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

type BERTRequest struct {
	Message string `json:"message"`
}

type BERTResponse struct {
	LabelID    string  `json:"label_id"`
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
}

var bertClient = &http.Client{
	Timeout: 5 * time.Second,
}

const bertServiceURL = "http://127.0.0.1:5000/classify"

func ClassifyWithBERT(ctx context.Context, msg string) (*models.ClassificationResult, error) {
	fmt.Println("DEBUG: BERT CALLED with:", msg)
	reqBody := BERTRequest{Message: msg}
	jsonData, err := json.Marshal(reqBody)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", bertServiceURL, bytes.NewBuffer((jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("DEBUG: BERT HTTP request starting")

	resp, err := bertClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call BERT service: %w", err)
	}
	fmt.Println("DEBUG: BERT HTTP status:", resp.StatusCode)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("BERT service returned status %d: %s", resp.StatusCode, string(body))

	}

	var bertResp BERTResponse
	if err := json.NewDecoder(resp.Body).Decode(&bertResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("DEBUG: BERT RAW RESPONSE: %+v\n", bertResp)

	// return bertResp.Label, nil
	return &models.ClassificationResult{
		LabelID:    bertResp.LabelID,
		Label:      bertResp.Label,
		Source:     "classifier",
		Confidence: bertResp.Confidence,
	}, nil

}
