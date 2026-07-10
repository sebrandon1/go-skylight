package lib

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
		header string
		want   time.Duration
	}{
		{"", 0},
		{"5", 5 * time.Second},
		{"60", 60 * time.Second},
		{"abc", 0},
		{"-1", 0},
	}
	for _, tc := range tests {
		t.Run(tc.header, func(t *testing.T) {
			got := parseRetryAfter(tc.header)
			if got != tc.want {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tc.header, got, tc.want)
			}
		})
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

func TestParseNotFoundFromPath(t *testing.T) {
	cases := []struct {
		path         string
		wantResource string
		wantID       string
	}{
		{"/api/frames/123/chores/456", "chores", "456"},
		{"/api/frames/123", "frames", "123"},
		{"/api/test", "test", ""},
		{"/", "resource", ""},
		{"", "resource", ""},
		{"frames/1/lists/2/list_items/3", "list_items", "3"},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			r, id := parseNotFoundFromPath(c.path)
			if r != c.wantResource || id != c.wantID {
				t.Errorf("parseNotFoundFromPath(%q) = (%q,%q), want (%q,%q)", c.path, r, id, c.wantResource, c.wantID)
			}
		})
	}
}

func TestCheckStatusNotFoundPopulatesID(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantResource string
		wantID       string
		wantInMsg    string
		wantNotInMsg string
	}{
		{
			name:         "nested resource with id",
			path:         "/api/frames/123/chores/456",
			wantResource: "chores",
			wantID:       "456",
			wantInMsg:    "456",
			wantNotInMsg: "/api/",
		},
		{
			name:         "collection-level 404 has no id",
			path:         "/api/test",
			wantResource: "test",
			wantID:       "",
			wantInMsg:    "test",
			wantNotInMsg: "/api/",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			resp := &http.Response{StatusCode: http.StatusNotFound, Request: req, Header: make(http.Header)}
			err := checkStatus(resp, nil)
			var nfe *NotFoundError
			if !errors.As(err, &nfe) {
				t.Fatalf("want *NotFoundError, got %T %v", err, err)
			}
			if nfe.Resource != tc.wantResource || nfe.ID != tc.wantID {
				t.Fatalf("got resource=%q id=%q, want resource=%q id=%q", nfe.Resource, nfe.ID, tc.wantResource, tc.wantID)
			}
			msg := nfe.Error()
			if !strings.Contains(msg, tc.wantInMsg) || strings.Contains(msg, tc.wantNotInMsg) {
				t.Fatalf("Error() = %q, want contains %q and not contains %q", msg, tc.wantInMsg, tc.wantNotInMsg)
			}
		})
	}
}
