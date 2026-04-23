package service

import (
	"errors"
	"testing"
	"time"
)

func TestWithRetry_SucceedsAfterRetries(t *testing.T) {
	t.Parallel()

	attempts := 0
	got, err := withRetry(3, 0, "test", func() (int, error) {
		attempts++
		if attempts < 3 {
			return 0, errors.New("temporary failure")
		}
		return 42, nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 42 {
		t.Fatalf("expected result 42, got %d", got)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestWithRetry_ReturnsLastError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("always failing")
	attempts := 0
	_, err := withRetry[string](2, 0, "test", func() (string, error) {
		attempts++
		return "", sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestWithRetry_WaitsBetweenAttempts(t *testing.T) {
	t.Parallel()

	start := time.Now()
	attempts := 0
	_, _ = withRetry[int](2, 20*time.Millisecond, "test", func() (int, error) {
		attempts++
		return 0, errors.New("fail")
	})

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		t.Fatalf("expected at least 20ms elapsed, got %v", elapsed)
	}
}
