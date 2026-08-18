package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
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

// ValidationError is returned when the API responds with 422 Unprocessable
// Entity. Fields contains per-field error messages parsed from the JSON body
// when the body is a map[string][]string; it is nil when parsing fails.
type ValidationError struct {
	StatusCode int
	Body       string
	Fields     map[string][]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed (%d): %s", e.StatusCode, e.Body)
}

// HTTPError is returned for unexpected HTTP status codes (e.g. 4xx not handled
// by a more specific type, or 5xx after retries are exhausted). It preserves
// the status code so callers can distinguish responses without string-parsing.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error %d: %s", e.StatusCode, e.Body)
}

// NetworkError wraps a transport-level error (DNS, TCP, TLS).
type NetworkError struct {
	Cause error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Cause)
}

func (e *NetworkError) Unwrap() error { return e.Cause }

// IsNotFound reports whether err is (or wraps) a *NotFoundError.
func IsNotFound(err error) bool { return errors.As(err, new(*NotFoundError)) }

// IsValidation reports whether err is (or wraps) a *ValidationError.
func IsValidation(err error) bool { return errors.As(err, new(*ValidationError)) }

// IsHTTPError reports whether err is (or wraps) a *HTTPError.
func IsHTTPError(err error) bool { return errors.As(err, new(*HTTPError)) }

// checkStatus inspects resp.StatusCode and returns a typed error for non-2xx
// responses. body should be the already-read response body (may be nil).
func checkStatus(resp *http.Response, body []byte) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return &AuthError{Message: string(body)}
	case http.StatusNotFound:
		resource, id := parseNotFoundPath(resp.Request.URL.Path)
		return &NotFoundError{Resource: resource, ID: id}
	case http.StatusTooManyRequests:
		return &RateLimitError{RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	case http.StatusUnprocessableEntity:
		return &ValidationError{StatusCode: resp.StatusCode, Body: string(body), Fields: parseValidationFields(body)}
	default:
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}
}

// parseValidationFields best-effort: API body shape is not guaranteed; returns nil on parse failure.
func parseValidationFields(body []byte) map[string][]string {
	if len(body) == 0 {
		return nil
	}
	var fields map[string][]string
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil
	}
	return fields
}

// parseNotFoundPath extracts a resource type and optional resource ID from a
// REST URL path. Assumes paths of the form /…/{collection}/{id} where
// collection names are English plurals ending in "s".
func parseNotFoundPath(urlPath string) (resource, id string) {
	if urlPath == "" {
		return "", ""
	}
	last := path.Base(urlPath)
	if last == "." || last == "/" {
		return "", ""
	}
	prev := path.Base(path.Dir(urlPath))
	// If prev ends in "s" it is a collection name; last is the resource ID.
	if strings.HasSuffix(prev, "s") {
		return singularize(prev), last
	}
	return singularize(last), ""
}

func singularize(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	return strings.TrimSuffix(s, "s")
}

// parseRetryAfter parses an HTTP Retry-After header value per RFC 7231:
// either delta-seconds (integer) or an HTTP-date. Returns 0 on parse failure
// or when an HTTP-date is already in the past.
func parseRetryAfter(header string) time.Duration {
	header = strings.TrimSpace(header)
	if header == "" {
		return 0
	}
	// Prefer delta-seconds when the whole value is a non-negative integer.
	if secs, err := strconv.Atoi(header); err == nil {
		if secs <= 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	// HTTP-date form (e.g. "Thu, 10 Jul 2026 15:04:05 GMT").
	t, err := http.ParseTime(header)
	if err != nil {
		return 0
	}
	d := time.Until(t)
	if d <= 0 {
		return 0
	}
	return d
}
