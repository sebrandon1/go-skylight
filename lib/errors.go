package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AuthError is returned when the API responds with 401 Unauthorized.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Message)
}

// NotFoundError is returned when the API responds with 404 Not Found.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s %q not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// RateLimitError is returned when the API responds with 429 Too Many Requests.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited: retry after %s", e.RetryAfter)
	}
	return "rate limited"
}

// NetworkError wraps a transport-level error (DNS, TCP, TLS).
type NetworkError struct {
	Cause error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Cause)
}

func (e *NetworkError) Unwrap() error { return e.Cause }

// checkStatus inspects resp.StatusCode and returns a typed error for non-2xx
// responses. body should be the already-read response body (may be nil).
func checkStatus(resp *http.Response, body []byte) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return &AuthError{Message: string(body)}
	case http.StatusNotFound:
		return &NotFoundError{Resource: resp.Request.URL.Path}
	case http.StatusTooManyRequests:
		return &RateLimitError{RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	default:
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, formatAPIErrorBody(body))
	}
}

// formatAPIErrorBody turns common Skylight/Rails-style JSON error bodies into
// a short human-readable string. Non-JSON or unrecognized shapes are returned
// as the original body text (trimmed).
func formatAPIErrorBody(body []byte) string {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return "(empty body)"
	}

	// {"errors":{"title":["can't be blank"],"points":["must be greater than 0"]}}
	var nested struct {
		Errors map[string][]string `json:"errors"`
	}
	if err := json.Unmarshal(body, &nested); err == nil && len(nested.Errors) > 0 {
		parts := make([]string, 0, len(nested.Errors))
		for field, msgs := range nested.Errors {
			if len(msgs) == 0 {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s: %s", field, strings.Join(msgs, "; ")))
		}
		if len(parts) > 0 {
			sort.Strings(parts) // map iteration order is random
			return strings.Join(parts, "; ")
		}
	}

	// {"error":"message"} or {"message":"..."} or {"errors":["a","b"]}
	var flat map[string]json.RawMessage
	if err := json.Unmarshal(body, &flat); err == nil {
		for _, key := range []string{"error", "message", "detail", "error_description"} {
			if v, ok := flat[key]; ok {
				var s string
				if json.Unmarshal(v, &s) == nil && s != "" {
					return s
				}
			}
		}
		if v, ok := flat["errors"]; ok {
			var list []string
			if json.Unmarshal(v, &list) == nil && len(list) > 0 {
				return strings.Join(list, "; ")
			}
		}
	}

	return raw
}

// parseRetryAfter parses an HTTP Retry-After header value (seconds as integer
// or delta-seconds). Returns 0 on parse failure.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	secs, err := strconv.Atoi(header)
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}
