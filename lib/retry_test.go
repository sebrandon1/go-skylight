package lib

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
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

func TestBackoffDelayMaxClamping(t *testing.T) {
	base := 100 * time.Millisecond
	max := 1 * time.Second
	d := backoffDelay(base, max, 20)
	if d > max {
		t.Errorf("delay %v exceeds max %v", d, max)
	}
	if d < max/2 {
		t.Errorf("delay %v too low, expected at least %v", d, max/2)
	}
}

func TestBackoffDelayHalfZero(t *testing.T) {
	// base=1ns, max=1ns -> exp=1ns, exp/2=0 -> early return exp
	d := backoffDelay(1*time.Nanosecond, 1*time.Nanosecond, 1)
	if d != 1*time.Nanosecond {
		t.Errorf("delay = %v, want 1ns", d)
	}
}

func TestDrainAndErrorNon429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("server error body")); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	drainErr := drainAndError(context.Background(), resp)
	if drainErr == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(drainErr.Error(), "server error body") {
		t.Errorf("error should contain body, got %q", drainErr.Error())
	}
}

func TestDrainAndError429ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	drainErr := drainAndError(ctx, resp)
	if !errors.Is(drainErr, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", drainErr)
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

func TestBufferBodyReadError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", io.NopCloser(&errorReader{}))
	_, err := bufferBody(req)
	if err == nil {
		t.Fatal("expected error")
	}
	var ne *NetworkError
	if !errors.As(err, &ne) {
		t.Errorf("expected NetworkError, got %T", err)
	}
}

type errorReader struct{}

func (e *errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read error")
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

func TestDrainAndError_429Instant(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Retry-After": []string{"0"}},
		Body:       io.NopCloser(strings.NewReader("")),
	}

	drainErr := drainAndError(context.Background(), resp)
	if drainErr == nil {
		t.Fatal("expected error")
	}
	var rle *RateLimitError
	if !errors.As(drainErr, &rle) {
		t.Errorf("expected *RateLimitError, got %T: %v", drainErr, drainErr)
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
