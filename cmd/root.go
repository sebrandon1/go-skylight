package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	email             string
	password          string
	token             string
	userID            string
	frameID           string
	refreshToken      string
	deviceFingerprint string
	autoClient        *lib.Client
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "skylight",
	Short:   "Skylight CLI interacts with the Skylight Calendar API",
	Version: version,
}

// SetVersion sets the version string for the root command.
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get objects from Skylight",
}

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Load config file first (CLI flags take precedence since they're already set)
		loadConfig()

		// Skip auto-login for login command itself and help
		if cmd.Name() == loginCmd.Name() || cmd.Name() == "help" {
			return nil
		}

		// Prefer OAuth2 refresh token flow (new API).
		if refreshToken != "" {
			fingerprint := deviceFingerprint
			if fingerprint == "" {
				fingerprint = defaultFingerprint()
			}
			c, err := lib.NewClientWithRefreshToken(refreshToken, fingerprint)
			if err != nil {
				return fmt.Errorf("auto-login failed: %w", err)
			}
			autoClient = c
			// Persist the rotated refresh token back to config.
			if c.RefreshToken != "" && c.RefreshToken != refreshToken {
				_ = saveConfig(map[string]string{
					"SKYLIGHT_REFRESH_TOKEN":      c.RefreshToken,
					"SKYLIGHT_DEVICE_FINGERPRINT": fingerprint,
				})
				refreshToken = c.RefreshToken
			}
			return nil
		}

		// Legacy: email/password via deprecated /api/sessions endpoint.
		if email != "" && password != "" && (token == "" || userID == "") {
			//nolint:staticcheck // intentional fallback for legacy callers.
			c, err := lib.NewClient(email, password)
			if err != nil {
				return fmt.Errorf("auto-login failed: %w", err)
			}
			userID = c.UserID
			token = c.APIToken
			autoClient = c
		}
		return nil
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path (default ~/.skylight/config)")
	rootCmd.PersistentFlags().StringVar(&email, "email", "", "Skylight account email (deprecated: use --refresh-token)")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Skylight account password (deprecated: use --refresh-token)")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Bearer access token (alternative to --refresh-token)")
	rootCmd.PersistentFlags().StringVar(&userID, "user-id", "", "User ID (used with --token)")
	rootCmd.PersistentFlags().StringVar(&frameID, "frame-id", "", "Frame ID")
	rootCmd.PersistentFlags().StringVar(&refreshToken, "refresh-token", "", "OAuth2 refresh token")
	rootCmd.PersistentFlags().StringVar(&deviceFingerprint, "device-fingerprint", "", "Device fingerprint UUID (stable per device)")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(bountyCmd)
	rootCmd.AddCommand(rotationCmd)

	// Top-level resource commands (flattened from get prefix)
	rootCmd.AddCommand(calendarCmd)
	rootCmd.AddCommand(choreCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(rewardCmd)
	rootCmd.AddCommand(mealCmd)
	rootCmd.AddCommand(categoryCmd)
	rootCmd.AddCommand(frameCmd)
	rootCmd.AddCommand(photoCmd)

	// Backward compatibility: keep get subcommands but hide them
	getCmd.AddCommand(calendarCmd)
	getCmd.AddCommand(choreCmd)
	getCmd.AddCommand(listCmd)
	getCmd.AddCommand(rewardCmd)
	getCmd.AddCommand(mealCmd)
	getCmd.AddCommand(categoryCmd)
	getCmd.AddCommand(frameCmd)
	getCmd.AddCommand(photoCmd)
	getCmd.Hidden = true
}

func requireFrameID() {
	if frameID == "" {
		fmt.Println("Error: --frame-id is required")
		os.Exit(1)
	}
}

func getClient() *lib.Client {
	if autoClient != nil {
		return autoClient
	}
	client, err := lib.NewClientWithToken(userID, token)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	return client
}

// defaultFingerprint returns a stable UUID derived from the hostname,
// or a fixed fallback if the hostname cannot be determined.
func defaultFingerprint() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "00000000-0000-0000-0000-000000000001"
	}
	// Derive a deterministic UUID from the hostname bytes (version 4 format,
	// not cryptographically random, but stable across invocations).
	h := fnv32(host)
	return fmt.Sprintf("%08x-0000-4000-8000-%012x", h, h)
}

func fnv32(s string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
