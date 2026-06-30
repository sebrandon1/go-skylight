package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	browserUA     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	browserAccept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
)

func newBrowserRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept", browserAccept)
	return req, nil
}

const (
	skylightClientID    = "skylight-mobile"
	skylightScope       = "everything"
	skylightRedirectURI = "https://ourskylight.com/welcome"
)

var (
	// OAuthURL is the Skylight OAuth2 token endpoint. Override in tests.
	OAuthURL = "https://app.ourskylight.com/oauth/token"

	// AuthSessionURL is the Skylight Rails session login endpoint. Override in tests.
	AuthSessionURL = "https://app.ourskylight.com/auth/session"

	// OAuthAuthorizeURL is the Skylight OAuth2 authorize endpoint. Override in tests.
	OAuthAuthorizeURL = "https://app.ourskylight.com/oauth/authorize"
)

// RefreshOAuthToken exchanges a refresh token for a new access token using
// the Skylight OAuth2 refresh token grant. The refresh token rotates on each
// use; callers must persist the new RefreshToken from the response.
func RefreshOAuthToken(refreshToken, fingerprint string) (*OAuthTokenResponse, error) {
	return postOAuthToken(url.Values{
		"grant_type":                             {"refresh_token"},
		"refresh_token":                          {refreshToken},
		"client_id":                              {skylightClientID},
		"skylight_api_client_device_fingerprint": {fingerprint},
	})
}

// LoginHeadless performs a headless OAuth2 authorization-code login using
// email and password. It returns a token response including both an access
// token and a refresh token. The fingerprint is a stable UUID that identifies
// this client device.
func LoginHeadless(email, password, fingerprint string) (*OAuthTokenResponse, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	hc := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Step 1: GET login page to extract CSRF token.
	csrfToken, err := fetchCSRFToken(hc)
	if err != nil {
		return nil, fmt.Errorf("fetching CSRF token: %w", err)
	}

	// Step 2: POST credentials to create a Rails session.
	if err := postSession(hc, email, password, csrfToken); err != nil {
		return nil, fmt.Errorf("logging in: %w", err)
	}

	// Step 3: GET OAuth authorize endpoint to receive the auth code.
	code, err := fetchAuthCode(hc, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("fetching auth code: %w", err)
	}

	// Step 4: Exchange auth code for tokens.
	return exchangeAuthCode(code, fingerprint)
}

// fetchCSRFToken GETs the login page and extracts the Rails authenticity_token.
func fetchCSRFToken(hc *http.Client) (string, error) {
	req, err := newBrowserRequest("GET", AuthSessionURL+"/new", nil)
	if err != nil {
		return "", err
	}

	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Match <input ... name="authenticity_token" ... value="..." />
	re := regexp.MustCompile(`name="authenticity_token"[^>]*value="([^"]+)"`)
	if m := re.FindSubmatch(body); len(m) > 1 {
		return string(m[1]), nil
	}
	// Also try value before name ordering.
	re2 := regexp.MustCompile(`value="([^"]+)"[^>]*name="authenticity_token"`)
	if m := re2.FindSubmatch(body); len(m) > 1 {
		return string(m[1]), nil
	}
	return "", fmt.Errorf("authenticity_token not found in login page")
}

// postSession POSTs credentials to the Rails session endpoint.
func postSession(hc *http.Client, email, password, csrfToken string) error {
	form := url.Values{}
	form.Set("authenticity_token", csrfToken)
	form.Set("email", email)
	form.Set("password", password)

	req, err := newBrowserRequest("POST", AuthSessionURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if u, e := url.Parse(AuthSessionURL); e == nil {
		req.Header.Set("Origin", u.Scheme+"://"+u.Host)
	}
	req.Header.Set("Referer", AuthSessionURL+"/new")

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login returned status %d: %s", resp.StatusCode, string(body))
	}

	// A 302 back to the login page (rather than to a dashboard/welcome
	// page) means the login was rejected; Rails still responds with a
	// redirect in that case, so status code alone can't distinguish
	// success from failure.
	if resp.StatusCode == http.StatusFound && isLoginPageURL(resp.Header.Get("Location")) {
		return errLoginRejected
	}
	return nil
}

// errLoginRejected is returned whenever Skylight redirects back to the login
// page instead of completing authentication. Wrong credentials are the most
// common cause, but a pending device/2FA verification can trigger the same
// redirect, so the message stays deliberately non-committal.
var errLoginRejected = fmt.Errorf("login rejected: check email/password, or complete any pending device/2FA verification in a browser")

// isLoginPageURL reports whether rawURL points back at the login form,
// which Skylight uses to signal a failed login attempt.
func isLoginPageURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	authPath, err := url.Parse(AuthSessionURL)
	if err != nil {
		return false
	}
	return u.Path == authPath.Path+"/new"
}

// fetchAuthCode GETs the OAuth authorize endpoint and extracts the auth code
// from the redirect Location header.
func fetchAuthCode(hc *http.Client, fingerprint string) (string, error) {
	params := url.Values{}
	params.Set("client_id", skylightClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", skylightRedirectURI)
	params.Set("scope", skylightScope)
	params.Set("skylight_api_client_device_fingerprint", fingerprint)

	authorizeURL := OAuthAuthorizeURL + "?" + params.Encode()
	req, err := newBrowserRequest("GET", authorizeURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("oauth authorize returned status %d with no Location: %s", resp.StatusCode, string(body))
	}

	u, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parsing redirect URL: %w", err)
	}

	code := u.Query().Get("code")
	if code == "" {
		if isLoginPageURL(location) {
			return "", errLoginRejected
		}
		return "", fmt.Errorf("no code in redirect URL: %s", location)
	}
	return code, nil
}

// exchangeAuthCode exchanges an OAuth2 authorization code for tokens.
func exchangeAuthCode(code, fingerprint string) (*OAuthTokenResponse, error) {
	return postOAuthToken(url.Values{
		"grant_type":                             {"authorization_code"},
		"code":                                   {code},
		"client_id":                              {skylightClientID},
		"redirect_uri":                           {skylightRedirectURI},
		"scope":                                  {skylightScope},
		"skylight_api_client_device_fingerprint": {fingerprint},
	})
}

// postOAuthToken posts form values to the OAuth token endpoint and returns the response.
func postOAuthToken(data url.Values) (*OAuthTokenResponse, error) {
	//nolint:gosec // OAuthURL is a package-level var, not user input; swappable in tests.
	resp, err := http.PostForm(OAuthURL, data)
	if err != nil {
		return nil, fmt.Errorf("oauth token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading oauth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth token request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var token OAuthTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("parsing oauth response: %w", err)
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("oauth response missing access_token: %s", string(body))
	}

	return &token, nil
}

// Login authenticates with the Skylight API using email and password via the
// legacy /api/sessions endpoint.
//
// Deprecated: Skylight no longer supports this endpoint. Use LoginHeadless or
// RefreshOAuthToken instead.
func (c *Client) Login(email, password string) (*Session, error) {
	reqBody := SessionRequest{
		Email:    email,
		Password: password,
	}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/sessions", c.effectiveURL()), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	var resp sessionResponse
	if err := c.post(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &Session{
		UserID:   resp.Data.ID,
		APIToken: resp.Data.Attributes.Token,
	}, nil
}
