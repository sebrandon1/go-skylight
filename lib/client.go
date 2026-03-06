package lib

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

var (
	SkylightURL = "https://app.ourskylight.com/api"
)

type Client struct {
	UserID     string
	APIToken   string
	HTTPClient *http.Client
}

// NewClient authenticates via email/password and returns an authenticated client.
func NewClient(email, password string) (*Client, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	c := &Client{
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}

	session, err := c.Login(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	c.UserID = session.UserID
	c.APIToken = session.APIToken

	return c, nil
}

// NewClientWithToken creates a client using a pre-existing userId and token.
func NewClientWithToken(userID, token string) (*Client, error) {
	if userID == "" || token == "" {
		return nil, errors.New("user ID and token are required")
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &Client{
		UserID:   userID,
		APIToken: token,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}, nil
}

func (c *Client) authHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(c.UserID+":"+c.APIToken))
}

func (c *Client) get(req *http.Request, v any) error {
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return decodeJSON(resp.Body, v)
}

func (c *Client) post(req *http.Request, v any) error {
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	if v != nil {
		return decodeJSON(resp.Body, v)
	}

	return nil
}

func (c *Client) put(req *http.Request, v any) error {
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		return decodeJSON(resp.Body, v)
	}

	return nil
}

func (c *Client) patch(req *http.Request, v any) error {
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		return decodeJSON(resp.Body, v)
	}

	return nil
}

func (c *Client) doDelete(req *http.Request) error {
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func decodeJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
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
