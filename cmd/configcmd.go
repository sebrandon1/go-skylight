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
	cfgEmail,
	cfgPassword,
	cfgToken,
	cfgUserID,
	cfgFrameID,
	cfgRefreshToken,
	cfgDeviceFingerprint,
	cfgOutput,
	cfgQuiet,
}

var sensitiveConfigKeys = map[string]bool{
	cfgPassword:     true,
	cfgToken:        true,
	cfgRefreshToken: true,
}

const notSet = "(not set)"
const masked = "****"

// configValues returns pointers to the package-level config globals.
// Must stay in sync with the vars map in loadConfig().
func configValues() map[string]*string {
	return map[string]*string{
		cfgEmail:             &email,
		cfgPassword:          &password,
		cfgToken:             &token,
		cfgUserID:            &userID,
		cfgFrameID:           &frameID,
		cfgRefreshToken:      &refreshToken,
		cfgDeviceFingerprint: &deviceFingerprint,
		cfgOutput:            &outputFormat,
		cfgQuiet:             &quietConfigStr,
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
	Use:   subConfig,
	Short: "View and modify the local configuration file",
}

var configShowCmd = &cobra.Command{
	Use:   subShow,
	Short: "Display all current configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := configPath
		if path == "" {
			path = defaultConfigPath()
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n\n", path)

		vals := configValues()
		for _, key := range knownConfigKeys {
			fmt.Fprintf(cmd.OutOrStdout(), "%-32s %s\n", key, maskValue(key, *vals[key]))
		}
		return nil
	},
}

var configGetReveal bool

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
			return nil
		}
		if configGetReveal {
			fmt.Fprintln(cmd.OutOrStdout(), *ptr)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), maskValue(key, *ptr))
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

		c := editorCommand(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

// editorCommand builds an *exec.Cmd for $EDITOR values that may include args
// (e.g. "code --wait", "emacs -nw").
func editorCommand(editor, path string) *exec.Cmd {
	fields := strings.Fields(editor)
	if len(fields) == 0 {
		fields = []string{"vi"}
	}
	args := append(append([]string{}, fields[1:]...), path)
	return exec.Command(fields[0], args...) //nolint:gosec
}

func init() {
	configGetCmd.Flags().BoolVar(&configGetReveal, "reveal", false, "Print the full value of sensitive keys instead of masking")
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configEditCmd)
}
