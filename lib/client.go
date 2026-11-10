package lib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/sebrandon1/go-toolbox/httpx"
	"golang.org/x/time/rate"
)

// DateFormat is the standard date layout used across the Skylight API (YYYY-MM-DD).
const DateFormat = "2006-01-02"

const (
	maxResponseBytes  = httpx.DefaultMaxResponseBytes
	maxErrorBodyBytes = httpx.DefaultMaxErrorBodyBytes
)

// readLimited prevents unbounded memory growth from misbehaving servers.
func readLimited(r io.Reader, limit int64) ([]byte, error) { //nolint:unparam // general helper used by tests with custom limits
	return httpx.ReadLimited(r, limit)
}

// readErrorBody reads up to maxErrorBodyBytes from r, discarding any excess.
func readErrorBody(r io.Reader) string {
	return httpx.ReadErrorBody(r, maxErrorBodyBytes)
}

var (
	// SkylightURL is the default base URL for all API calls.
	// Override via WithBaseURL or by assigning directly in tests.
	SkylightURL = "https://app.ourskylight.com/api"
)

// Client is a Skylight API client. Create one with NewClient,
// NewClientWithToken, or NewClientWithRefreshToken; use With* functional
// options for advanced configuration.
type Client struct {
	UserID       string
	APIToken     string // Bearer access token (used for Authorization header).
	RefreshToken string // OAuth2 refresh token; rotates on each use.
	HTTPClient   *http.Client
	authCache    string

	baseURL    string
	apiVersion string
	logger     *slog.Logger
	limiter    *rate.Limiter
	retry      retryConfig

	// authMu guards APIToken, RefreshToken, authCache, authGeneration and
	// refreshFailure once the client is in use.
	authMu         sync.RWMutex
	authGeneration uint64
	refreshFailure error

	// refreshMu lets one request at a time exchange the refresh token, which
	// is single-use.
	refreshMu      sync.Mutex
	fingerprint    string
	tokenExchange  TokenExchangeFunc
	onTokenRefresh func(refreshToken string) error
}

func newHTTPClient() *http.Client {
	return httpx.NewHTTPClient(httpx.Options{})
}

// NewClient authenticates via email/password using the legacy /api/sessions
// endpoint and returns a ready-to-use client.
//
// Deprecated: Skylight no longer supports the legacy sessions endpoint. Use
// NewClientWithRefreshToken instead, with a refresh token obtained via
// LoginHeadless or the "skylight login --save" command.
func NewClient(email, password string, opts ...ClientOption) (*Client, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	cfg := defaultClientConfig()
	for _, o := range opts {
		o(&cfg)
	}

	c := &Client{
		HTTPClient: cfg.httpClient,
		baseURL:    cfg.baseURL,
		apiVersion: cfg.apiVersion,
		logger:     cfg.logger,
		limiter:    cfg.limiter,
		retry:      cfg.retry,
	}

	session, err := c.Login(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	c.UserID = session.UserID
	c.APIToken = session.APIToken
	c.authCache = "Bearer " + c.APIToken

	return c, nil
}

// NewClientWithRefreshToken authenticates using an OAuth2 refresh token and
// returns a ready-to-use client. The refresh token rotates on each use and the
// client renews it again on a 401, so a caller that keeps tokens across restarts
// should save each one with WithTokenRefreshHandler.
func NewClientWithRefreshToken(refreshToken, fingerprint string, opts ...ClientOption) (*Client, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}
	if fingerprint == "" {
		return nil, errors.New("device fingerprint is required")
	}

	cfg := defaultClientConfig()
	for _, o := range opts {
		o(&cfg)
	}

	exchange := cfg.tokenExchange
	if exchange == nil {
		exchange = RefreshOAuthToken
	}
	tok, err := exchange(refreshToken, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	c := &Client{
		APIToken:       tok.AccessToken,
		RefreshToken:   tok.RefreshToken,
		authCache:      "Bearer " + tok.AccessToken,
		HTTPClient:     cfg.httpClient,
		baseURL:        cfg.baseURL,
		apiVersion:     cfg.apiVersion,
		logger:         cfg.logger,
		limiter:        cfg.limiter,
		retry:          cfg.retry,
		fingerprint:    fingerprint,
		tokenExchange:  exchange,
		onTokenRefresh: cfg.onTokenRefresh,
	}
	c.notifyTokenRefresh(tok.RefreshToken)
	return c, nil
}

// NewClientWithToken creates a client from a pre-existing user ID and Bearer
// access token, skipping the login round-trip. Optional ClientOption values
// apply as usual.
func NewClientWithToken(userID, token string, opts ...ClientOption) (*Client, error) {
	if userID == "" || token == "" {
		return nil, errors.New("user ID and token are required")
	}

	cfg := defaultClientConfig()
	for _, o := range opts {
		o(&cfg)
	}

	return &Client{
		UserID:     userID,
		APIToken:   token,
		authCache:  "Bearer " + token,
		HTTPClient: cfg.httpClient,
		baseURL:    cfg.baseURL,
		apiVersion: cfg.apiVersion,
		logger:     cfg.logger,
		limiter:    cfg.limiter,
		retry:      cfg.retry,
	}, nil
}

// setHeaders applies the current credentials and returns their generation.
func (c *Client) setHeaders(req *http.Request) uint64 {
	c.authMu.RLock()
	auth, generation := c.authCache, c.authGeneration
	c.authMu.RUnlock()

	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("skylight-api-version", c.apiVersion)
	return generation
}

// effectiveURL returns the client's base URL, falling back to the package-level
// SkylightURL so existing code that swaps SkylightURL in tests still works.
func (c *Client) effectiveURL() string {
	if c.baseURL != "" {
		return c.baseURL
	}
	return SkylightURL
}

// pathSeg escapes a single path segment (IDs from CLI/import) before URL
// interpolation so values containing '/', '..', or percent sequences cannot
// rewrite the request path.
func pathSeg(s string) string {
	return httpx.PathSeg(s)
}

// do sends req. A client built from a refresh token renews its access token on
// a 401 and replays the request once.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	generation := c.setHeaders(req)
	resp, err := c.send(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized || c.tokenExchange == nil {
		return resp, err
	}

	replay := req.Clone(req.Context())
	// doWithRetry leaves GetBody set on any request that has a body.
	if req.GetBody != nil {
		if replay.Body, err = req.GetBody(); err != nil {
			return resp, nil //nolint:nilerr // the original 401 is the useful result
		}
	}
	resp.Body.Close()

	if err := c.refreshAccessToken(generation); err != nil {
		return nil, fmt.Errorf("refreshing access token: %w", err)
	}
	c.setHeaders(replay)
	return c.send(replay)
}

// refreshAccessToken exchanges the refresh token unless another request already
// did so after generation, the credentials the caller's 401 was for.
func (c *Client) refreshAccessToken(generation uint64) error {
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	c.authMu.RLock()
	current, token := c.authGeneration, c.RefreshToken
	c.authMu.RUnlock()
	if current != generation {
		return nil
	}

	tok, err := c.tokenExchange(token, c.fingerprint)
	if err != nil {
		if IsAuthError(err) {
			c.authMu.Lock()
			c.refreshFailure = err
			c.authMu.Unlock()
		}
		return err
	}

	c.authMu.Lock()
	c.APIToken = tok.AccessToken
	c.authCache = "Bearer " + tok.AccessToken
	if tok.RefreshToken != "" {
		c.RefreshToken = tok.RefreshToken
	}
	c.authGeneration++
	c.refreshFailure = nil
	c.authMu.Unlock()

	c.notifyTokenRefresh(tok.RefreshToken)
	return nil
}

// notifyTokenRefresh hands a newly issued refresh token to the caller's handler.
func (c *Client) notifyTokenRefresh(refreshToken string) {
	if c.onTokenRefresh == nil || refreshToken == "" {
		return
	}
	if err := c.onTokenRefresh(refreshToken); err != nil && c.logger != nil {
		c.logger.Warn("token refresh handler failed", slog.String("error", err.Error()))
	}
}

// RefreshFailure returns the error from the most recent token refresh the
// server rejected, or nil if the last refresh succeeded. Once it is non-nil the
// client cannot recover without new credentials, so long-running callers should
// stop.
func (c *Client) RefreshFailure() error {
	c.authMu.RLock()
	defer c.authMu.RUnlock()
	return c.refreshFailure
}

func (c *Client) send(req *http.Request) (*http.Response, error) {

	if c.logger != nil {
		c.logger.Debug("skylight request",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.String("auth", "Bearer [REDACTED]"),
		)
	}

	var lim limiterIface
	if c.limiter != nil {
		lim = c.limiter
	}
	resp, err := doWithRetry(req.Context(), c.HTTPClient, lim, c.retry, req)
	if err != nil {
		return nil, err
	}

	if c.logger != nil {
		c.logger.Debug("skylight response",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.Int("status", resp.StatusCode),
		)
	}

	return resp, nil
}

// handleResponse closes resp.Body, routes non-2xx statuses through checkStatus,
// and unmarshals into v when v is non-nil. 204 No Content returns immediately
// without reading the body.
func handleResponse(resp *http.Response, v any) error {
	return httpx.HandleResponse(resp, v, maxResponseBytes)
}

func (c *Client) get(req *http.Request, v any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	return handleResponse(resp, v)
}

func (c *Client) post(req *http.Request, v any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	return handleResponse(resp, v)
}

func (c *Client) put(req *http.Request, v any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	return handleResponse(resp, v)
}

func (c *Client) patch(req *http.Request, v any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	return handleResponse(resp, v)
}

func (c *Client) doDelete(req *http.Request) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

func addQueryParams(req *http.Request, params map[string]string) {
	httpx.AddQueryParams(req, params)
}

func newRequest(ctx context.Context, method, url string) (*http.Request, error) {
	return httpx.NewRequest(ctx, method, url)
}

func newRequestWithBody(ctx context.Context, method, url string, body any) (*http.Request, error) {
	return httpx.NewRequestWithBody(ctx, method, url, body)
}
