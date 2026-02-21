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

type LLMRequest struct {
	Message string `json:"message"`
}

type LLMResponse struct {
	Label string `json:"label"`
}

var llmClient = &http.Client{
	Timeout: 3 * time.Second,
}

const llmServiceURL = "http://llm-service:5001/classify"

var llmLabelMap = map[string]ClassificationResult{
	"workflow error": {
		LabelID:    "WORKFLOW_ERROR",
		Label:      "Workflow Error",
		Source:     "llm",
		Confidence: 0.65,
	},
	"deprecation warning": {
		LabelID:    "DEPRECATION_WARNING",
		Label:      "Deprecation Warning",
		Source:     "llm",
		Confidence: 0.65,
	},
	"unclassified": {
		LabelID:    "UNCLASSIFIED",
		Label:      "Unclassified",
		Source:     "llm",
		Confidence: 0.30,
	},
}

func callLLMInternal(msg string) (string, error) {
	return llmBreaker.Call(func() (string, error) {
		reqBody := LLMRequest{Message: msg}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("failed to marshal LLM request: %w", err)
		}

		resp, err := llmClient.Post(llmServiceURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to call LLM service: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("LLM service returned status %d: %s", resp.StatusCode, string(body))
		}

		var llmResp LLMResponse
		if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
			return "", fmt.Errorf("failed to decode LLM response: %w", err)
		}

		return llmResp.Label, nil
	})
}

func CallLLMWithTimeout(ctx context.Context, msg string) (string, error) {
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		label, err := callLLMInternal(msg)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- label
	}()

	select {
	case label := <-resultCh:
		return label, nil
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
