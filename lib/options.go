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

	tokenExchange  TokenExchangeFunc
	onTokenRefresh func(refreshToken string) error
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

const defaultAPIVersion = "2026-08-05"

// WithAPIVersion overrides the skylight-api-version header sent with every
// request. The default is "2026-08-05".
func WithAPIVersion(version string) ClientOption {
	return func(c *clientConfig) {
		c.apiVersion = version
	}
}

// WithRetry configures exponential-backoff retry behavior. By default the
// client makes exactly one attempt. maxAttempts is the total number of tries
// (1 = no retry), baseDelay is the initial wait before the second attempt,
// and maxDelay caps the per-attempt wait.
// TokenExchangeFunc exchanges a refresh token for new tokens.
type TokenExchangeFunc func(refreshToken, fingerprint string) (*OAuthTokenResponse, error)

// WithTokenExchange replaces RefreshOAuthToken as the way a client built with
// NewClientWithRefreshToken renews its tokens. Use it when tokens are shared
// between processes and the exchange must be coordinated with their storage.
func WithTokenExchange(fn TokenExchangeFunc) ClientOption {
	return func(c *clientConfig) { c.tokenExchange = fn }
}

// WithTokenRefreshHandler registers fn to receive each refresh token issued to
// the client, including the one from NewClientWithRefreshToken. Refresh tokens
// are single-use, so a caller that keeps tokens across restarts must save each
// one. An error from fn is logged and does not fail the request.
func WithTokenRefreshHandler(fn func(refreshToken string) error) ClientOption {
	return func(c *clientConfig) { c.onTokenRefresh = fn }
}

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
