package lib

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLogin_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"uid1","type":"authenticated_user","attributes":{"token":"tok1"}}}`)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	sess, err := client.Login("user@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.UserID != "uid1" {
		t.Errorf("UserID: want uid1, got %s", sess.UserID)
	}
	if sess.APIToken != "tok1" {
		t.Errorf("APIToken: want tok1, got %s", sess.APIToken)
	}
}

func TestLogin_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"invalid credentials"}`)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.Login("bad@example.com", "wrong")
	if err == nil {
		t.Error("expected error for unauthorized response, got nil")
	}
}

func TestLogin_BadURL(t *testing.T) {
	old := SkylightURL
	SkylightURL = "://bad"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.Login("u@example.com", "pw")
	if err == nil {
		t.Error("expected error for bad URL, got nil")
	}
}

func TestRefreshOAuthToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: want POST, got %s", r.Method)
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if r.FormValue("grant_type") != "refresh_token" {
			t.Errorf("grant_type: want refresh_token, got %s", r.FormValue("grant_type"))
		}
		if r.FormValue("refresh_token") != "myrefresh" {
			t.Errorf("refresh_token: want myrefresh, got %s", r.FormValue("refresh_token"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"newaccess","refresh_token":"newrefresh","expires_in":3600}`)
	}))
	defer srv.Close()

	old := OAuthURL
	OAuthURL = srv.URL
	defer func() { OAuthURL = old }()

	tok, err := RefreshOAuthToken("myrefresh", "fp1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "newaccess" {
		t.Errorf("AccessToken: want newaccess, got %s", tok.AccessToken)
	}
	if tok.RefreshToken != "newrefresh" {
		t.Errorf("RefreshToken: want newrefresh, got %s", tok.RefreshToken)
	}
}

func TestRefreshOAuthToken_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer srv.Close()

	old := OAuthURL
	OAuthURL = srv.URL
	defer func() { OAuthURL = old }()

	_, err := RefreshOAuthToken("badtoken", "fp1")
	if err == nil {
		t.Error("expected error for 401 response, got nil")
	}
}

func TestRefreshOAuthToken_MissingAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"refresh_token":"r1"}`)
	}))
	defer srv.Close()

	old := OAuthURL
	OAuthURL = srv.URL
	defer func() { OAuthURL = old }()

	_, err := RefreshOAuthToken("tok", "fp1")
	if err == nil {
		t.Error("expected error for missing access_token, got nil")
	}
}

func TestPostOAuthToken_Timeout(t *testing.T) {
	unblock := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until the test signals done or the client disconnects.
		select {
		case <-unblock:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() {
		close(unblock)
		srv.Close()
	})

	old := OAuthURL
	OAuthURL = srv.URL
	defer func() { OAuthURL = old }()

	oldTimeout := oauthHTTPTimeout
	oauthHTTPTimeout = 50 * time.Millisecond
	defer func() { oauthHTTPTimeout = oldTimeout }()

	_, err := RefreshOAuthToken("tok", "fp1")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "deadline exceeded") && !strings.Contains(err.Error(), "context deadline") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected a timeout error, got: %v", err)
	}
}

func TestLoginHeadless_CSRFFetchFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body>no token here</body></html>`)
	}))
	defer srv.Close()

	oldAuth := AuthSessionURL
	AuthSessionURL = srv.URL + "/auth/session"
	defer func() { AuthSessionURL = oldAuth }()

	_, err := LoginHeadless("u@example.com", "pw", "fp1")
	if err == nil {
		t.Error("expected error when CSRF token not found, got nil")
	}
}

func TestLoginHeadless_InvalidCredentials(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/session/new", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<input name="authenticity_token" value="csrf123">`)
	})
	mux.HandleFunc("/auth/session", func(w http.ResponseWriter, r *http.Request) {
		// Rails redirects back to the login page on a rejected login.
		http.Redirect(w, r, "/auth/session/new", http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldAuth := AuthSessionURL
	AuthSessionURL = srv.URL + "/auth/session"
	defer func() { AuthSessionURL = oldAuth }()

	_, err := LoginHeadless("u@example.com", "wrongpw", "fp1")
	if err == nil {
		t.Fatal("expected error for rejected credentials, got nil")
	}
	if !strings.Contains(err.Error(), "login rejected") {
		t.Errorf("expected error to mention rejected login, got: %v", err)
	}
}

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := generatePKCE()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verifier == "" {
		t.Error("verifier should not be empty")
	}
	if challenge == "" {
		t.Error("challenge should not be empty")
	}
	h := sha256.Sum256([]byte(verifier))
	want := base64.RawURLEncoding.EncodeToString(h[:])
	if challenge != want {
		t.Errorf("challenge does not match S256(verifier): got %s, want %s", challenge, want)
	}
	// Two calls must produce different verifiers.
	verifier2, _, err := generatePKCE()
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if verifier == verifier2 {
		t.Error("generatePKCE produced the same verifier twice")
	}
}

func TestFetchAuthCode_NoLocation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<html><body>some page</body></html>`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldAuthorize := OAuthAuthorizeURL
	OAuthAuthorizeURL = srv.URL + "/oauth/authorize"
	defer func() { OAuthAuthorizeURL = oldAuthorize }()

	hc := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	_, err := fetchAuthCode(hc, "fp1", "testchallenge")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no Location") {
		t.Errorf("expected error to contain 'no Location', got: %v", err)
	}
}

func TestFetchAuthCode_NoCodeInRedirect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://example.com/welcome")
		w.WriteHeader(http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldAuthorize := OAuthAuthorizeURL
	OAuthAuthorizeURL = srv.URL + "/oauth/authorize"
	defer func() { OAuthAuthorizeURL = oldAuthorize }()

	hc := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	_, err := fetchAuthCode(hc, "fp1", "testchallenge")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no code in redirect URL") {
		t.Errorf("expected error to contain 'no code in redirect URL', got: %v", err)
	}
}

func TestExchangeAuthCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse form: %v", err)
			}
			if r.FormValue("grant_type") != "authorization_code" {
				t.Errorf("grant_type: want authorization_code, got %s", r.FormValue("grant_type"))
			}
			if r.FormValue("code") != "mycode" {
				t.Errorf("code: want mycode, got %s", r.FormValue("code"))
			}
			if r.FormValue("code_verifier") != "testverifier" {
				t.Errorf("code_verifier: want testverifier, got %s", r.FormValue("code_verifier"))
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"at1","refresh_token":"rt1","expires_in":3600}`)
		}))
		defer srv.Close()

		old := OAuthURL
		OAuthURL = srv.URL
		defer func() { OAuthURL = old }()

		tok, err := exchangeAuthCode("mycode", "fp1", "testverifier")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tok.AccessToken != "at1" {
			t.Errorf("AccessToken: want at1, got %s", tok.AccessToken)
		}
		if tok.RefreshToken != "rt1" {
			t.Errorf("RefreshToken: want rt1, got %s", tok.RefreshToken)
		}
	})

	t.Run("error response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"invalid_grant"}`)
		}))
		defer srv.Close()

		old := OAuthURL
		OAuthURL = srv.URL
		defer func() { OAuthURL = old }()

		_, err := exchangeAuthCode("badcode", "fp1", "testverifier")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestIsLoginPageURL(t *testing.T) {
	old := AuthSessionURL
	AuthSessionURL = "https://app.ourskylight.com/auth/session"
	defer func() { AuthSessionURL = old }()

	tests := []struct {
		name    string
		url     string
		authURL string
		want    bool
	}{
		{"login page", "https://app.ourskylight.com/auth/session/new", "", true},
		{"dashboard", "https://app.ourskylight.com/dashboard", "", false},
		{"empty string", "", "", false},
		{"oauth redirect with code", "https://ourskylight.com/welcome?code=abc123", "", false},
		// url.Parse error on rawURL (null byte in URL)
		{"invalid rawURL", "http://bad\x00url", "", false},
		// url.Parse error on AuthSessionURL
		{"invalid auth session URL", "https://app.ourskylight.com/auth/session/new", "http://bad\x00url", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.authURL != "" {
				prev := AuthSessionURL
				AuthSessionURL = tc.authURL
				defer func() { AuthSessionURL = prev }()
			}
			got := isLoginPageURL(tc.url)
			if got != tc.want {
				t.Errorf("isLoginPageURL(%q) = %v, want %v", tc.url, got, tc.want)
			}
		})
	}
}

func TestFetchCSRFToken_AlternateAttrOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<input value="altcsrf123" name="authenticity_token">`)
	}))
	defer srv.Close()

	old := AuthSessionURL
	AuthSessionURL = srv.URL + "/auth/session"
	defer func() { AuthSessionURL = old }()

	tok, err := fetchCSRFToken(&http.Client{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "altcsrf123" {
		t.Errorf("CSRF token: want altcsrf123, got %q", tok)
	}
}

func TestFetchCSRFToken_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srvURL := srv.URL
	srv.Close()

	old := AuthSessionURL
	AuthSessionURL = srvURL + "/auth/session"
	defer func() { AuthSessionURL = old }()

	_, err := fetchCSRFToken(&http.Client{})
	if err == nil {
		t.Error("expected error for network failure, got nil")
	}
}

func TestPostSession_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "server error")
	}))
	defer srv.Close()

	old := AuthSessionURL
	AuthSessionURL = srv.URL + "/auth/session"
	defer func() { AuthSessionURL = old }()

	err := postSession(&http.Client{}, "u@example.com", "pw", "csrf")
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error to mention status 500, got: %v", err)
	}
}

func TestNewBrowserRequest_BadMethod(t *testing.T) {
	_, err := newBrowserRequest("G\x00ET", "https://example.com", nil)
	if err == nil {
		t.Error("expected error for invalid method, got nil")
	}
}

func TestFetchAuthCode_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("code_challenge") != "testchallenge" {
			t.Errorf("code_challenge: want testchallenge, got %q", r.URL.Query().Get("code_challenge"))
		}
		if r.URL.Query().Get("code_challenge_method") != "S256" {
			t.Errorf("code_challenge_method: want S256, got %q", r.URL.Query().Get("code_challenge_method"))
		}
		w.Header().Set("Location", "https://ourskylight.com/welcome?code=myauthcode")
		w.WriteHeader(http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	old := OAuthAuthorizeURL
	OAuthAuthorizeURL = srv.URL + "/oauth/authorize"
	defer func() { OAuthAuthorizeURL = old }()

	hc := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	code, err := fetchAuthCode(hc, "fp1", "testchallenge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "myauthcode" {
		t.Errorf("code: want myauthcode, got %q", code)
	}
}

func TestPostSession_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srvURL := srv.URL
	srv.Close()

	old := AuthSessionURL
	AuthSessionURL = srvURL + "/auth/session"
	defer func() { AuthSessionURL = old }()

	err := postSession(&http.Client{}, "u@example.com", "pw", "csrf")
	if err == nil {
		t.Error("expected error for connection refused, got nil")
	}
}

func TestFetchAuthCode_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srvURL := srv.URL
	srv.Close()

	old := OAuthAuthorizeURL
	OAuthAuthorizeURL = srvURL + "/oauth/authorize"
	defer func() { OAuthAuthorizeURL = old }()

	_, err := fetchAuthCode(&http.Client{}, "fp1", "testchallenge")
	if err == nil {
		t.Error("expected error for connection refused, got nil")
	}
}

// startHeadlessMux starts a mock server handling steps 1–3 of LoginHeadless
// (CSRF fetch, session POST, OAuth authorize) and wires AuthSessionURL and
// OAuthAuthorizeURL to point at it. Cleanup is registered via t.Cleanup;
// the caller only needs to set OAuthURL for the token endpoint.
func startHeadlessMux(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/session/new", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<input name="authenticity_token" value="csrf123">`)
	})
	mux.HandleFunc("/auth/session", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://ourskylight.com/welcome?code=myauthcode")
		w.WriteHeader(http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	oldAuth, oldAuthorize := AuthSessionURL, OAuthAuthorizeURL
	AuthSessionURL = srv.URL + "/auth/session"
	OAuthAuthorizeURL = srv.URL + "/oauth/authorize"
	t.Cleanup(func() {
		srv.Close()
		AuthSessionURL = oldAuth
		OAuthAuthorizeURL = oldAuthorize
	})
}

func TestLoginHeadless_Success(t *testing.T) {
	startHeadlessMux(t)

	oauthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"at1","refresh_token":"rt1","expires_in":3600}`)
	}))
	defer oauthSrv.Close()

	old := OAuthURL
	OAuthURL = oauthSrv.URL
	defer func() { OAuthURL = old }()

	tok, err := LoginHeadless("u@example.com", "pw", "fp1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "at1" {
		t.Errorf("AccessToken: want at1, got %s", tok.AccessToken)
	}
	if tok.RefreshToken != "rt1" {
		t.Errorf("RefreshToken: want rt1, got %s", tok.RefreshToken)
	}
}

func TestLoginHeadless_TokenExchangeFailure(t *testing.T) {
	startHeadlessMux(t)

	oauthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer oauthSrv.Close()

	old := OAuthURL
	OAuthURL = oauthSrv.URL
	defer func() { OAuthURL = old }()

	_, err := LoginHeadless("u@example.com", "pw", "fp1")
	if err == nil {
		t.Fatal("expected error for token exchange failure, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to mention 401, got: %v", err)
	}
}

func TestLoginHeadless_NotLoggedInAtAuthorize(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/session/new", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<input name="authenticity_token" value="csrf123">`)
	})
	mux.HandleFunc("/auth/session", func(w http.ResponseWriter, r *http.Request) {
		// Devise-style re-render: 200 with the form, not a redirect, so
		// postSession can't detect the failure - it only surfaces once
		// the authorize step bounces back to the login page.
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<html>invalid email or password</html>`)
	})
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/auth/session/new", http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldAuth, oldAuthorize := AuthSessionURL, OAuthAuthorizeURL
	AuthSessionURL = srv.URL + "/auth/session"
	OAuthAuthorizeURL = srv.URL + "/oauth/authorize"
	defer func() {
		AuthSessionURL = oldAuth
		OAuthAuthorizeURL = oldAuthorize
	}()

	_, err := LoginHeadless("u@example.com", "pw", "fp1")
	if err == nil {
		t.Fatal("expected error when authorize redirects back to login, got nil")
	}
	if !strings.Contains(err.Error(), "login rejected") {
		t.Errorf("expected error to mention rejected login, got: %v", err)
	}
}
