package lib

import (
	"net/http"
	"time"

	"github.com/sebrandon1/go-toolbox/httpx"
)

// The error types are aliases of httpx's types so existing callers keep
// compiling while the implementation lives in go-toolbox.
type (
	// AuthError is returned when the API responds with 401 Unauthorized.
	AuthError = httpx.AuthError
	// NotFoundError is returned when the API responds with 404 Not Found.
	NotFoundError = httpx.NotFoundError
	// RateLimitError is returned when the API responds with 429 Too Many Requests.
	RateLimitError = httpx.RateLimitError
	// ValidationError is returned when the API responds with 422 Unprocessable Entity.
	ValidationError = httpx.ValidationError
	// HTTPError is returned for unexpected HTTP status codes.
	HTTPError = httpx.HTTPError
	// NetworkError wraps a transport-level error.
	NetworkError = httpx.NetworkError
)

// IsAuthError reports whether err is (or wraps) an *AuthError.
func IsAuthError(err error) bool { return httpx.IsAuthError(err) }

// IsNotFound reports whether err is (or wraps) a *NotFoundError.
func IsNotFound(err error) bool { return httpx.IsNotFound(err) }

// IsRateLimited reports whether err is (or wraps) a *RateLimitError.
func IsRateLimited(err error) bool { return httpx.IsRateLimited(err) }

// IsValidation reports whether err is (or wraps) a *ValidationError.
func IsValidation(err error) bool { return httpx.IsValidation(err) }

// IsHTTPError reports whether err is (or wraps) a *HTTPError.
func IsHTTPError(err error) bool { return httpx.IsHTTPError(err) }

// checkStatus inspects resp.StatusCode and returns a typed error for non-2xx
// responses. body should be the already-read response body (may be nil).
func checkStatus(resp *http.Response, body []byte) error {
	return httpx.CheckStatus(resp, body)
}

// parseNotFoundPath extracts a resource type and optional resource ID from a
// REST URL path.
func parseNotFoundPath(urlPath string) (resource, id string) {
	return httpx.ParseNotFoundPath(urlPath)
}

// parseRetryAfter parses an HTTP Retry-After header value per RFC 9110.
func parseRetryAfter(header string) time.Duration {
	return httpx.ParseRetryAfter(header)
}
