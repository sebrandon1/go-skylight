package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var saveCredentials bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Skylight and get API credentials",
	Run: func(cmd *cobra.Command, args []string) {
		if email == "" || password == "" {
			fmt.Println("Error: --email and --password are required for login")
			os.Exit(1)
		}

		client, err := lib.NewClient(email, password)
		if err != nil {
			fmt.Printf("Error logging in: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Login successful!\n")
		fmt.Printf("User ID: %s\n", client.UserID)
		fmt.Printf("API Token: %s\n", client.APIToken)

		if saveCredentials {
			values := map[string]string{
				"SKYLIGHT_EMAIL":    email,
				"SKYLIGHT_PASSWORD": password,
				"SKYLIGHT_TOKEN":    client.APIToken,
				"SKYLIGHT_USER_ID":  client.UserID,
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
	},
}

func init() {
	loginCmd.Flags().BoolVar(&saveCredentials, "save", false, "Save credentials to config file")
}
