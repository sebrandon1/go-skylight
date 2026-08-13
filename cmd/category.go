package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	categoryID    string
	categoryName  string
	categoryColor string
)

var categoryCmd = &cobra.Command{
	Use:   subCategory,
	Short: "Category (profile/label) management commands",
	Long: `Manage family member categories (profiles/labels) on a Skylight frame.

Categories identify family members and are used as assignee IDs on chores,
rewards, and routines. Use "category list" to find the ID to pass to
--assignee-id or --category-id on other commands.

  # List all categories, then create a new one
  skylight category list --output table
  skylight category create --name "Alex" --color "#4A90D9"`,
}

var categoryListCmd = &cobra.Command{
	Use:   subList,
	Short: "List family members/categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		categories, err := client.ListCategories(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing categories: %w", err)
		}

		printOutput(categories)
		return nil
	},
}

var categoryCreateCmd = &cobra.Command{
	Use:   subCreate,
	Short: "Create a category (profile/label)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.CategoryData{
			Name: categoryName,
		}
		if cmd.Flags().Changed("color") {
			data.Color = categoryColor
		}

		category, err := client.CreateCategory(cmd.Context(), frameID, data)
		if err != nil {
			return fmt.Errorf("creating category: %w", err)
		}

		printJSON(category)
		return nil
	},
}

var categoryDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a category (profile/label)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete category %s", categoryID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete category %s?", categoryID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteCategory(cmd.Context(), frameID, categoryID); err != nil {
			return fmt.Errorf("deleting category: %w", err)
		}

		printSuccess("Category deleted successfully")
		return nil
	},
}

var categoryUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update a category (profile/label)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.CategoryData{}
		if cmd.Flags().Changed("name") {
			data.Name = categoryName
		}
		if cmd.Flags().Changed("color") {
			data.Color = categoryColor
		}

		category, err := client.UpdateCategory(cmd.Context(), frameID, categoryID, data)
		if err != nil {
			return fmt.Errorf("updating category: %w", err)
		}

		printJSON(category)
		return nil
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
	categoryDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	categoryDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(categoryDeleteCmd, "category-id")

	categoryUpdateCmd.Flags().StringVar(&categoryID, "category-id", "", "Category ID to update")
	categoryUpdateCmd.Flags().StringVar(&categoryName, "name", "", "Category name")
	categoryUpdateCmd.Flags().StringVar(&categoryColor, "color", "", "Category color (hex, e.g. #FF0000)")
	markFlagRequired(categoryUpdateCmd, "category-id")
}
