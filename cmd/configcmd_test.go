package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaskValueEmpty(t *testing.T) {
	got := maskValue("SKYLIGHT_TOKEN", "")
	if got != notSet {
		t.Errorf("maskValue empty = %q, want %q", got, notSet)
	}
}

func TestMaskValueNonSensitive(t *testing.T) {
	got := maskValue("SKYLIGHT_FRAME_ID", "frame123")
	if got != "frame123" {
		t.Errorf("maskValue non-sensitive = %q, want %q", got, "frame123")
	}
}

func TestMaskValueSensitiveLong(t *testing.T) {
	got := maskValue("SKYLIGHT_TOKEN", "abcdefgh")
	if got != "abcd****" {
		t.Errorf("maskValue sensitive long = %q, want %q", got, "abcd****")
	}
}

func TestMaskValueSensitiveShort(t *testing.T) {
	got := maskValue("SKYLIGHT_TOKEN", "ab")
	if got != "****" {
		t.Errorf("maskValue sensitive short = %q, want %q", got, "****")
	}
}

func TestMaskValueSensitiveExactlyFour(t *testing.T) {
	got := maskValue("SKYLIGHT_PASSWORD", "abcd")
	if got != "****" {
		t.Errorf("maskValue sensitive exactly 4 = %q, want %q", got, "****")
	}
}

func TestConfigShowAllNotSet(t *testing.T) {
	resetGlobals(t)

	var buf bytes.Buffer
	configShowCmd.SetOut(&buf)
	if err := configShowCmd.RunE(configShowCmd, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Config file:") {
		t.Error("config show output missing 'Config file:' line")
	}
	for _, key := range knownConfigKeys {
		if !strings.Contains(out, key) {
			t.Errorf("config show output missing key %s", key)
		}
	}
	if count := strings.Count(out, notSet); count != len(knownConfigKeys) {
		t.Errorf("expected %d %q entries, got %d", len(knownConfigKeys), notSet, count)
	}
}

func TestConfigShowMasksSensitiveKeys(t *testing.T) {
	resetGlobals(t)
	token = "mysecrettoken"
	t.Cleanup(func() { token = "" })

	var buf bytes.Buffer
	configShowCmd.SetOut(&buf)
	if err := configShowCmd.RunE(configShowCmd, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "mysecrettoken") {
		t.Error("config show should not print plaintext token")
	}
	if !strings.Contains(out, "myse****") {
		t.Error("config show should print masked token prefix")
	}
}

func TestConfigShowNonSensitiveUnmasked(t *testing.T) {
	resetGlobals(t)
	frameID = "frame-abc-123"
	t.Cleanup(func() { frameID = "" })

	var buf bytes.Buffer
	configShowCmd.SetOut(&buf)
	if err := configShowCmd.RunE(configShowCmd, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "frame-abc-123") {
		t.Error("config show should print non-sensitive value unmasked")
	}
}

func TestConfigGetUnknownKey(t *testing.T) {
	resetGlobals(t)
	err := configGetCmd.RunE(configGetCmd, []string{"SKYLIGHT_INVALID"})
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
}

func TestConfigGetEmptyValue(t *testing.T) {
	resetGlobals(t)

	var buf bytes.Buffer
	configGetCmd.SetOut(&buf)
	err := configGetCmd.RunE(configGetCmd, []string{"SKYLIGHT_TOKEN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != notSet {
		t.Errorf("expected %q, got %q", notSet, buf.String())
	}
}

func TestConfigGetMasksSensitiveKeys(t *testing.T) {
	resetGlobals(t)
	token = "plaintext-value"
	configGetReveal = false
	t.Cleanup(func() { token = ""; configGetReveal = false })

	var buf bytes.Buffer
	configGetCmd.SetOut(&buf)
	err := configGetCmd.RunE(configGetCmd, []string{"SKYLIGHT_TOKEN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if strings.Contains(out, "plaintext-value") {
		t.Errorf("config get should mask sensitive value, got %q", out)
	}
	if !strings.Contains(out, masked) {
		t.Errorf("config get should contain mask marker %q, got %q", masked, out)
	}
}

func TestConfigGetRevealFlag(t *testing.T) {
	resetGlobals(t)
	token = "plaintext-value"
	configGetReveal = true
	t.Cleanup(func() { token = ""; configGetReveal = false })

	var buf bytes.Buffer
	configGetCmd.SetOut(&buf)
	err := configGetCmd.RunE(configGetCmd, []string{"SKYLIGHT_TOKEN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "plaintext-value" {
		t.Errorf("config get --reveal should print full value, got %q", buf.String())
	}
}

func TestConfigGetNonSensitiveUnmasked(t *testing.T) {
	resetGlobals(t)
	frameID = "frame-abc-123"
	t.Cleanup(func() { frameID = "" })

	var buf bytes.Buffer
	configGetCmd.SetOut(&buf)
	err := configGetCmd.RunE(configGetCmd, []string{"SKYLIGHT_FRAME_ID"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "frame-abc-123" {
		t.Errorf("config get for non-sensitive key should print full value, got %q", buf.String())
	}
}

func TestConfigSetUnknownKey(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = "" })

	err := configSetCmd.RunE(configSetCmd, []string{"SKYLIGHT_INVALID", "val"})
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
	if _, statErr := os.Stat(configPath); !os.IsNotExist(statErr) {
		t.Error("config file should not be created on unknown key error")
	}
}

func TestConfigSetWritesToFile(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = "" })

	var buf bytes.Buffer
	configSetCmd.SetOut(&buf)
	err := configSetCmd.RunE(configSetCmd, []string{"SKYLIGHT_FRAME_ID", "test-frame-789"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("could not read config file: %v", readErr)
	}
	if !strings.Contains(string(data), "SKYLIGHT_FRAME_ID=test-frame-789") {
		t.Errorf("config file missing expected entry, got: %s", data)
	}
}

func TestConfigSetPreservesExistingKeys(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = "" })

	// Pre-populate file
	if err := os.WriteFile(configPath, []byte("SKYLIGHT_USER_ID=existing-user\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := configSetCmd.RunE(configSetCmd, []string{"SKYLIGHT_FRAME_ID", "new-frame"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)
	if !strings.Contains(content, "SKYLIGHT_USER_ID=existing-user") {
		t.Error("config set should preserve existing keys")
	}
	if !strings.Contains(content, "SKYLIGHT_FRAME_ID=new-frame") {
		t.Error("config set should write new key")
	}
}

func TestConfigSetUpdatesExistingKey(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = "" })

	if err := os.WriteFile(configPath, []byte("SKYLIGHT_FRAME_ID=old-frame\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := configSetCmd.RunE(configSetCmd, []string{"SKYLIGHT_FRAME_ID", "new-frame"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)
	if strings.Contains(content, "old-frame") {
		t.Error("config set should overwrite existing key value")
	}
	if !strings.Contains(content, "SKYLIGHT_FRAME_ID=new-frame") {
		t.Error("config set should write updated value")
	}
}

func TestConfigUnsetUnknownKey(t *testing.T) {
	resetGlobals(t)
	err := configUnsetCmd.RunE(configUnsetCmd, []string{"SKYLIGHT_INVALID"})
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
}

func TestConfigUnsetKeyNotInFile(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = "" })

	if err := os.WriteFile(configPath, []byte("SKYLIGHT_FRAME_ID=frame1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var errBuf bytes.Buffer
	configUnsetCmd.SetErr(&errBuf)
	if err := configUnsetCmd.RunE(configUnsetCmd, []string{"SKYLIGHT_TOKEN"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "warning") {
		t.Errorf("expected warning for missing key, got: %q", errBuf.String())
	}
}

func TestConfigUnsetRemovesKey(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = "" })

	if err := os.WriteFile(configPath, []byte("SKYLIGHT_FRAME_ID=frame1\nSKYLIGHT_TOKEN=tok123\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var outBuf bytes.Buffer
	configUnsetCmd.SetOut(&outBuf)
	if err := configUnsetCmd.RunE(configUnsetCmd, []string{"SKYLIGHT_TOKEN"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)
	if strings.Contains(content, "SKYLIGHT_TOKEN") {
		t.Error("config unset should have removed SKYLIGHT_TOKEN")
	}
	if !strings.Contains(content, "SKYLIGHT_FRAME_ID=frame1") {
		t.Error("config unset should preserve other keys")
	}
	if !strings.Contains(outBuf.String(), "Unset") {
		t.Errorf("expected 'Unset' in output, got: %q", outBuf.String())
	}
}

func TestConfigEditCreatesFileIfMissing(t *testing.T) {
	resetGlobals(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config")
	t.Cleanup(func() { configPath = "" })

	t.Setenv("EDITOR", "true")

	if err := configEditCmd.RunE(configEditCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config edit should create config file if missing")
	}
}

func TestEditorCommandSplitsArgs(t *testing.T) {
	// #261: EDITOR="code --wait" must not treat the whole string as argv0.
	c := editorCommand("code --wait", "/tmp/config")
	if c.Path != "code" && c.Args[0] != "code" {
		// Path may be resolved; Args[0] is the program name as invoked.
		if len(c.Args) == 0 || c.Args[0] != "code" {
			t.Fatalf("program: Path=%q Args=%v", c.Path, c.Args)
		}
	}
	if len(c.Args) < 3 || c.Args[1] != "--wait" || c.Args[len(c.Args)-1] != "/tmp/config" {
		t.Fatalf("args = %v, want [code --wait /tmp/config]", c.Args)
	}

	c2 := editorCommand("vi", "/tmp/config")
	if len(c2.Args) != 2 || c2.Args[0] != "vi" || c2.Args[1] != "/tmp/config" {
		t.Fatalf("simple editor args = %v", c2.Args)
	}
}

// resetGlobals clears all config globals and resets them via t.Cleanup.
func resetGlobals(t *testing.T) {
	t.Helper()
	prev := struct{ e, p, tok, uid, fid, rt, df, cp string }{
		email, password, token, userID, frameID, refreshToken, deviceFingerprint, configPath,
	}
	email, password, token, userID, frameID, refreshToken, deviceFingerprint, configPath = "", "", "", "", "", "", "", ""
	t.Cleanup(func() {
		email = prev.e
		password = prev.p
		token = prev.tok
		userID = prev.uid
		frameID = prev.fid
		refreshToken = prev.rt
		deviceFingerprint = prev.df
		configPath = prev.cp
	})
}
