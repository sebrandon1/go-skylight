package cmd

import (
	"crypto/rand"
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var saveCredentials bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Skylight and save OAuth2 credentials",
	Long: `Performs a headless OAuth2 login using email and password, then saves
the refresh token and device fingerprint to the config file. Use --save to
persist the credentials so subsequent commands authenticate automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if email == "" || password == "" {
			return fmt.Errorf("--email and --password are required for login")
		}

		// Generate a stable device fingerprint if not provided.
		fingerprint := deviceFingerprint
		if fingerprint == "" {
			fingerprint = newUUID()
		}

		fmt.Printf("Logging in as %s...\n", email)
		tok, err := lib.LoginHeadless(email, password, fingerprint)
		if err != nil {
			return fmt.Errorf("logging in: %w", err)
		}

		printSuccess("Login successful!")
		fmt.Printf("Access Token:  %s\n", tok.AccessToken)
		fmt.Printf("Refresh Token: %s\n", tok.RefreshToken)
		fmt.Printf("Fingerprint:   %s\n", fingerprint)
		fmt.Printf("Expires In:    %d seconds\n", tok.ExpiresIn)

		if saveCredentials {
			values := map[string]string{
				"SKYLIGHT_REFRESH_TOKEN":      tok.RefreshToken,
				"SKYLIGHT_DEVICE_FINGERPRINT": fingerprint,
			}
			if frameID != "" {
				values["SKYLIGHT_FRAME_ID"] = frameID
			}
			if err := saveConfig(values); err != nil {
				fmt.Printf("Warning: could not save config: %v\n", err)
			} else {
				path := configPath
				if path == "" {
					path = defaultConfigPath()
				}
				fmt.Printf("Credentials saved to %s\n", path)
			}
		}
		return nil
	},
}

func init() {
	loginCmd.Flags().BoolVar(&saveCredentials, "save", false, "Save refresh token and fingerprint to config file")
}

// newUUID generates a random UUID v4 string.
func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
