package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
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

// TestLoginCmd_MissingCredentials_Crasher is invoked as a subprocess by
// TestLoginCmd_MissingCredentials to exercise loginCmd's os.Exit(1) path
// without terminating the real test binary.
func TestLoginCmd_MissingCredentials_Crasher(t *testing.T) {
	if os.Getenv("WANT_LOGIN_CRASH") != "1" {
		t.Skip("only runs as a subprocess of TestLoginCmd_MissingCredentials")
	}
	email, password = "", ""
	loginCmd.Run(loginCmd, nil)
}

func TestLoginCmd_MissingCredentials(t *testing.T) {
	//nolint:gosec // os.Args[0] is the test binary itself and the flag is a fixed string, not user input.
	cmd := exec.Command(os.Args[0], "-test.run=TestLoginCmd_MissingCredentials_Crasher")
	cmd.Env = append(os.Environ(), "WANT_LOGIN_CRASH=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.Success() {
		t.Fatalf("expected loginCmd.Run to exit with a non-zero status, got err=%v", err)
	}
	if !strings.Contains(stderr.String(), "are required for login") {
		t.Errorf("expected missing-credentials error on stderr, got: %s", stderr.String())
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

	out := captureStdout(func() { loginCmd.Run(loginCmd, nil) })

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

	out := captureStdout(func() { loginCmd.Run(loginCmd, nil) })
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

	out := captureStdout(func() { loginCmd.Run(loginCmd, nil) })
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

// TestLoginCmd_LoginFailure_Crasher is invoked as a subprocess by
// TestLoginCmd_LoginFailure to exercise loginCmd's fatal()-triggered
// os.Exit(1) path without terminating the real test binary.
func TestLoginCmd_LoginFailure_Crasher(t *testing.T) {
	if os.Getenv("WANT_LOGIN_FAILURE_CRASH") != "1" {
		t.Skip("only runs as a subprocess of TestLoginCmd_LoginFailure")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	lib.AuthSessionURL = srv.URL + "/auth/session"

	email, password = "u@example.com", "pw"
	loginCmd.Run(loginCmd, nil)
}

func TestLoginCmd_LoginFailure(t *testing.T) {
	//nolint:gosec // os.Args[0] is the test binary itself and the flag is a fixed string, not user input.
	cmd := exec.Command(os.Args[0], "-test.run=TestLoginCmd_LoginFailure_Crasher")
	cmd.Env = append(os.Environ(), "WANT_LOGIN_FAILURE_CRASH=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.Success() {
		t.Fatalf("expected loginCmd.Run to exit with a non-zero status on login failure, got err=%v", err)
	}
	if !strings.Contains(stderr.String(), "Error: logging in") {
		t.Errorf("expected login error on stderr, got: %s", stderr.String())
	}
}

func TestLoginCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "login" {
			found = true
			break
		}
	}
	if !found {
		t.Error("login command not registered on root")
	}
}
