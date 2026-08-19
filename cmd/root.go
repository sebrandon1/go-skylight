package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	Use:   subGet,
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
	// Default is "" (not "json") so loadConfig() can distinguish "flag set" from "unset".
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format: json or table")
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

	if outputFormat == "" {
		outputFormat = outputJSON
	}

	if err := validateEnum(outputFormat, []string{outputJSON, outputTable}); err != nil {
		return err
	}

	// Skip auto-login for login command itself, help, and config subcommands
	// (config commands must work even when credentials are missing or expired).
	if cmd.Name() == loginCmd.Name() || cmd.Name() == "help" ||
		(cmd.Parent() != nil && cmd.Parent().Name() == subConfig) {
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
	generated := fingerprint == ""
	if generated {
		fingerprint = newUUID()
	}
	c, err := lib.NewClientWithRefreshToken(refreshToken, fingerprint)
	if err != nil {
		return true, fmt.Errorf("auto-login failed: %w", err)
	}
	autoClient = c
	tokenRotated := c.RefreshToken != "" && c.RefreshToken != refreshToken
	if tokenRotated || generated {
		toSave := map[string]string{cfgDeviceFingerprint: fingerprint}
		if tokenRotated {
			toSave[cfgRefreshToken] = c.RefreshToken
			refreshToken = c.RefreshToken
		}
		if err := saveConfig(toSave); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to persist credentials: %v\n", err)
		}
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

// Execute executes the root command.
func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return rootCmd.ExecuteContext(ctx)
}
