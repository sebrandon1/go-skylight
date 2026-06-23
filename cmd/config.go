package cmd

import (
	"bufio"
	"fmt"
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

	vars := map[string]*string{
		"SKYLIGHT_EMAIL":              &email,
		"SKYLIGHT_PASSWORD":           &password,
		"SKYLIGHT_TOKEN":              &token,
		"SKYLIGHT_USER_ID":            &userID,
		"SKYLIGHT_FRAME_ID":           &frameID,
		"SKYLIGHT_REFRESH_TOKEN":      &refreshToken,
		"SKYLIGHT_DEVICE_FINGERPRINT": &deviceFingerprint,
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
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if ptr, exists := vars[key]; exists && *ptr == "" {
			*ptr = value
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: error reading config file: %v\n", err)
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
		scanner := bufio.NewScanner(f)
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
			existing[key] = val
			orderedKeys = append(orderedKeys, key)
		}
		f.Close()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading existing config: %w", err)
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
