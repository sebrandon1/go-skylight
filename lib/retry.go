package lib

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net/http"
	"time"
)

// doWithRetry executes req via httpClient, retrying on transient errors and
// 5xx / 429 responses according to cfg. The rate limiter is consulted before
// each attempt. Callers must not reuse req after this call.
//
// A buffered copy of the request body is kept so it can be replayed on retry.
func doWithRetry(ctx context.Context, httpClient *http.Client, limiter limiterIface, cfg retryConfig, req *http.Request) (*http.Response, error) {
	bodyBytes, err := bufferBody(req)
	if err != nil {
		return nil, err
	}

	maxAttempts := cfg.maxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := backoffDelay(cfg.baseDelay, cfg.maxDelay, attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		if limiter != nil {
			if err := limiter.Wait(ctx); err != nil {
				return nil, err
			}
		}

		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.ContentLength = int64(len(bodyBytes))
		}

		resp, doErr := httpClient.Do(req)
		if doErr != nil {
			lastErr = &NetworkError{Cause: doErr}
			continue
		}

		if shouldRetry(resp.StatusCode) {
			lastErr = drainAndError(ctx, resp)
			if lastErr == ctx.Err() {
				return nil, lastErr
			}
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

// bufferBody reads and buffers req.Body so it can be replayed on retries.
func bufferBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}
	b, err := io.ReadAll(req.Body)
	req.Body.Close() //nolint:errcheck
	if err != nil {
		return nil, &NetworkError{Cause: err}
	}
	return b, nil
}

// shouldRetry reports whether a status code warrants a retry.
func shouldRetry(status int) bool {
	return status >= 500 || status == http.StatusTooManyRequests
}

// drainAndError drains resp.Body, optionally waits for Retry-After, then
// returns the appropriate error. It returns ctx.Err() if the context is
// canceled while waiting.
func drainAndError(ctx context.Context, resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusTooManyRequests {
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	wait := parseRetryAfter(resp.Header.Get("Retry-After"))
	if wait > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return &RateLimitError{RetryAfter: wait}
}

// limiterIface is a subset of *rate.Limiter so the caller can pass nil without
// importing golang.org/x/time/rate in every file.
type limiterIface interface {
	Wait(ctx context.Context) error
}

// backoffDelay returns the wait before attempt n (1-based) with full jitter.
// Formula: min(maxDelay, base * 2^(n-1)) with a random value in [exp/2, exp].
func backoffDelay(base, max time.Duration, attempt int) time.Duration {
	exp := base
	for i := 1; i < attempt; i++ {
		exp *= 2
		if exp > max {
			exp = max
			break
		}
	}
	if exp > max {
		exp = max
	}
	// Jitter: crypto/rand-based value in [exp/2, exp] to avoid correlated retries.
	// exp is always positive here (capped to maxDelay), so the conversions are safe.
	half := uint64(exp / 2) //nolint:gosec
	if half == 0 {
		return exp
	}
	var b [8]byte
	// crypto/rand.Read is guaranteed to return len(b) bytes on supported
	// platforms; an error here would indicate a broken OS entropy source.
	if _, err := rand.Read(b[:]); err != nil {
		return exp
	}
	n := binary.LittleEndian.Uint64(b[:]) % half
	return time.Duration(n) + exp/2 //nolint:gosec
}
