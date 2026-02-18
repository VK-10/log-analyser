package classifier

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// wrappers for non-retryable errors
type permanentError struct{ err error }

func (e *permanentError) Error() string { return e.err.Error() }
func (e *permanentError) Unwrap() error { return e.err }

// Permanents error marks (non-retryable)
func Permanent(err error) error {
	return &permanentError{err}
}

func isPermanent(err error) bool {
	var p *permanentError
	return errors.As(err, &p)
}

func Retry(ctx context.Context, attempts int, fn func() (string, error)) (string, error) {
	var errs []error

	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return "", fmt.Errorf("context cancelled before attempt %d: %w", i+1, ctx.Err())
		}

		label, err := fn()
		if err == nil {
			return label, nil
		}

		if isPermanent(err) {
			return "", err
		}

		if i < attempts-1 {
			select {
			case <-time.After(time.Duration(i+1) * 100 * time.Duration(time.Millisecond)):
				//continue
			case <-ctx.Done():
				return "", fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
			}
		}

	}

	return "", errors.Join(errs...)
}
