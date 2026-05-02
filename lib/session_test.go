package lib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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
