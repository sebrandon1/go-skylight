package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	email   string
	password string
	token   string
	userID  string
	frameID string
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

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
