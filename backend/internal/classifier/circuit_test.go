package classifier

import (
	"errors"
	"testing"
	"time"
)

func TestBreaker_AllowsCallsWhenClosed(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, time.Second)

	result, err := CallWithBreaker(cb, func() (string, error) {
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("unexpected result: %s", result)
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected CLOSED, got %v", cb.State())
	}
}

func TestBreaker_OpensAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, time.Second)

	failFn := func() (string, error) {
		return "", errors.New("fail")
	}

	// first failure
	_, _ = CallWithBreaker(cb, failFn)

	// second failure → should OPEN
	_, _ = CallWithBreaker(cb, failFn)

	if cb.State() != StateOpen {
		t.Fatalf("expected OPEN, got %v", cb.State())
	}
}

func TestBreaker_BlocksWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, time.Second)

	failFn := func() (string, error) {
		return "", errors.New("fail")
	}

	// cause OPEN
	_, _ = CallWithBreaker(cb, failFn)

	_, err := CallWithBreaker(cb, func() (string, error) {
		return "should not run", nil
	})

	if err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreaker_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 50*time.Millisecond)

	failFn := func() (string, error) {
		return "", errors.New("fail")
	}

	// open breaker
	_, _ = CallWithBreaker(cb, failFn)

	if cb.State() != StateOpen {
		t.Fatalf("expected OPEN, got %v", cb.State())
	}

	// wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// first probe allowed (HALF-OPEN)
	called := false
	_, err := CallWithBreaker(cb, func() (string, error) {
		called = true
		return "", errors.New("still failing")
	})

	if !called {
		t.Fatalf("half-open probe was not executed")
	}

	if err == nil {
		t.Fatalf("expected failure during half-open")
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected OPEN after failed probe, got %v", cb.State())
	}
}

func TestBreaker_ClosesAfterSuccessfulHalfOpenProbe(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 50*time.Millisecond)

	failFn := func() (string, error) {
		return "", errors.New("fail")
	}

	// open breaker
	_, _ = CallWithBreaker(cb, failFn)

	time.Sleep(60 * time.Millisecond)

	// successful probe
	result, err := CallWithBreaker(cb, func() (string, error) {
		return "recovered", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "recovered" {
		t.Fatalf("unexpected result: %s", result)
	}

	if cb.State() != StateClosed {
		t.Fatalf("expected CLOSED after recovery, got %v", cb.State())
	}
}

func TestBreaker_RejectsConcurrentHalfOpenRequests(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 50*time.Millisecond)

	failFn := func() (string, error) {
		return "", errors.New("fail")
	}

	// open breaker
	_, _ = CallWithBreaker(cb, failFn)
	time.Sleep(60 * time.Millisecond)

	// first probe allowed
	_, _ = CallWithBreaker(cb, func() (string, error) {
		return "", errors.New("probe fail")
	})

	// second probe should be rejected
	_, err := CallWithBreaker(cb, func() (string, error) {
		return "nope", nil
	})

	if err != ErrTooManyRequests {
		t.Fatalf("expected ErrTooManyRequests, got %v", err)
	}
}
