package lib

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackoffDelay(t *testing.T) {
	base := 100 * time.Millisecond
	max := 5 * time.Second

	for attempt := 1; attempt <= 6; attempt++ {
		d := backoffDelay(base, max, attempt)
		if d < 0 {
			t.Errorf("attempt %d: negative delay %v", attempt, d)
		}
		if d > max {
			t.Errorf("attempt %d: delay %v exceeds max %v", attempt, d, max)
		}
	}
}

func TestDoWithRetrySuccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	cfg := retryConfig{maxAttempts: 3, baseDelay: 10 * time.Millisecond, maxDelay: 100 * time.Millisecond}

	resp, err := doWithRetry(context.Background(), http.DefaultClient, nil, cfg, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if calls.Load() != 1 {
		t.Errorf("want 1 call, got %d", calls.Load())
	}
}

func TestDoWithRetryOn5xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	cfg := retryConfig{maxAttempts: 5, baseDelay: 5 * time.Millisecond, maxDelay: 50 * time.Millisecond}

	resp, err := doWithRetry(context.Background(), http.DefaultClient, nil, cfg, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if calls.Load() != 3 {
		t.Errorf("want 3 calls, got %d", calls.Load())
	}
}

func TestDoWithRetryExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	cfg := retryConfig{maxAttempts: 3, baseDelay: 5 * time.Millisecond, maxDelay: 20 * time.Millisecond}

	_, err := doWithRetry(context.Background(), http.DefaultClient, nil, cfg, req)
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
}

func TestDoWithRetryRespects429RetryAfter(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0") // 0 = no sleep but still retry
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	cfg := retryConfig{maxAttempts: 3, baseDelay: 5 * time.Millisecond, maxDelay: 20 * time.Millisecond}

	resp, err := doWithRetry(context.Background(), http.DefaultClient, nil, cfg, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if calls.Load() != 2 {
		t.Errorf("want 2 calls (1 retry after 429), got %d", calls.Load())
	}
}

func TestDoWithRetryContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	cfg := retryConfig{maxAttempts: 10, baseDelay: 50 * time.Millisecond, maxDelay: 500 * time.Millisecond}

	_, err := doWithRetry(ctx, http.DefaultClient, nil, cfg, req)
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		// Context may surface as NetworkError wrapping the deadline error.
		var ne *NetworkError
		if !errors.As(err, &ne) {
			t.Logf("got error type %T: %v", err, err)
		}
	}
}
