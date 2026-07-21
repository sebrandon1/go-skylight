package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewUUID_Format(t *testing.T) {
	id := newUUID()
	if !uuidRe.MatchString(id) {
		t.Errorf("newUUID() = %q, want UUID v4 format", id)
	}
}

func TestNewUUID_Unique(t *testing.T) {
	a, b := newUUID(), newUUID()
	if a == b {
		t.Errorf("newUUID() returned identical values: %q", a)
	}
}

// newHeadlessLoginServer stands up a mock for the full LoginHeadless flow:
// CSRF fetch, session POST, OAuth authorize redirect, and token exchange.
func newHeadlessLoginServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/session/new", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<input name="authenticity_token" value="csrf123">`)
	})
	mux.HandleFunc("/auth/session", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/welcome", http.StatusFound)
	})
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://ourskylight.com/welcome?code=auth-code-123", http.StatusFound)
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"access1","refresh_token":"refresh1","expires_in":3600}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func withHeadlessLoginServer(t *testing.T) {
	t.Helper()
	srv := newHeadlessLoginServer(t)

	origAuth, origAuthorize, origOAuth := lib.AuthSessionURL, lib.OAuthAuthorizeURL, lib.OAuthURL
	lib.AuthSessionURL = srv.URL + "/auth/session"
	lib.OAuthAuthorizeURL = srv.URL + "/oauth/authorize"
	lib.OAuthURL = srv.URL + "/oauth/token"
	t.Cleanup(func() {
		lib.AuthSessionURL, lib.OAuthAuthorizeURL, lib.OAuthURL = origAuth, origAuthorize, origOAuth
	})
}

func TestLoginCmd_MissingCredentials(t *testing.T) {
	origEmail, origPassword := email, password
	email, password = "", ""
	t.Cleanup(func() { email, password = origEmail, origPassword })

	err := loginCmd.RunE(loginCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
	if !strings.Contains(err.Error(), "are required for login") {
		t.Errorf("expected 'are required for login' in error, got: %v", err)
	}
}

func TestLoginCmd_Success(t *testing.T) {
	withHeadlessLoginServer(t)

	origEmail, origPassword, origFingerprint, origSave := email, password, deviceFingerprint, saveCredentials
	email, password = "u@example.com", "pw"
	deviceFingerprint = "fp-fixed"
	saveCredentials = false
	t.Cleanup(func() {
		email, password, deviceFingerprint, saveCredentials = origEmail, origPassword, origFingerprint, origSave
	})

	out := captureStdout(func() {
		if err := loginCmd.RunE(loginCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Login successful!") {
		t.Errorf("expected success message, got: %s", out)
	}
	if !strings.Contains(out, "access1") || !strings.Contains(out, "refresh1") {
		t.Errorf("expected tokens in output, got: %s", out)
	}
}

func TestLoginCmd_GeneratesFingerprintWhenMissing(t *testing.T) {
	withHeadlessLoginServer(t)

	origEmail, origPassword, origFingerprint, origSave := email, password, deviceFingerprint, saveCredentials
	email, password = "u@example.com", "pw"
	deviceFingerprint = ""
	saveCredentials = false
	t.Cleanup(func() {
		email, password, deviceFingerprint, saveCredentials = origEmail, origPassword, origFingerprint, origSave
	})

	out := captureStdout(func() {
		if err := loginCmd.RunE(loginCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Fingerprint:") {
		t.Errorf("expected a generated fingerprint in output, got: %s", out)
	}
}

func TestLoginCmd_SavesCredentials(t *testing.T) {
	withHeadlessLoginServer(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	origEmail, origPassword, origFingerprint, origSave, origConfigPath, origFrameID :=
		email, password, deviceFingerprint, saveCredentials, configPath, frameID
	email, password = "u@example.com", "pw"
	deviceFingerprint = "fp-fixed"
	saveCredentials = true
	configPath = path
	frameID = "f1"
	t.Cleanup(func() {
		email, password, deviceFingerprint, saveCredentials, configPath, frameID =
			origEmail, origPassword, origFingerprint, origSave, origConfigPath, origFrameID
	})

	out := captureStdout(func() {
		if err := loginCmd.RunE(loginCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Credentials saved to "+path) {
		t.Errorf("expected save confirmation, got: %s", out)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved config: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "SKYLIGHT_REFRESH_TOKEN=refresh1") {
		t.Errorf("expected refresh token persisted, got: %s", content)
	}
	if !strings.Contains(content, "SKYLIGHT_FRAME_ID=f1") {
		t.Errorf("expected frame id persisted, got: %s", content)
	}
}

func TestLoginCmd_LoginFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	origAuth := lib.AuthSessionURL
	lib.AuthSessionURL = srv.URL + "/auth/session"
	t.Cleanup(func() { lib.AuthSessionURL = origAuth })

	origEmail, origPassword := email, password
	email, password = "u@example.com", "pw"
	t.Cleanup(func() { email, password = origEmail, origPassword })

	err := loginCmd.RunE(loginCmd, nil)
	if err == nil {
		t.Fatal("expected error for login failure, got nil")
	}
	if !strings.Contains(err.Error(), "logging in") {
		t.Errorf("expected 'logging in' in error, got: %v", err)
	}
}

func TestLoginCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "login")
}
