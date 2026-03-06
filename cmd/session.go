package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

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
	},
}
