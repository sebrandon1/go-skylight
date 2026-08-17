package cmd

import (
	"errors"
	"testing"
	"time"
)

func TestRunConcurrent_AllPass(t *testing.T) {
	var a, b int
	err := runConcurrent(
		func() error { a = 1; return nil },
		func() error { b = 2; return nil },
	)
	if err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
	if a != 1 || b != 2 {
		t.Errorf("expected side effects applied, got a=%d b=%d", a, b)
	}
}

func TestRunConcurrent_AllErrorsJoined(t *testing.T) {
	first := errors.New("first error")
	second := errors.New("second error")
	err := runConcurrent(
		func() error { return first },
		func() error { return second },
	)
	if !errors.Is(err, first) {
		t.Errorf("expected first error in joined result, got: %v", err)
	}
	if !errors.Is(err, second) {
		t.Errorf("expected second error in joined result, got: %v", err)
	}
}

func TestRunConcurrent_OnlyOneErrors(t *testing.T) {
	sentinel := errors.New("only error")
	err := runConcurrent(
		func() error { return sentinel },
		func() error { return nil },
	)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
}

func TestRunConcurrent_Empty(t *testing.T) {
	if err := runConcurrent(); err != nil {
		t.Errorf("expected nil for empty input, got: %v", err)
	}
}

func TestRunConcurrent_RunsConcurrently(t *testing.T) {
	// Two 50ms sleeps should finish in ~50ms total, not 100ms.
	start := time.Now()
	if err := runConcurrent(
		func() error { time.Sleep(50 * time.Millisecond); return nil },
		func() error { time.Sleep(50 * time.Millisecond); return nil },
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 80*time.Millisecond {
		t.Errorf("expected concurrent execution (<80ms), took %v", elapsed)
	}
}
