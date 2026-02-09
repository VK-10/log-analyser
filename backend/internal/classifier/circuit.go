package classifier

import (
	"errors"
	"sync"
	"time"
)

type CircuitBreaker struct {
	failures int
	open     bool
	mu       sync.Mutex
}

var breaker = CircuitBreaker{}

func callLLMInternal(msg string) (string, error) {
	breaker.mu.Lock()
	if breaker.open {
		breaker.mu.Unlock()
		return "", errors.New("circuit open")
	}
	breaker.mu.Unlock()

	// Simulated LLM behavior
	var err error
	var label string

	if msg == "fail" {
		err = errors.New("llm failed")
	} else {
		label = "llm_label"
	}

	if err != nil {
		recordFailure()
		return "", err
	}

	resetBreaker()
	return label, nil
}

func recordFailure() {
	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	breaker.failures++
	if breaker.failures > 3 {
		breaker.open = true

		go func() {
			time.Sleep(5 * time.Second)
			breaker.mu.Lock()
			breaker.open = false
			breaker.failures = 0
			breaker.mu.Unlock()
		}()
	}
}

func resetBreaker() {
	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	breaker.failures = 0
}
