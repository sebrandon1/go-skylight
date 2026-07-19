package lib

import (
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
	_, err := fetchAuthCode(hc, "fp1")
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
	_, err := fetchAuthCode(hc, "fp1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no code in redirect URL") {
		t.Errorf("expected error to contain 'no code in redirect URL', got: %v", err)
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
