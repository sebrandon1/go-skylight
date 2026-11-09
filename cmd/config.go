package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	configPath     string
	quietConfigStr string // config-file value of SKYLIGHT_QUIET, for config show/get
)

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

	values, _, err := parseConfigFile(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: error reading config file: %v\n", err)
	}
	for key, value := range values {
		if ptr, exists := configValues()[key]; exists && *ptr == "" {
			*ptr = value
		}
	}
	// cfgQuiet is stored as a string in the config file but drives a bool at runtime.
	if v, ok := values[cfgQuiet]; ok && !quiet && strings.EqualFold(v, "true") {
		quiet = true
	}
}

// resolveConfigPath returns the config file to read and write. A symlinked
// config resolves to its target, so an atomic rename replaces the real file
// rather than the link.
func resolveConfigPath() (string, error) {
	path := configPath
	if path == "" {
		path = defaultConfigPath()
	}
	if path == "" {
		return "", errors.New("could not determine config path")
	}
	if target, err := filepath.EvalSymlinks(path); err == nil {
		return target, nil
	}
	return path, nil
}

// lockConfig takes an exclusive advisory lock on <config>.lock, blocking until
// it is available, and returns the function that releases it.
func lockConfig() (unlock func(), err error) {
	path, err := resolveConfigPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("creating config directory: %w", err)
	}
	f, err := os.OpenFile(path+".lock", os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("opening config lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("locking config: %w", err)
	}
	// Closing the descriptor releases the flock.
	return func() { f.Close() }, nil
}

// writeFileAtomic replaces path with data, mode 0600, through a temp file in
// the same directory, so a failed write leaves any existing file untouched.
func writeFileAtomic(path string, data []byte) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// saveConfig merges values into the config file under the config lock.
func saveConfig(values map[string]string) error {
	unlock, err := lockConfig()
	if err != nil {
		return err
	}
	defer unlock()
	return saveConfigLocked(values)
}

// saveConfigLocked merges values into the config file. The caller holds the
// config lock.
func saveConfigLocked(values map[string]string) error {
	path, err := resolveConfigPath()
	if err != nil {
		return err
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

	var buf bytes.Buffer
	for _, k := range orderedKeys {
		fmt.Fprintf(&buf, "%s=%s\n", k, existing[k])
	}
	if err := writeFileAtomic(path, buf.Bytes()); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// readConfigValues returns the key/value pairs currently in the config file.
func readConfigValues() (map[string]string, error) {
	path, err := resolveConfigPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	values, _, err := parseConfigFile(f)
	return values, err
}

// deleteFromConfig removes key from the config file.
// Returns true if found and removed, false if the key was not present.
func deleteFromConfig(key string) (bool, error) {
	path, err := resolveConfigPath()
	if err != nil {
		return false, err
	}

	unlock, err := lockConfig()
	if err != nil {
		return false, err
	}
	defer unlock()

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

	var buf bytes.Buffer
	for _, line := range lines {
		fmt.Fprintln(&buf, line)
	}
	if err := writeFileAtomic(path, buf.Bytes()); err != nil {
		return false, fmt.Errorf("writing config: %w", err)
	}
	return true, nil
}
