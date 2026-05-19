//go:build integration

package lib

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// loadSkylightConfig reads ~/.skylight/config and populates the provided
// pointers only when the corresponding value is currently empty.
// Env vars take precedence — call this after os.Getenv.
func loadSkylightConfig(email, password, frameID *string) {
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
		"SKYLIGHT_EMAIL":    email,
		"SKYLIGHT_PASSWORD": password,
		"SKYLIGHT_FRAME_ID": frameID,
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
