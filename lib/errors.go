package lib

import (
	"fmt"
	"net/http"
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
		resource, id := parseNotFoundFromPath(resp.Request.URL.Path)
		return &NotFoundError{Resource: resource, ID: id}
	case http.StatusTooManyRequests:
		return &RateLimitError{RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	default:
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
}

// parseNotFoundFromPath extracts a resource type and ID from an API path so
// NotFoundError can print e.g. chores "456" not found instead of the raw URL.
// Example: /api/frames/123/chores/456 → resource=chores, id=456
func parseNotFoundFromPath(p string) (resource, id string) {
	p = strings.Trim(p, "/")
	if p == "" {
		return "resource", ""
	}
	parts := strings.Split(p, "/")
	if len(parts) > 0 && parts[0] == "api" {
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return "resource", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[len(parts)-2], parts[len(parts)-1]
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
