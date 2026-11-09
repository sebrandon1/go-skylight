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

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte("new")); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("content = %q, want new", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("mode = %04o, want 0600", perm)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected only the config in %s, found %d entries", dir, len(entries))
	}
}

func TestWriteFileAtomic_FailureLeavesOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("KEY=original\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	if err := writeFileAtomic(path, []byte("KEY=replaced\n")); err == nil {
		t.Fatal("expected error writing into a read-only directory")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "KEY=original\n" {
		t.Errorf("original changed: %q", got)
	}
}

func TestSaveConfig_ReplacesSymlinkTarget(t *testing.T) {
	resetAuthState(t)
	dir := t.TempDir()
	target := filepath.Join(dir, "real-config")
	if err := os.WriteFile(target, []byte("A=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	configPath = link

	if err := saveConfig(map[string]string{"B": "2"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("symlink was replaced by a regular file")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "A=1\nB=2\n" {
		t.Errorf("target content = %q", got)
	}
}

func TestDeleteFromConfig(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		key       string
		wantFound bool
		want      string
	}{
		{"removes key and keeps comments", "# note\nA=1\nB=2\n", "A", true, "# note\nB=2\n"},
		{"missing key leaves file alone", "A=1\n", "B", false, "A=1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAuthState(t)
			configPath = filepath.Join(t.TempDir(), "config")
			if err := os.WriteFile(configPath, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}
			found, err := deleteFromConfig(tt.key)
			if err != nil {
				t.Fatalf("deleteFromConfig: %v", err)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			got, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("content = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeleteFromConfig_MissingFile(t *testing.T) {
	resetAuthState(t)
	configPath = filepath.Join(t.TempDir(), "config")
	found, err := deleteFromConfig("A")
	if err != nil || found {
		t.Fatalf("got found=%v err=%v, want false, nil", found, err)
	}
}

func TestConfigFunctions_NoConfigPath(t *testing.T) {
	resetAuthState(t)
	configPath = ""
	t.Setenv("HOME", "")
	if err := saveConfig(map[string]string{"A": "1"}); err == nil {
		t.Error("saveConfig: expected an error with no config path")
	}
	if _, err := deleteFromConfig("A"); err == nil {
		t.Error("deleteFromConfig: expected an error with no config path")
	}
	if _, err := readConfigValues(); err == nil {
		t.Error("readConfigValues: expected an error with no config path")
	}
}

func TestConfigWrites_ReadOnlyDirectoryKeepsFile(t *testing.T) {
	resetAuthState(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	const original = "A=1\n"
	if err := os.WriteFile(configPath, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	// An existing lock file can still be opened in a read-only directory, so
	// the failure comes from the atomic write.
	if err := os.WriteFile(configPath+".lock", nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	if err := saveConfig(map[string]string{"B": "2"}); err == nil {
		t.Error("saveConfig: expected a write error")
	}
	if _, err := deleteFromConfig("A"); err == nil {
		t.Error("deleteFromConfig: expected a write error")
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Errorf("config changed: %q", got)
	}
}

func TestWriteFileAtomic_RenameFailureRemovesTemp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target")
	if err := os.MkdirAll(filepath.Join(path, "child"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte("data")); err == nil {
		t.Fatal("expected a rename error when the target is a directory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("temp file left behind: %v", entries)
	}
}
