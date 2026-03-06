package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "List family members/categories",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		categories, err := client.ListCategories(frameID)
		if err != nil {
			fmt.Printf("Error listing categories: %v\n", err)
			os.Exit(1)
		}

		printJSON(categories)
	},
}
