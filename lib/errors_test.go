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

func TestParseNotFoundPath(t *testing.T) {
	tests := []struct {
		path         string
		wantResource string
		wantID       string
	}{
		// specific resource: collection + ID
		{"/api/frames/frame1/chores/chore1", "chore", "chore1"},
		{"/api/frames/frame1/rewards/reward1", "reward", "reward1"},
		{"/api/frames/frame1/calendar_events/ev1", "calendar_event", "ev1"},
		{"/api/frames/frame1/categories/cat1", "category", "cat1"},
		{"/api/frames/frame1/lists/list1", "list", "list1"},
		{"/api/frames/frame1/routines/rt1", "routine", "rt1"},
		{"/api/frames/frame1/meals/sittings/sit1", "sitting", "sit1"},
		{"/api/frames/frame1/meals/recipes/rec1", "recipe", "rec1"},
		// frame itself
		{"/api/frames/frame1", "frame", "frame1"},
		// collection only (no specific ID in path)
		{"/api/frames/frame1/chores", "chore", ""},
		// edge cases
		{"", "", ""},
		{"/single", "single", ""},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			r, id := parseNotFoundPath(tc.path)
			if r != tc.wantResource {
				t.Errorf("resource: got %q, want %q", r, tc.wantResource)
			}
			if id != tc.wantID {
				t.Errorf("id: got %q, want %q", id, tc.wantID)
			}
		})
	}
}

func TestCheckStatus404IDPopulated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/frames/frame1/chores/chore1", nil)
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Request:    req,
		Header:     make(http.Header),
	}
	err := checkStatus(resp, nil)
	var nfe *NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("want *NotFoundError, got %T", err)
	}
	if nfe.Resource != "chore" {
		t.Errorf("resource: got %q, want %q", nfe.Resource, "chore")
	}
	if nfe.ID != "chore1" {
		t.Errorf("id: got %q, want %q", nfe.ID, "chore1")
	}
	want := `chore "chore1" not found`
	if nfe.Error() != want {
		t.Errorf("error string: got %q, want %q", nfe.Error(), want)
	}
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
