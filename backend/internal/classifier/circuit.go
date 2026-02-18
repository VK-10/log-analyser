package classifier

import (
	"errors"
	"log-classifier/internal/metrics"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	maxFailures    int
	resetTimeout   time.Duration
	halfOpenMaxReq int

	mu               sync.RWMutex
	state            State
	failures         int
	lastFailureTime  time.Time
	halfOpenAttempts int
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:    maxFailures,
		resetTimeout:   resetTimeout,
		halfOpenMaxReq: 1, //one test at time of tesing
		state:          StateClosed,
	}
}

var (
	ErrCicuitOpen      = errors.New("Cicuit breaker is open")
	ErrTooManyRequests = errors.New("circuit breaker: too many requests")
)

func (cb *CircuitBreaker) Call(fn func() (string, error)) (string, error) {
	cb.mu.RLock()
	state := cb.state

	if state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mu.RUnlock()
			cb.transitionToHalfOpen()
			return cb.Call(fn)
		}
		cb.mu.RUnlock()
		return "", ErrCicuitOpen
	}

	if state == StateHalfOpen {
		if cb.halfOpenAttempts >= cb.halfOpenMaxReq {
			cb.mu.RUnlock()
			return "", ErrTooManyRequests
		}
		cb.halfOpenAttempts++
	}
	cb.mu.RUnlock()

	//executing the function
	result, err := fn()

	// Record result
	if err != nil {
		cb.recordFailure()
		return "", err
	}

	cb.recordSuccess()
	return result, nil
}

func (cb *CircuitBreaker) transitionToHalfOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen {
		cb.setState(StateHalfOpen)
		cb.halfOpenAttempts = 0
	}
}

func (cb *CircuitBreaker) setState(newState State) {
	cb.state = newState

	var stateValue float64
	if newState == StateOpen {
		stateValue = 1
	}

	metrics.CircuitBreakerState.WithLabelValues("llm").Set(stateValue)
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		cb.setState(StateClosed)
		cb.failures = 0
		cb.halfOpenAttempts = 0
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.setState(StateOpen)
		}
	case StateHalfOpen:
		cb.setState(StateOpen)
	}

}

var breaker = CircuitBreaker{}

func resetBreaker() {
	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	breaker.failures = 0
}

// Create breakers for different services
var (
	llmBreaker  = NewCircuitBreaker(3, 5*time.Second)
	bertBreaker = NewCircuitBreaker(5, 10*time.Second) // more tolerant
)

func callLLMInternal(msg string) (string, error) {
	return llmBreaker.Call(func() (string, error) {

		if msg == "fail" {
			return "", errors.New("llm failed")
		}
		return "llm_label", nil
	})
}
