package lib

import (
	"context"
	"net/http"

	"github.com/sebrandon1/go-toolbox/retry"
)

// limiterIface is a subset of *rate.Limiter so the caller can pass nil without
// importing golang.org/x/time/rate in every file.
type limiterIface interface {
	Wait(ctx context.Context) error
}

// doWithRetry executes req via httpClient, retrying on transient errors and
// 5xx / 429 responses according to cfg. The rate limiter is consulted before
// each attempt. Callers must not reuse req after this call.
//
// A buffered copy of the request body is kept so it can be replayed on retry.
func doWithRetry(ctx context.Context, httpClient *http.Client, limiter limiterIface, cfg retryConfig, req *http.Request) (*http.Response, error) {
	opts := []retry.HTTPOption{
		retry.WithMaxAttempts(cfg.maxAttempts),
		retry.WithBaseDelay(cfg.baseDelay),
		retry.WithHTTPMaxDelay(cfg.maxDelay),
		// Skylight historically retried every method, and every 5xx (not just
		// the usual transient set), so preserve that behavior.
		retry.WithRetryMethods(
			http.MethodGet, http.MethodHead, http.MethodPost,
			http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions,
		),
		retry.WithRetryStatus(func(code int) bool {
			return code >= 500 || code == http.StatusTooManyRequests
		}),
	}
	if limiter != nil {
		opts = append(opts, retry.WithLimiter(limiter))
	}
	return retry.Do(ctx, httpClient, req, opts...)
}
