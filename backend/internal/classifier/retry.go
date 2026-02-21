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
	return &permanentError{err: err}
}

func isPermanent(err error) bool {
	var p *permanentError
	return errors.As(err, &p)
}

func Retry[T any](ctx context.Context, attempts int, fn func() (T, error)) (T, error) {
	var zero T
	var errs []error

	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return zero, fmt.Errorf("context cancelled before attempt %d: %w", i+1, ctx.Err())
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		if isPermanent(err) {
			return zero, err
		}

		errs = append(errs, fmt.Errorf("attempt %d: %w", i+1, err))

		if i < attempts-1 {
			select {
			case <-time.After(time.Duration(i+1) * 100 * time.Duration(time.Millisecond)):
				//continue
			case <-ctx.Done():
				return zero, fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
			}
		}

	}

	return zero, errors.Join(errs...)
}
