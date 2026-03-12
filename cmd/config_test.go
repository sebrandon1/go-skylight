package cmd

import (
	"os"
	"path/filepath"
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

func TestLoadConfigMissingFile(t *testing.T) {
	configPath = "/nonexistent/path/config"
	email = ""

	loadConfig()

	if email != "" {
		t.Errorf("email should be empty when config file is missing")
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
