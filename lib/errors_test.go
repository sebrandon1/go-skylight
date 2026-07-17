package lib

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantErr    bool
		wantType   string
	}{
		{"ok 200", http.StatusOK, nil, false, ""},
		{"ok 201", http.StatusCreated, nil, false, ""},
		{"ok 204", http.StatusNoContent, nil, false, ""},
		{"auth 401", http.StatusUnauthorized, []byte("unauthorized"), true, "*lib.AuthError"},
		{"not found 404", http.StatusNotFound, nil, true, "*lib.NotFoundError"},
		{"rate limit 429", http.StatusTooManyRequests, nil, true, "*lib.RateLimitError"},
		{"server error 500", http.StatusInternalServerError, []byte("oops"), true, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			resp := &http.Response{
				StatusCode: tc.statusCode,
				Request:    req,
				Header:     make(http.Header),
			}
			err := checkStatus(resp, tc.body)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr {
				return
			}
			switch tc.wantType {
			case "*lib.AuthError":
				var ae *AuthError
				if !errors.As(err, &ae) {
					t.Errorf("want *AuthError, got %T", err)
				}
			case "*lib.NotFoundError":
				var nfe *NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("want *NotFoundError, got %T", err)
				}
			case "*lib.RateLimitError":
				var rle *RateLimitError
				if !errors.As(err, &rle) {
					t.Errorf("want *RateLimitError, got %T", err)
				}
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   time.Duration
	}{
		{name: "empty", header: "", want: 0},
		{name: "delta-5", header: "5", want: 5 * time.Second},
		{name: "delta-60", header: "60", want: 60 * time.Second},
		{name: "garbage", header: "abc", want: 0},
		{name: "negative", header: "-1", want: 0},
		{name: "padded-delta", header: " 30 ", want: 30 * time.Second},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseRetryAfter(tc.header)
			if got != tc.want {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tc.header, got, tc.want)
			}
		})
	}

	t.Run("http-date-future", func(t *testing.T) {
		// ~45s in the future; allow clock skew and scheduling latency.
		future := time.Now().UTC().Add(45 * time.Second).Format(http.TimeFormat)
		got := parseRetryAfter(future)
		if got < 30*time.Second || got > 45*time.Second {
			t.Errorf("parseRetryAfter(HTTP-date future) = %v, want ~45s (30-45s)", got)
		}
	})

	t.Run("http-date-past", func(t *testing.T) {
		past := time.Now().UTC().Add(-2 * time.Minute).Format(http.TimeFormat)
		got := parseRetryAfter(past)
		if got != 0 {
			t.Errorf("parseRetryAfter(HTTP-date past) = %v, want 0", got)
		}
	})
}

func TestErrorTypes(t *testing.T) {
	t.Run("AuthError", func(t *testing.T) {
		err := &AuthError{Message: "bad creds"}
		if err.Error() == "" {
			t.Error("empty error string")
		}
	})
	t.Run("NotFoundError with ID", func(t *testing.T) {
		err := &NotFoundError{Resource: "chore", ID: "ch1"}
		if err.Error() == "" {
			t.Error("empty error string")
		}
	})
	t.Run("NotFoundError without ID", func(t *testing.T) {
		err := &NotFoundError{Resource: "chore"}
		if err.Error() == "" {
			t.Error("empty error string")
		}
	})
	t.Run("RateLimitError with duration", func(t *testing.T) {
		err := &RateLimitError{RetryAfter: 5 * time.Second}
		if err.Error() == "" {
			t.Error("empty error string")
		}
	})
	t.Run("RateLimitError zero", func(t *testing.T) {
		err := &RateLimitError{}
		if err.Error() == "" {
			t.Error("empty error string")
		}
	})
	t.Run("NetworkError", func(t *testing.T) {
		cause := errors.New("dial tcp: timeout")
		err := &NetworkError{Cause: cause}
		if err.Error() == "" {
			t.Error("empty error string")
		}
		if !errors.Is(err, cause) {
			t.Error("Unwrap should return cause")
		}
	})
}
