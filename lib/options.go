package lib

import (
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// ClientOption configures a Client at construction time.
// Use the With* functions to build option slices:
//
//	client, err := lib.NewClientWithToken("uid", "tok",
//	    lib.WithBaseURL("https://staging.ourskylight.com/api"),
//	    lib.WithRetry(3, 500*time.Millisecond, 10*time.Second),
//	    lib.WithRateLimit(5, 10),
//	)
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL    string
	apiVersion string
	httpClient *http.Client
	logger     *slog.Logger
	limiter    *rate.Limiter
	retry      retryConfig
}

type retryConfig struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

// WithBaseURL overrides the Skylight API base URL. Primarily used in tests.
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) {
		c.baseURL = url
	}
}

// WithHTTPClient replaces the default http.Client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *clientConfig) {
		c.httpClient = hc
	}
}

// WithLogger enables slog-based debug logging. Authorization headers are
// redacted — only the scheme prefix ("Basic") is logged.
func WithLogger(l *slog.Logger) ClientOption {
	return func(c *clientConfig) {
		c.logger = l
	}
}

// WithRateLimit applies a token-bucket rate limiter to all outgoing requests.
// r is the steady-state rate (requests per second) and b is the burst size.
func WithRateLimit(r rate.Limit, b int) ClientOption {
	return func(c *clientConfig) {
		c.limiter = rate.NewLimiter(r, b)
	}
}

const defaultAPIVersion = "2026-03-01"

// WithAPIVersion overrides the skylight-api-version header sent with every
// request. The default is "2026-03-01".
func WithAPIVersion(version string) ClientOption {
	return func(c *clientConfig) {
		c.apiVersion = version
	}
}

// WithRetry configures exponential-backoff retry behavior. By default the
// client makes exactly one attempt. maxAttempts is the total number of tries
// (1 = no retry), baseDelay is the initial wait before the second attempt,
// and maxDelay caps the per-attempt wait.
func WithRetry(maxAttempts int, baseDelay, maxDelay time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.retry = retryConfig{
			maxAttempts: maxAttempts,
			baseDelay:   baseDelay,
			maxDelay:    maxDelay,
		}
	}
}

func defaultClientConfig() clientConfig {
	return clientConfig{
		baseURL:    SkylightURL,
		apiVersion: defaultAPIVersion,
		httpClient: newHTTPClient(),
		// Retry is disabled by default (maxAttempts=1 means one attempt, no retry).
		// Use WithRetry to opt in.
		retry: retryConfig{
			maxAttempts: 1,
			baseDelay:   500 * time.Millisecond,
			maxDelay:    30 * time.Second,
		},
	}
}
