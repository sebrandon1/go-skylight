package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "List family members/categories",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		categories, err := client.ListCategories(frameID)
		if err != nil {
			fmt.Printf("Error listing categories: %v\n", err)
			os.Exit(1)
		}

		printJSON(categories)
	},
}
