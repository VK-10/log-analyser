package classifier

import (
	"context"
	"time"
)

func ClassifyWithBERT(ctx context.Context, msg string) (string, error) {
	select {
	case <-time.After(300 * time.Millisecond):
		return "bert_label", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
