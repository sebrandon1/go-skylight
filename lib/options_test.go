package lib

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestWithBaseURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(calendarAPIResponse{Data: []calendarAPIEntry{{ID: "1"}}}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	client, err := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"))
	if err != nil {
		t.Fatalf("NewClientWithToken: %v", err)
	}
	if client.baseURL != srv.URL+"/api" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, srv.URL+"/api")
	}
	events, err := client.ListCalendarEvents("f1", "", "")
	if err != nil {
		t.Fatalf("ListCalendarEvents: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("want 1 event, got %d", len(events))
	}
}

func TestWithHTTPClient(t *testing.T) {
	hc := &http.Client{Timeout: 5 * time.Second}
	client, err := NewClientWithToken("u", "t", WithHTTPClient(hc))
	if err != nil {
		t.Fatalf("NewClientWithToken: %v", err)
	}
	if client.HTTPClient != hc {
		t.Error("HTTPClient not set")
	}
}

func TestWithRateLimit(t *testing.T) {
	client, err := NewClientWithToken("u", "t", WithRateLimit(rate.Limit(10), 5))
	if err != nil {
		t.Fatalf("NewClientWithToken: %v", err)
	}
	if client.limiter == nil {
		t.Error("limiter should not be nil")
	}
}

func TestWithRetry(t *testing.T) {
	client, err := NewClientWithToken("u", "t", WithRetry(5, 100*time.Millisecond, 5*time.Second))
	if err != nil {
		t.Fatalf("NewClientWithToken: %v", err)
	}
	if client.retry.maxAttempts != 5 {
		t.Errorf("maxAttempts = %d, want 5", client.retry.maxAttempts)
	}
	if client.retry.baseDelay != 100*time.Millisecond {
		t.Errorf("baseDelay = %v, want 100ms", client.retry.baseDelay)
	}
}

func TestWithLogger(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(calendarAPIResponse{Data: []calendarAPIEntry{{ID: "1"}}}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client, err := NewClientWithToken("u", "t",
		WithBaseURL(srv.URL+"/api"),
		WithLogger(logger),
	)
	if err != nil {
		t.Fatalf("NewClientWithToken: %v", err)
	}
	if client.logger != logger {
		t.Error("logger not set on client")
	}
	if _, err = client.ListCalendarEvents("f1", "", ""); err != nil {
		t.Fatalf("ListCalendarEvents: %v", err)
	}
}
