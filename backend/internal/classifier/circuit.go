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

//call
func (cb *CircuitBreaker) Call[T any](fn func() (T, error)) (T, error) {
	var zero T

	cb.mu.Lock()

	// OPEN state
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenAttempts = 0
		} else {
			cb.mu.Unlock()
			return zero, ErrCircuitOpen
		}
	}

	// HALF-OPEN state
	if cb.state == StateHalfOpen {
		if cb.halfOpenAttempts >= cb.halfOpenMaxReq {
			cb.mu.Unlock()
			return zero, ErrTooManyRequests
		}
		cb.halfOpenAttempts++
	}

	cb.mu.Unlock()

	// ---- execute protected call ----
	result, err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailureLocked()
		return zero, err
	}

	cb.recordSuccessLocked()
	return result, nil
}


//state transitions

func (cb *CircuitBreaker) recordSuccessLocked() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		cb.state = StateClosed
		cb.failures = 0
		cb.halfOpenAttempts = 0
	}

	metrics.CircuitBreakerState.WithLabelValues("llm").Set(0)
}

func (cb *CircuitBreaker) recordFailureLocked() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			metrics.CircuitBreakerState.WithLabelValues("llm").Set(1)
		}
	case StateHalfOpen:
		cb.state = StateOpen
		metrics.CircuitBreakerState.WithLabelValues("llm").Set(1)
	}
}

//

// func (cb *CircuitBreaker) transitionToHalfOpen() {
// 	cb.mu.Lock()
// 	defer cb.mu.Unlock()

// 	if cb.state == StateOpen {
// 		cb.setState(StateHalfOpen)
// 		cb.halfOpenAttempts = 0
// 	}
// }

// func (cb *CircuitBreaker) setState(newState State) {
// 	cb.state = newState

// 	var stateValue float64
// 	if newState == StateOpen {
// 		stateValue = 1
// 	}

// 	metrics.CircuitBreakerState.WithLabelValues("llm").Set(stateValue)
// }

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// func (cb *CircuitBreaker) recordSuccess() {
// 	cb.mu.Lock()
// 	defer cb.mu.Unlock()

// 	switch cb.state {
// 	case StateClosed:
// 		cb.failures = 0
// 	case StateHalfOpen:
// 		cb.setState(StateClosed)
// 		cb.failures = 0
// 		cb.halfOpenAttempts = 0
// 	}
// }

// func (cb *CircuitBreaker) recordFailure() {
// 	cb.mu.Lock()
// 	defer cb.mu.Unlock()

// 	cb.failures++
// 	cb.lastFailureTime = time.Now()

// 	switch cb.state {
// 	case StateClosed:
// 		if cb.failures >= cb.maxFailures {
// 			cb.setState(StateOpen)
// 		}
// 	case StateHalfOpen:
// 		cb.setState(StateOpen)
// 	}

// }

// var breaker = CircuitBreaker{}

// func resetBreaker() {
// 	breaker.mu.Lock()
// 	defer breaker.mu.Unlock()

// 	breaker.failures = 0
// }


// Create breakers for different services
var (
	llmBreaker  = NewCircuitBreaker(3, 5*time.Second)
	bertBreaker = NewCircuitBreaker(5, 10*time.Second) // more tolerant
)
