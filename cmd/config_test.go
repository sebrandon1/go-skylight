package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `SKYLIGHT_EMAIL=test@example.com
SKYLIGHT_PASSWORD=secret123
SKYLIGHT_TOKEN=tok123
SKYLIGHT_USER_ID=uid456
SKYLIGHT_FRAME_ID=fid789
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Reset globals
	email = ""
	password = ""
	token = ""
	userID = ""
	frameID = ""
	configPath = path

	loadConfig()

	if email != "test@example.com" {
		t.Errorf("email = %q, want %q", email, "test@example.com")
	}
	if password != "secret123" {
		t.Errorf("password = %q, want %q", password, "secret123")
	}
	if token != "tok123" {
		t.Errorf("token = %q, want %q", token, "tok123")
	}
	if userID != "uid456" {
		t.Errorf("userID = %q, want %q", userID, "uid456")
	}
	if frameID != "fid789" {
		t.Errorf("frameID = %q, want %q", frameID, "fid789")
	}

	// Reset for other tests
	email = ""
	password = ""
	token = ""
	userID = ""
	frameID = ""
	configPath = ""
}

func TestLoadConfigCLIFlagsTakePrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `SKYLIGHT_EMAIL=config@example.com
SKYLIGHT_TOKEN=config-token
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Simulate CLI flags already set
	email = "cli@example.com"
	token = ""
	configPath = path

	loadConfig()

	if email != "cli@example.com" {
		t.Errorf("email = %q, want CLI value %q", email, "cli@example.com")
	}
	if token != "config-token" {
		t.Errorf("token = %q, want config value %q", token, "config-token")
	}

	// Reset
	email = ""
	token = ""
	configPath = ""
}

func TestLoadConfigSkipsComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `# This is a comment
SKYLIGHT_EMAIL=test@example.com

# Another comment
SKYLIGHT_TOKEN=tok123
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	email = ""
	token = ""
	configPath = path

	loadConfig()

	if email != "test@example.com" {
		t.Errorf("email = %q, want %q", email, "test@example.com")
	}
	if token != "tok123" {
		t.Errorf("token = %q, want %q", token, "tok123")
	}

	email = ""
	token = ""
	configPath = ""
}

func TestLoadConfigWarnsOnWorldReadablePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("SKYLIGHT_EMAIL=test@example.com\n"), 0o644); err != nil { //nolint:gosec // intentionally world-readable to test the permission warning
		t.Fatal(err)
	}

	email = ""
	configPath = path

	stderr := captureStderr(func() {
		loadConfig()
	})

	if !strings.Contains(stderr, "readable by group/other") {
		t.Errorf("expected permission warning on stderr, got %q", stderr)
	}
	if email != "test@example.com" {
		t.Errorf("email = %q, want %q (loading should still succeed)", email, "test@example.com")
	}

	email = ""
	configPath = ""
}

func TestLoadConfigNoWarningOnSecurePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("SKYLIGHT_EMAIL=test@example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	email = ""
	configPath = path

	stderr := captureStderr(func() {
		loadConfig()
	})

	if strings.Contains(stderr, "readable by group/other") {
		t.Errorf("did not expect permission warning on stderr, got %q", stderr)
	}

	email = ""
	configPath = ""
}

func TestLoadConfigMissingFile(t *testing.T) {
	configPath = "/nonexistent/path/config"
	email = ""

	loadConfig()

	if email != "" {
		t.Errorf("email should be empty when config file is missing")
	}

	configPath = ""
}

func TestLoadConfig_OutputFormat(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	if err := os.WriteFile(configPath, []byte("SKYLIGHT_OUTPUT=table\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	loadConfig()
	if outputFormat != outputTable {
		t.Errorf("outputFormat = %q, want %q", outputFormat, outputTable)
	}
}

func TestLoadConfig_OutputFormatFlagTakesPrecedence(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	if err := os.WriteFile(configPath, []byte("SKYLIGHT_OUTPUT=json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	outputFormat = outputTable // simulate --output table flag already set
	loadConfig()
	if outputFormat != outputTable {
		t.Errorf("outputFormat = %q, want CLI value %q", outputFormat, outputTable)
	}
}

func TestLoadConfig_QuietTrue(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	if err := os.WriteFile(configPath, []byte("SKYLIGHT_QUIET=true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	loadConfig()
	if !quiet {
		t.Error("quiet should be true after loading SKYLIGHT_QUIET=true")
	}
	if quietConfigStr != "true" {
		t.Errorf("quietConfigStr = %q, want %q", quietConfigStr, "true")
	}
}

func TestLoadConfig_QuietFalse(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	if err := os.WriteFile(configPath, []byte("SKYLIGHT_QUIET=false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	loadConfig()
	if quiet {
		t.Error("quiet should remain false after loading SKYLIGHT_QUIET=false")
	}
	if quietConfigStr != "false" {
		t.Errorf("quietConfigStr = %q, want %q", quietConfigStr, "false")
	}
}

func TestLoadConfig_QuietFlagTakesPrecedence(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	if err := os.WriteFile(configPath, []byte("SKYLIGHT_QUIET=true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	quiet = true // simulate --quiet flag already set
	loadConfig()
	if !quiet {
		t.Error("quiet should remain true when CLI flag was set")
	}
}

func TestSaveConfigMkdirAllError(t *testing.T) {
	// Create a file where a directory should be
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	configPath = filepath.Join(blocker, "subdir", "config")
	err := saveConfig(map[string]string{"KEY": "VALUE"})
	if err == nil {
		t.Error("expected error when MkdirAll fails")
	}
	configPath = ""
}

func TestSaveConfigWriteError(t *testing.T) {
	dir := t.TempDir()
	readonlyDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(readonlyDir, 0o500); err != nil {
		t.Fatal(err)
	}
	configPath = filepath.Join(readonlyDir, "config")
	err := saveConfig(map[string]string{"KEY": "VALUE"})
	if err == nil {
		t.Error("expected error when file cannot be written")
	}
	configPath = ""
}

func TestSaveConfigMergesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	// Write initial config
	if err := os.WriteFile(path, []byte("EXISTING_KEY=old_value\nSKYLIGHT_EMAIL=test@test.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	configPath = path

	err := saveConfig(map[string]string{"SKYLIGHT_TOKEN": "newtok"})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// EXISTING_KEY should be preserved
	if !strings.Contains(content, "EXISTING_KEY=old_value") {
		t.Error("existing key should be preserved")
	}
	// New key should be added
	if !strings.Contains(content, "SKYLIGHT_TOKEN=newtok") {
		t.Error("new key should be added")
	}
	// Original key should be preserved
	if !strings.Contains(content, "SKYLIGHT_EMAIL=test@test.com") {
		t.Error("original SKYLIGHT_EMAIL should be preserved")
	}

	configPath = ""
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".skylight", "config")
	configPath = path

	err := saveConfig(map[string]string{
		"SKYLIGHT_EMAIL":   "test@example.com",
		"SKYLIGHT_TOKEN":   "tok123",
		"SKYLIGHT_USER_ID": "uid456",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if content == "" {
		t.Error("config file should not be empty")
	}

	// Verify the saved file can be read back
	email = ""
	token = ""
	userID = ""
	loadConfig()

	if email != "test@example.com" {
		t.Errorf("email = %q, want %q", email, "test@example.com")
	}
	if token != "tok123" {
		t.Errorf("token = %q, want %q", token, "tok123")
	}
	if userID != "uid456" {
		t.Errorf("userID = %q, want %q", userID, "uid456")
	}

	email = ""
	token = ""
	userID = ""
	configPath = ""
}
