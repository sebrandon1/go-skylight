package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

const (
	outputJSON  = "json"
	outputTable = "table"
)

var (
	email             string
	password          string
	token             string
	userID            string
	frameID           string
	refreshToken      string
	deviceFingerprint string
	outputFormat      string
	quiet             bool
	autoClient        *lib.Client
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:          "skylight",
	Short:        "Skylight CLI interacts with the Skylight Calendar API",
	Version:      version,
	SilenceUsage: true,
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
	rootCmd.PersistentPreRunE = rootPersistentPreRun

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path (default ~/.skylight/config)")
	rootCmd.PersistentFlags().StringVar(&email, "email", "", "Skylight account email (deprecated: use --refresh-token)")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Skylight account password (deprecated: use --refresh-token)")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Bearer access token (alternative to --refresh-token)")
	rootCmd.PersistentFlags().StringVar(&userID, "user-id", "", "User ID (used with --token)")
	rootCmd.PersistentFlags().StringVar(&frameID, "frame-id", "", "Frame ID")
	rootCmd.PersistentFlags().StringVar(&refreshToken, "refresh-token", "", "OAuth2 refresh token")
	rootCmd.PersistentFlags().StringVar(&deviceFingerprint, "device-fingerprint", "", "Device fingerprint UUID (stable per device)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", outputJSON, "Output format: json or table")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential success messages")
	registerCommonFlagCompletions()

	getCmd.Hidden = true

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(bountyCmd)
	rootCmd.AddCommand(rotationCmd)
	rootCmd.AddCommand(addonCmd)
	rootCmd.AddCommand(calendarCmd)
	rootCmd.AddCommand(choreCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(rewardCmd)
	rootCmd.AddCommand(mealCmd)
	rootCmd.AddCommand(categoryCmd)
	rootCmd.AddCommand(frameCmd)
	rootCmd.AddCommand(photoCmd)
	rootCmd.AddCommand(configCmd)
}

func rootPersistentPreRun(cmd *cobra.Command, args []string) error {
	// Load config file first (CLI flags take precedence since they're already set)
	loadConfig()

	if err := validateEnum(outputFormat, []string{outputJSON, outputTable}); err != nil {
		return err
	}

	// Skip auto-login for login command itself, help, and config subcommands
	// (config commands must work even when credentials are missing or expired).
	if cmd.Name() == loginCmd.Name() || cmd.Name() == "help" ||
		(cmd.Parent() != nil && cmd.Parent().Name() == "config") {
		return nil
	}

	// Prefer OAuth2 refresh token flow (new API).
	if handled, err := tryRefreshTokenAuth(); handled {
		return err
	}

	// Legacy: email/password via deprecated /api/sessions endpoint.
	return tryLegacyEmailPasswordAuth()
}

// tryRefreshTokenAuth attempts OAuth2 refresh-token auto-login. handled reports
// whether a refresh token was configured (and thus this flow was attempted).
func tryRefreshTokenAuth() (handled bool, err error) {
	if refreshToken == "" {
		return false, nil
	}

	fingerprint := deviceFingerprint
	if fingerprint == "" {
		fingerprint = defaultFingerprint()
	}
	c, err := lib.NewClientWithRefreshToken(refreshToken, fingerprint)
	if err != nil {
		return true, fmt.Errorf("auto-login failed: %w", err)
	}
	autoClient = c
	// Persist the rotated refresh token back to config.
	if c.RefreshToken != "" && c.RefreshToken != refreshToken {
		persistRotatedToken(c.RefreshToken, fingerprint)
		refreshToken = c.RefreshToken
	}
	return true, nil
}

func tryLegacyEmailPasswordAuth() error {
	if email == "" || password == "" || (token != "" && userID != "") {
		return nil
	}
	//nolint:staticcheck // intentional fallback for legacy callers.
	c, err := lib.NewClient(email, password)
	if err != nil {
		return fmt.Errorf("auto-login failed: %w", err)
	}
	userID = c.UserID
	token = c.APIToken
	autoClient = c
	return nil
}

func requireFrameID() error {
	if frameID == "" {
		return fmt.Errorf("--frame-id is required. Set --frame-id or SKYLIGHT_FRAME_ID; run 'skylight frame devices' to find your frame ID")
	}
	return nil
}

func getClient() (*lib.Client, error) {
	if autoClient != nil {
		return autoClient, nil
	}
	client, err := lib.NewClientWithToken(userID, token)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}
	return client, nil
}

// persistRotatedToken writes a newly rotated refresh token back to the config
// file. Auth has already succeeded at this point, so failure is non-fatal.
func persistRotatedToken(newToken, fingerprint string) {
	if err := saveConfig(map[string]string{
		"SKYLIGHT_REFRESH_TOKEN":      newToken,
		"SKYLIGHT_DEVICE_FINGERPRINT": fingerprint,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "warning: refresh token rotated but failed to persist: %v\n", err)
	}
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
