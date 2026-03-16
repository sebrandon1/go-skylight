package lib

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

var (
	// SkylightURL is the default base URL for all API calls.
	// Override via WithBaseURL or by assigning directly in tests.
	SkylightURL = "https://app.ourskylight.com/api"
)

// Client is a Skylight API client. Create one with NewClient or
// NewClientWithToken; use With* functional options for advanced configuration.
type Client struct {
	UserID     string
	APIToken   string
	HTTPClient *http.Client
	authCache  string

	baseURL string
	logger  *slog.Logger
	limiter *rate.Limiter
	retry   retryConfig
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

// NewClient authenticates via email/password and returns a ready-to-use client.
// Optional ClientOption values customize HTTP client, logging, retry, etc.
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
	c.authCache = "Basic " + base64.StdEncoding.EncodeToString([]byte(c.UserID+":"+c.APIToken))

	return c, nil
}

// NewClientWithToken creates a client from a pre-existing user ID and API token,
// skipping the login round-trip. Optional ClientOption values apply as usual.
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
		authCache:  "Basic " + base64.StdEncoding.EncodeToString([]byte(userID+":"+token)),
		HTTPClient: cfg.httpClient,
		baseURL:    cfg.baseURL,
		logger:     cfg.logger,
		limiter:    cfg.limiter,
		retry:      cfg.retry,
	}, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", c.authCache)
	req.Header.Set("Content-Type", "application/json")
}

// effectiveURL returns the client's base URL, falling back to the package-level
// SkylightURL so existing code that swaps SkylightURL in tests still works.
func (c *Client) effectiveURL() string {
	if c.baseURL != "" {
		return c.baseURL
	}
	return SkylightURL
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	c.setHeaders(req)

	if c.logger != nil {
		c.logger.Debug("skylight request",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.String("auth", "Basic [REDACTED]"),
		)
	}

	var lim limiterIface
	if c.limiter != nil {
		lim = c.limiter
	}
	resp, err := doWithRetry(context.Background(), c.HTTPClient, lim, c.retry, req)
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

func (c *Client) get(req *http.Request, v any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if err := checkStatus(resp, body); err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func (c *Client) post(req *http.Request, v any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return checkStatus(resp, body)
	}
	if v != nil {
		return json.Unmarshal(body, v)
	}
	return nil
}

func (c *Client) put(req *http.Request, v any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return checkStatus(resp, body)
	}
	if v != nil {
		return json.Unmarshal(body, v)
	}
	return nil
}

func (c *Client) patch(req *http.Request, v any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return checkStatus(resp, body)
	}
	if v != nil {
		return json.Unmarshal(body, v)
	}
	return nil
}

func (c *Client) doDelete(req *http.Request) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return checkStatus(resp, body)
}

func addQueryParams(req *http.Request, params map[string]string) {
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
}

//nolint:unparam
func newRequest(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func newRequestWithBody(method, url string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	return http.NewRequest(method, url, bodyReader)
}
