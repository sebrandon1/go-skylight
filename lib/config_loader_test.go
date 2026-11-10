//go:build integration

package lib

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadSkylightConfig reads ~/.skylight/config and populates the provided
// pointers only when the corresponding value is currently empty.
func loadSkylightConfig(email, password, frameID, refreshToken, deviceFingerprint *string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	f, err := os.Open(filepath.Join(home, ".skylight", "config"))
	if err != nil {
		return
	}
	defer f.Close()

	vars := map[string]*string{
		"SKYLIGHT_EMAIL":              email,
		"SKYLIGHT_PASSWORD":           password,
		"SKYLIGHT_FRAME_ID":           frameID,
		"SKYLIGHT_REFRESH_TOKEN":      refreshToken,
		"SKYLIGHT_DEVICE_FINGERPRINT": deviceFingerprint,
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if ptr, exists := vars[strings.TrimSpace(key)]; exists && *ptr == "" {
			*ptr = strings.TrimSpace(value)
		}
	}
}

// persistRotatedRefreshToken replaces oldToken with newToken in the saved
// config, so a test run leaves the CLI login working. It only touches a config
// that holds oldToken, leaving a config alone when the token came from the
// environment. Tokens are never printed.
func persistRotatedRefreshToken(oldToken, newToken string) {
	if newToken == "" || newToken == oldToken {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path := filepath.Join(home, ".skylight", "config")
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(raw), "\n")
	changed := false
	for i, line := range lines {
		if strings.TrimSpace(line) == "SKYLIGHT_REFRESH_TOKEN="+oldToken {
			lines[i] = "SKYLIGHT_REFRESH_TOKEN=" + newToken
			changed = true
		}
	}
	if !changed {
		return
	}

	if err := replaceFile(path, []byte(strings.Join(lines, "\n"))); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save the rotated refresh token, the next run will need a new login: %v\n", err)
	}
}

// replaceFile swaps in new contents atomically, keeping the file private.
func replaceFile(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-integration-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
