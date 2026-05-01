package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var knownConfigKeys = []string{
	"SKYLIGHT_EMAIL",
	"SKYLIGHT_PASSWORD",
	"SKYLIGHT_TOKEN",
	"SKYLIGHT_USER_ID",
	"SKYLIGHT_FRAME_ID",
	"SKYLIGHT_REFRESH_TOKEN",
	"SKYLIGHT_DEVICE_FINGERPRINT",
}

var sensitiveConfigKeys = map[string]bool{
	"SKYLIGHT_PASSWORD":      true,
	"SKYLIGHT_TOKEN":         true,
	"SKYLIGHT_REFRESH_TOKEN": true,
}

const notSet = "(not set)"
const masked = "****"

// configValues returns pointers to the package-level config globals.
// Must stay in sync with the vars map in loadConfig().
func configValues() map[string]*string {
	return map[string]*string{
		"SKYLIGHT_EMAIL":              &email,
		"SKYLIGHT_PASSWORD":           &password,
		"SKYLIGHT_TOKEN":              &token,
		"SKYLIGHT_USER_ID":            &userID,
		"SKYLIGHT_FRAME_ID":           &frameID,
		"SKYLIGHT_REFRESH_TOKEN":      &refreshToken,
		"SKYLIGHT_DEVICE_FINGERPRINT": &deviceFingerprint,
	}
}

func maskValue(key, value string) string {
	if value == "" {
		return notSet
	}
	if sensitiveConfigKeys[key] {
		if len(value) <= 4 {
			return masked
		}
		return value[:4] + masked
	}
	return value
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify the local configuration file",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display all current configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		path := configPath
		if path == "" {
			path = defaultConfigPath()
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n\n", path)

		vals := configValues()
		for _, key := range knownConfigKeys {
			fmt.Fprintf(cmd.OutOrStdout(), "%-32s %s\n", key, maskValue(key, *vals[key]))
		}
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get the value of a configuration key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		vals := configValues()
		ptr, ok := vals[key]
		if !ok {
			return fmt.Errorf("unknown config key %q; valid keys: %s", key, strings.Join(knownConfigKeys, ", "))
		}
		if *ptr == "" {
			fmt.Fprintln(cmd.OutOrStdout(), notSet)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), *ptr)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration key to a value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		vals := configValues()
		if _, ok := vals[key]; !ok {
			return fmt.Errorf("unknown config key %q; valid keys: %s", key, strings.Join(knownConfigKeys, ", "))
		}

		if err := saveConfig(map[string]string{key: value}); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		path := configPath
		if path == "" {
			path = defaultConfigPath()
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Set %s in %s\n", key, path)
		return nil
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Remove a key from the configuration file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if _, ok := configValues()[key]; !ok {
			return fmt.Errorf("unknown config key %q; valid keys: %s", key, strings.Join(knownConfigKeys, ", "))
		}

		removed, err := deleteFromConfig(key)
		if err != nil {
			return fmt.Errorf("updating config: %w", err)
		}

		path := configPath
		if path == "" {
			path = defaultConfigPath()
		}
		if !removed {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s was not set in %s\n", key, path)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Unset %s in %s\n", key, path)
		}
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the configuration file in $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := configPath
		if path == "" {
			path = defaultConfigPath()
		}
		if path == "" {
			return fmt.Errorf("could not determine config path")
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return fmt.Errorf("creating config directory: %w", err)
			}
			f, err := os.OpenFile(path, os.O_CREATE, 0o600)
			if err != nil {
				return fmt.Errorf("creating config file: %w", err)
			}
			f.Close()
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vi"
		}

		c := exec.Command(editor, path) //nolint:gosec
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configEditCmd)
}
