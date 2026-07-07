package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var configPath string

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".skylight", "config")
}

// parseConfigFile reads KEY=VALUE entries from r, skipping blank lines and # comments.
// Returns the parsed values and the keys in file order.
func parseConfigFile(r io.Reader) (map[string]string, []string, error) {
	values := make(map[string]string)
	var keys []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		values[key] = val
		keys = append(keys, key)
	}
	return values, keys, scanner.Err()
}

func loadConfig() {
	path := configPath
	if path == "" {
		path = defaultConfigPath()
	}
	if path == "" {
		return
	}

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	if info, statErr := f.Stat(); statErr == nil {
		if perm := info.Mode().Perm(); perm&0o077 != 0 { // 0o077: group/other read-write-execute bits
			fmt.Fprintf(os.Stderr, "warning: config file %s is readable by group/other (mode %04o); it contains credentials, run: chmod 600 %s\n", path, perm, path)
		}
	}

	vars := map[string]*string{
		"SKYLIGHT_EMAIL":              &email,
		"SKYLIGHT_PASSWORD":           &password,
		"SKYLIGHT_TOKEN":              &token,
		"SKYLIGHT_USER_ID":            &userID,
		"SKYLIGHT_FRAME_ID":           &frameID,
		"SKYLIGHT_REFRESH_TOKEN":      &refreshToken,
		"SKYLIGHT_DEVICE_FINGERPRINT": &deviceFingerprint,
	}

	values, _, err := parseConfigFile(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: error reading config file: %v\n", err)
	}
	for key, value := range values {
		if ptr, exists := vars[key]; exists && *ptr == "" {
			*ptr = value
		}
	}
}

func saveConfig(values map[string]string) error {
	path := configPath
	if path == "" {
		path = defaultConfigPath()
	}
	if path == "" {
		return fmt.Errorf("could not determine config path")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Read existing config to preserve unknown keys
	existing := make(map[string]string)
	var orderedKeys []string
	if f, err := os.Open(path); err == nil {
		var scanErr error
		existing, orderedKeys, scanErr = parseConfigFile(f)
		f.Close()
		if scanErr != nil {
			return fmt.Errorf("reading existing config: %w", scanErr)
		}
	}

	// Merge new values
	for k, v := range values {
		if _, exists := existing[k]; !exists {
			orderedKeys = append(orderedKeys, k)
		}
		existing[k] = v
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	defer f.Close()

	for _, k := range orderedKeys {
		fmt.Fprintf(f, "%s=%s\n", k, existing[k])
	}

	return nil
}

// deleteFromConfig removes key from the config file.
// Returns true if found and removed, false if the key was not present.
func deleteFromConfig(key string) (bool, error) {
	path := configPath
	if path == "" {
		path = defaultConfigPath()
	}
	if path == "" {
		return false, fmt.Errorf("could not determine config path")
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	var lines []string
	found := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}
		k, _, ok := strings.Cut(trimmed, "=")
		if ok && strings.TrimSpace(k) == key {
			found = true
			continue
		}
		lines = append(lines, line)
	}
	f.Close()
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("reading config: %w", err)
	}

	if !found {
		return false, nil
	}

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return false, err
	}
	defer out.Close()
	for _, line := range lines {
		fmt.Fprintln(out, line)
	}
	return true, nil
}
