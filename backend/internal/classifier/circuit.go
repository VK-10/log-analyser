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
	name           string
	maxFailures    int
	resetTimeout   time.Duration
	halfOpenMaxReq int

	mu               sync.Mutex
	state            State
	failures         int
	lastFailureTime  time.Time
	halfOpenAttempts int
}

func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:           name,
		maxFailures:    maxFailures,
		resetTimeout:   resetTimeout,
		halfOpenMaxReq: 1, //one test at time of tesing
		state:          StateClosed,
	}
}

var (
	ErrCircuitOpen     = errors.New("Cicuit breaker is open")
	ErrTooManyRequests = errors.New("circuit breaker: too many requests")
)

// call
func CallWithBreaker[T any](cb *CircuitBreaker, fn func() (T, error)) (T, error) {
	var zero T

	cb.mu.Lock()

	//open

	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenAttempts = 0
		} else {
			cb.mu.Unlock()
			return zero, ErrCircuitOpen
		}
	}

	//half-open
	if cb.state == StateHalfOpen {
		if cb.halfOpenAttempts >= cb.halfOpenMaxReq {
			cb.mu.Unlock()
			return zero, ErrTooManyRequests
		}
		cb.halfOpenAttempts++
	}
	cb.mu.Unlock()

	//executing the protected function
	result, err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Record result
	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()
		if cb.state == StateHalfOpen || cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			metrics.CircuitBreakerState.WithLabelValues(cb.name).Set(1)

		}

		return zero, err
	}

	//success
	cb.failures = 0
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.halfOpenAttempts = 0
	}
	metrics.CircuitBreakerState.WithLabelValues(cb.name).Set(0)
	return result, nil
}

//state

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Create breakers for different services
var (
	llmBreaker  = NewCircuitBreaker("llm", 3, 5*time.Second)
	bertBreaker = NewCircuitBreaker("bert", 5, 10*time.Second) // more tolerant
)
