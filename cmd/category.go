package cmd

import (
	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	categoryID    string
	categoryName  string
	categoryColor string
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Category (profile/label) management commands",
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List family members/categories",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		categories, err := client.ListCategories(frameID)
		if err != nil {
			fatal("listing categories", err)
		}

		printOutput(categories)
	},
}

var categoryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a category (profile/label)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.CategoryData{
			Name: categoryName,
		}
		if cmd.Flags().Changed("color") {
			data.Color = categoryColor
		}

		category, err := client.CreateCategory(frameID, data)
		if err != nil {
			fatal("creating category", err)
		}

		printJSON(category)
	},
}

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a category (profile/label)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.DeleteCategory(frameID, categoryID)
		if err != nil {
			fatal("deleting category", err)
		}

		printSuccess("Category deleted successfully")
	},
}

var categoryUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a category (profile/label)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.CategoryData{}
		if cmd.Flags().Changed("name") {
			data.Name = categoryName
		}
		if cmd.Flags().Changed("color") {
			data.Color = categoryColor
		}

		category, err := client.UpdateCategory(frameID, categoryID, data)
		if err != nil {
			fatal("updating category", err)
		}

		printJSON(category)
	},
}

func init() {
	categoryCmd.AddCommand(categoryListCmd)
	categoryCmd.AddCommand(categoryCreateCmd)
	categoryCmd.AddCommand(categoryDeleteCmd)
	categoryCmd.AddCommand(categoryUpdateCmd)

	categoryCreateCmd.Flags().StringVar(&categoryName, "name", "", "Category name")
	categoryCreateCmd.Flags().StringVar(&categoryColor, "color", "", "Category color (hex, e.g. #FF0000)")
	markFlagRequired(categoryCreateCmd, "name")

	categoryDeleteCmd.Flags().StringVar(&categoryID, "category-id", "", "Category ID")
	markFlagRequired(categoryDeleteCmd, "category-id")

	categoryUpdateCmd.Flags().StringVar(&categoryID, "category-id", "", "Category ID to update")
	categoryUpdateCmd.Flags().StringVar(&categoryName, "name", "", "Category name")
	categoryUpdateCmd.Flags().StringVar(&categoryColor, "color", "", "Category color (hex, e.g. #FF0000)")
	markFlagRequired(categoryUpdateCmd, "category-id")
}
