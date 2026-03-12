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
		"SKYLIGHT_EMAIL":    &email,
		"SKYLIGHT_PASSWORD": &password,
		"SKYLIGHT_TOKEN":    &token,
		"SKYLIGHT_USER_ID":  &userID,
		"SKYLIGHT_FRAME_ID": &frameID,
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
