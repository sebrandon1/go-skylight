package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	email      string
	password   string
	token      string
	userID     string
	frameID    string
	autoClient *lib.Client
)

var rootCmd = &cobra.Command{
	Use:   "skylight",
	Short: "Skylight CLI interacts with the Skylight Calendar API",
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get objects from Skylight",
}

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip auto-login for login command itself and help
		if cmd.Name() == "login" || cmd.Name() == "help" {
			return nil
		}
		// Auto-login if email/password set but no token/userID
		if email != "" && password != "" && (token == "" || userID == "") {
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

	rootCmd.PersistentFlags().StringVar(&email, "email", "", "Skylight account email")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Skylight account password")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "API token (alternative to email/password)")
	rootCmd.PersistentFlags().StringVar(&userID, "user-id", "", "User ID (required with --token)")
	rootCmd.PersistentFlags().StringVar(&frameID, "frame-id", "", "Frame ID")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(loginCmd)

	getCmd.AddCommand(calendarCmd)
	getCmd.AddCommand(choreCmd)
	getCmd.AddCommand(listCmd)
	getCmd.AddCommand(rewardCmd)
	getCmd.AddCommand(mealCmd)
	getCmd.AddCommand(categoryCmd)
	getCmd.AddCommand(frameCmd)
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

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
