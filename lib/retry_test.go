package lib

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

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
	var se *HTTPError
	if !errors.As(err, &se) {
		t.Errorf("expected *HTTPError after exhausted retries, got %T: %v", err, err)
	}
	if se != nil && se.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", se.StatusCode, http.StatusInternalServerError)
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

func TestDoWithRetryWithRateLimiter(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	limiter := rate.NewLimiter(rate.Limit(100), 10)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	cfg := retryConfig{maxAttempts: 1, baseDelay: 10 * time.Millisecond, maxDelay: 100 * time.Millisecond}

	resp, err := doWithRetry(context.Background(), http.DefaultClient, limiter, cfg, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if calls.Load() != 1 {
		t.Errorf("want 1 call, got %d", calls.Load())
	}
}

func TestDoWithRetryMaxAttemptsZero(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	cfg := retryConfig{maxAttempts: 0, baseDelay: 10 * time.Millisecond, maxDelay: 100 * time.Millisecond}

	resp, err := doWithRetry(context.Background(), http.DefaultClient, nil, cfg, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if calls.Load() != 1 {
		t.Errorf("maxAttempts=0 should default to 1, got %d calls", calls.Load())
	}
}

func TestDoWithRetry_BodyReplayOnRetry(t *testing.T) {
	type testPayload struct {
		Value string `json:"value"`
	}

	var calls atomic.Int32
	var mu sync.Mutex
	var bodies []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		mu.Unlock()
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req, _ := newRequestWithBody(context.Background(), "POST", srv.URL, testPayload{Value: "hello"})
	cfg := retryConfig{maxAttempts: 2, baseDelay: 5 * time.Millisecond, maxDelay: 20 * time.Millisecond}

	resp, err := doWithRetry(context.Background(), http.DefaultClient, nil, cfg, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if calls.Load() != 2 {
		t.Errorf("want 2 calls, got %d", calls.Load())
	}
	mu.Lock()
	b := bodies
	mu.Unlock()
	if len(b) < 2 {
		t.Fatalf("want 2 request bodies captured, got %d", len(b))
	}
	if b[0] == "" || b[1] == "" {
		t.Errorf("expected non-empty body on both attempts, got %q and %q", b[0], b[1])
	}
	if b[0] != b[1] {
		t.Errorf("body mismatch: first=%q second=%q", b[0], b[1])
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
		var ne *NetworkError
		if !errors.As(err, &ne) {
			t.Errorf("expected DeadlineExceeded or *NetworkError, got %T: %v", err, err)
		}
	}
}

func TestDoWithRetry_ContextCancelledMidFlight(t *testing.T) {
	ready := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(ready)
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() { <-ready; cancel() }()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	cfg := retryConfig{maxAttempts: 3, baseDelay: 10 * time.Millisecond, maxDelay: 100 * time.Millisecond}
	_, err := doWithRetry(ctx, http.DefaultClient, nil, cfg, req)
	if err == nil {
		t.Fatal("expected error from canceled context mid-flight")
	}
	if !errors.Is(err, context.Canceled) {
		var ne *NetworkError
		if !errors.As(err, &ne) {
			t.Errorf("expected context.Canceled or *NetworkError, got %T: %v", err, err)
		}
	}
}
