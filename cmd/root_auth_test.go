package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

// resetAuthState saves all package-level auth/config vars touched by
// PersistentPreRunE and getClient, restoring them after the test.
func resetAuthState(t *testing.T) {
	t.Helper()
	origEmail, origPassword := email, password
	origToken, origUserID := token, userID
	origRefreshToken, origFingerprint := refreshToken, deviceFingerprint
	origAutoClient := autoClient
	origConfigPath := configPath

	t.Cleanup(func() {
		email, password = origEmail, origPassword
		token, userID = origToken, origUserID
		refreshToken, deviceFingerprint = origRefreshToken, origFingerprint
		autoClient = origAutoClient
		configPath = origConfigPath
	})

	// loadConfig() (called at the top of PersistentPreRunE) silently fills
	// any empty var from ~/.skylight/config when configPath is "". Point it
	// at a guaranteed-empty path so tests aren't at the mercy of whatever
	// config happens to exist on the machine running them.
	configPath = filepath.Join(t.TempDir(), "nonexistent-config")
}

func TestRequireFrameID_PassesWhenSet(t *testing.T) {
	origFrameID := frameID
	frameID = "f1"
	t.Cleanup(func() { frameID = origFrameID })

	requireFrameID()
}

func TestRequireFrameID_ReturnsErrorWhenEmpty(t *testing.T) {
	origFrameID := frameID
	frameID = ""
	t.Cleanup(func() { frameID = origFrameID })

	if err := requireFrameID(); err == nil {
		t.Error("expected error when frameID is empty")
	}
}

func TestGetClient_ReturnsAutoClient(t *testing.T) {
	resetAuthState(t)
	want, err := lib.NewClientWithToken("u", "t")
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	autoClient = want

	got, err := getClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Error("expected getClient to return the package-level autoClient unchanged")
	}
}

func TestGetClient_BuildsFromUserIDAndToken(t *testing.T) {
	resetAuthState(t)
	autoClient = nil
	userID, token = "u1", "t1"

	got, err := getClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected a non-nil client")
	}
	if got.UserID != "u1" || got.APIToken != "t1" {
		t.Errorf("got UserID=%q APIToken=%q, want u1/t1", got.UserID, got.APIToken)
	}
}

func TestGetClient_ReturnsErrorWithNoCredentials(t *testing.T) {
	resetAuthState(t)
	autoClient = nil
	userID, token = "", ""

	_, err := getClient()
	if err == nil {
		t.Fatal("expected error when no credentials are set, got nil")
	}
}

func TestSaveConfig_WritesRefreshAndFingerprint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	origConfigPath := configPath
	configPath = path
	t.Cleanup(func() { configPath = origConfigPath })

	if err := saveConfig(map[string]string{
		cfgRefreshToken:      "new-refresh",
		cfgDeviceFingerprint: "fp-123",
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "SKYLIGHT_REFRESH_TOKEN=new-refresh") {
		t.Errorf("expected refresh token in config, got: %s", content)
	}
	if !strings.Contains(content, "SKYLIGHT_DEVICE_FINGERPRINT=fp-123") {
		t.Errorf("expected fingerprint in config, got: %s", content)
	}
}

func TestPersistentPreRunE_SkipsAutoLoginForLogin(t *testing.T) {
	resetAuthState(t)
	autoClient = nil
	email, password = "", ""
	refreshToken = ""

	cmd := &cobra.Command{Use: "login"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if autoClient != nil {
		t.Error("expected autoClient to stay nil when skipping login command")
	}
}

func TestPersistentPreRunE_SkipsAutoLoginForHelp(t *testing.T) {
	resetAuthState(t)
	autoClient = nil

	cmd := &cobra.Command{Use: "help"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersistentPreRunE_SkipsAutoLoginForConfigSubcommands(t *testing.T) {
	resetAuthState(t)
	autoClient = nil

	parent := &cobra.Command{Use: "config"}
	child := &cobra.Command{Use: "show"}
	parent.AddCommand(child)

	if err := rootCmd.PersistentPreRunE(child, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if autoClient != nil {
		t.Error("expected autoClient to stay nil for config subcommands")
	}
}

func TestPersistentPreRunE_RefreshTokenFlow(t *testing.T) {
	resetAuthState(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"newaccess","refresh_token":"rotated-refresh","expires_in":3600}`)
	}))
	defer srv.Close()

	origOAuthURL := lib.OAuthURL
	lib.OAuthURL = srv.URL
	t.Cleanup(func() { lib.OAuthURL = origOAuthURL })

	dir := t.TempDir()
	origConfigPath := configPath
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = origConfigPath })

	autoClient = nil
	refreshToken = "orig-refresh"
	deviceFingerprint = "fp-1"
	email, password = "", ""

	cmd := &cobra.Command{Use: "frame"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if autoClient == nil {
		t.Fatal("expected autoClient to be set")
	}
	if autoClient.APIToken != "newaccess" {
		t.Errorf("got APIToken=%q, want newaccess", autoClient.APIToken)
	}
	if refreshToken != "rotated-refresh" {
		t.Errorf("expected refreshToken to be updated to the rotated value, got %q", refreshToken)
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading persisted config: %v", err)
	}
	if !strings.Contains(string(raw), "SKYLIGHT_REFRESH_TOKEN=rotated-refresh") {
		t.Errorf("expected rotated refresh token persisted to config, got: %s", string(raw))
	}
}

func TestPersistentPreRunE_RefreshTokenFlow_DefaultsFingerprint(t *testing.T) {
	resetAuthState(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"newaccess","refresh_token":"orig-refresh","expires_in":3600}`)
	}))
	defer srv.Close()

	origOAuthURL := lib.OAuthURL
	lib.OAuthURL = srv.URL
	t.Cleanup(func() { lib.OAuthURL = origOAuthURL })

	dir := t.TempDir()
	origConfigPath := configPath
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = origConfigPath })

	autoClient = nil
	refreshToken = "orig-refresh"
	deviceFingerprint = ""
	email, password = "", ""

	cmd := &cobra.Command{Use: "frame"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if autoClient == nil {
		t.Fatal("expected autoClient to be set")
	}

	// No token rotation, but a fresh UUID fingerprint must be generated and persisted.
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected fingerprint to be persisted, config not written: %v", err)
	}
	if !strings.Contains(string(raw), "SKYLIGHT_DEVICE_FINGERPRINT=") {
		t.Errorf("expected SKYLIGHT_DEVICE_FINGERPRINT in persisted config, got: %s", string(raw))
	}
}

func TestPersistentPreRunE_RefreshTokenFlow_PersistFailureIsNonFatal(t *testing.T) {
	resetAuthState(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"newaccess","refresh_token":"orig-refresh","expires_in":3600}`)
	}))
	defer srv.Close()

	origOAuthURL := lib.OAuthURL
	lib.OAuthURL = srv.URL
	t.Cleanup(func() { lib.OAuthURL = origOAuthURL })

	// Point config at a path where MkdirAll will fail (file blocks the dir).
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	origConfigPath := configPath
	configPath = filepath.Join(blocker, "subdir", "config")
	t.Cleanup(func() { configPath = origConfigPath })

	autoClient = nil
	refreshToken = "orig-refresh"
	deviceFingerprint = ""
	email, password = "", ""

	// A save failure must not propagate as an error — auth already succeeded.
	cmd := &cobra.Command{Use: "frame"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Errorf("expected non-fatal save failure, got error: %v", err)
	}
	if autoClient == nil {
		t.Error("expected autoClient to be set even when persist fails")
	}
}

func TestPersistentPreRunE_RefreshTokenFlow_Failure(t *testing.T) {
	resetAuthState(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer srv.Close()

	origOAuthURL := lib.OAuthURL
	lib.OAuthURL = srv.URL
	t.Cleanup(func() { lib.OAuthURL = origOAuthURL })

	autoClient = nil
	refreshToken = "bad-refresh"
	deviceFingerprint = "fp-1"
	email, password = "", ""

	cmd := &cobra.Command{Use: "frame"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err == nil {
		t.Error("expected an error when the refresh token exchange fails")
	}
}

func TestPersistentPreRunE_LegacyEmailPasswordFlow(t *testing.T) {
	resetAuthState(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"uid1","type":"authenticated_user","attributes":{"token":"tok1"}}}`)
	}))
	defer srv.Close()

	origSkylightURL := lib.SkylightURL
	lib.SkylightURL = srv.URL + "/api"
	t.Cleanup(func() { lib.SkylightURL = origSkylightURL })

	autoClient = nil
	refreshToken = ""
	email, password = "u@example.com", "pw"
	token, userID = "", ""

	cmd := &cobra.Command{Use: "frame"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != "uid1" || token != "tok1" {
		t.Errorf("got userID=%q token=%q, want uid1/tok1", userID, token)
	}
	if autoClient == nil {
		t.Error("expected autoClient to be set")
	}
}

func TestPersistentPreRunE_LegacyEmailPasswordFlow_Failure(t *testing.T) {
	resetAuthState(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"invalid credentials"}`)
	}))
	defer srv.Close()

	origSkylightURL := lib.SkylightURL
	lib.SkylightURL = srv.URL + "/api"
	t.Cleanup(func() { lib.SkylightURL = origSkylightURL })

	autoClient = nil
	refreshToken = ""
	email, password = "u@example.com", "wrong"
	token, userID = "", ""

	cmd := &cobra.Command{Use: "frame"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err == nil {
		t.Error("expected an error when legacy login fails")
	}
}

func TestPersistentPreRunE_NoCredentials(t *testing.T) {
	resetAuthState(t)
	autoClient = nil
	refreshToken = ""
	email, password = "", ""
	token, userID = "", ""

	cmd := &cobra.Command{Use: "frame"}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if autoClient != nil {
		t.Error("expected autoClient to stay nil with no credentials configured")
	}
}

func TestPersistentPreRunE_RejectsInvalidOutputFormat(t *testing.T) {
	resetAuthState(t)
	origOutputFormat := outputFormat
	t.Cleanup(func() { outputFormat = origOutputFormat })
	outputFormat = "xml"

	cmd := &cobra.Command{Use: "frame"}
	err := rootCmd.PersistentPreRunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid --output value, got nil")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("expected error to mention invalid value, got: %v", err)
	}
}

func TestPersistentPreRunE_AcceptsValidOutputFormats(t *testing.T) {
	resetAuthState(t)
	origOutputFormat := outputFormat
	t.Cleanup(func() { outputFormat = origOutputFormat })

	for _, v := range []string{outputJSON, outputTable} {
		outputFormat = v
		cmd := &cobra.Command{Use: "frame"}
		if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
			t.Errorf("unexpected error for --output=%q: %v", v, err)
		}
	}
}
