package lib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

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
}

func newHTTPClient() *http.Client {
	return httpx.NewHTTPClient(httpx.Options{})
}

// NewClient authenticates via email/password using the legacy /api/sessions
// endpoint and returns a ready-to-use client.
//
// Deprecated: Skylight no longer supports the legacy sessions endpoint. Use
// NewClientWithRefreshToken instead, with a refresh token obtained via
// LoginHeadless or the "skylight auth login --save" command.
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
// returns a ready-to-use client. The refresh token rotates on each use; the
// caller should persist Client.RefreshToken after a successful call.
func NewClientWithRefreshToken(refreshToken, fingerprint string, opts ...ClientOption) (*Client, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}
	if fingerprint == "" {
		return nil, errors.New("device fingerprint is required")
	}

	tok, err := RefreshOAuthToken(refreshToken, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	cfg := defaultClientConfig()
	for _, o := range opts {
		o(&cfg)
	}

	return &Client{
		APIToken:     tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		authCache:    "Bearer " + tok.AccessToken,
		HTTPClient:   cfg.httpClient,
		baseURL:      cfg.baseURL,
		apiVersion:   cfg.apiVersion,
		logger:       cfg.logger,
		limiter:      cfg.limiter,
		retry:        cfg.retry,
	}, nil
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

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", c.authCache)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("skylight-api-version", c.apiVersion)
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

func (c *Client) do(req *http.Request) (*http.Response, error) {
	c.setHeaders(req)

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
