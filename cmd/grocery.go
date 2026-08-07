package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	groceryListID    string
	groceryTitle     string
	groceryRetailer  string
	groceryItems     []string
	groceryRecipeID  string
	groceryItemID    string
	groceryItemTitle string
	groceryCompleted bool
)

var groceryCmd = &cobra.Command{
	Use:   "grocery",
	Short: "Grocery list management commands",
}

var groceryListCmd = &cobra.Command{
	Use:   subList,
	Short: "List grocery lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		all, err := client.ListLists(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing lists: %w", err)
		}

		grocery := []lib.List{}
		for _, l := range all {
			if l.Kind == lib.ListKindGrocery {
				grocery = append(grocery, l)
			}
		}

		printOutput(grocery)
		return nil
	},
}

var groceryCreateCmd = &cobra.Command{
	Use:   subCreate,
	Short: "Create a grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		list, err := client.CreateList(cmd.Context(), frameID, lib.ListData{
			Title: groceryTitle,
			Kind:  lib.ListKindGrocery,
		})
		if err != nil {
			return fmt.Errorf("creating grocery list: %w", err)
		}

		printJSON(list)
		return nil
	},
}

var groceryOrganizeCmd = &cobra.Command{
	Use:   "organize",
	Short: "Deduplicate and sort a grocery list by aisle",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.OrganizeGroceryList(cmd.Context(), frameID, groceryListID); err != nil {
			return fmt.Errorf("organizing grocery list: %w", err)
		}

		printSuccess("Grocery list organized successfully")
		return nil
	},
}

var groceryOrderCmd = &cobra.Command{
	Use:   "order",
	Short: "Send grocery list to Instacart",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		url, err := client.OrderGroceryList(cmd.Context(), frameID, groceryListID, groceryRetailer)
		if err != nil {
			return fmt.Errorf("ordering grocery list: %w", err)
		}

		if url != "" {
			fmt.Printf("Order URL: %s\n", url)
		} else {
			printSuccess("Order submitted successfully")
		}
		return nil
	},
}

var groceryShowCmd = &cobra.Command{
	Use:   subShow,
	Short: "Display the current grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		list, err := client.GetList(cmd.Context(), frameID, groceryListID)
		if err != nil {
			return fmt.Errorf("getting grocery list: %w", err)
		}

		printOutput(list)
		return nil
	},
}

var groceryAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add items to a grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		added := 0
		ctx := cmd.Context()
		for _, item := range groceryItems {
			if _, err := client.AddListItem(ctx, frameID, groceryListID, lib.ListItemData{Title: item}); err != nil {
				if added > 0 {
					return fmt.Errorf("added %d item(s) before error on %q: %w", added, item, err)
				}
				return fmt.Errorf("adding item %q: %w", item, err)
			}
			added++
		}

		printSuccessf("Added %d item(s) to grocery list\n", added)
		return nil
	},
}

var groceryAddRecipeCmd = &cobra.Command{
	Use:   "add-recipe",
	Short: "Add all ingredients from a recipe to the grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.AddRecipeToGroceryList(cmd.Context(), frameID, groceryRecipeID); err != nil {
			return fmt.Errorf("adding recipe to grocery list: %w", err)
		}

		printSuccess("Recipe added to grocery list successfully")
		return nil
	},
}

var groceryDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete grocery list %s", groceryListID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete grocery list %s?", groceryListID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteList(cmd.Context(), frameID, groceryListID); err != nil {
			return fmt.Errorf("deleting grocery list: %w", err)
		}

		printSuccess("Grocery list deleted successfully")
		return nil
	},
}

var groceryDeleteItemCmd = &cobra.Command{
	Use:   subDeleteItem,
	Short: "Delete an item from a grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete item %s from grocery list %s", groceryItemID, groceryListID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete item %s from grocery list %s?", groceryItemID, groceryListID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteListItem(cmd.Context(), frameID, groceryListID, groceryItemID); err != nil {
			return fmt.Errorf("deleting grocery list item: %w", err)
		}

		printSuccess("Grocery list item deleted successfully")
		return nil
	},
}

var groceryUpdateItemCmd = &cobra.Command{
	Use:   "update-item",
	Short: "Update an item in a grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.ListItemData{}
		if cmd.Flags().Changed(subTitle) {
			data.Title = groceryItemTitle
		}
		if cmd.Flags().Changed("completed") {
			data.Completed = groceryCompleted
		}

		item, err := client.UpdateListItem(cmd.Context(), frameID, groceryListID, groceryItemID, data)
		if err != nil {
			return fmt.Errorf("updating grocery list item: %w", err)
		}

		printJSON(item)
		return nil
	},
}

var groceryClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear completed items from a grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		n, err := client.ClearCompletedListItems(cmd.Context(), frameID, groceryListID)
		if err != nil {
			if n > 0 {
				return fmt.Errorf("deleted %d item(s) before error: %w", n, err)
			}
			return fmt.Errorf("clearing grocery list: %w", err)
		}

		printSuccessf("Cleared %d completed item(s) from grocery list\n", n)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(groceryCmd)

	groceryCmd.AddCommand(groceryListCmd)
	groceryCmd.AddCommand(groceryCreateCmd)
	groceryCmd.AddCommand(groceryShowCmd)
	groceryCmd.AddCommand(groceryAddCmd)
	groceryCmd.AddCommand(groceryAddRecipeCmd)
	groceryCmd.AddCommand(groceryClearCmd)
	groceryCmd.AddCommand(groceryOrganizeCmd)
	groceryCmd.AddCommand(groceryOrderCmd)
	groceryCmd.AddCommand(groceryDeleteCmd)
	groceryCmd.AddCommand(groceryDeleteItemCmd)
	groceryCmd.AddCommand(groceryUpdateItemCmd)

	groceryCreateCmd.Flags().StringVar(&groceryTitle, subTitle, "", "Grocery list title")
	markFlagRequired(groceryCreateCmd, subTitle)

	groceryShowCmd.Flags().StringVar(&groceryListID, "list-id", "", "List ID")
	markFlagRequired(groceryShowCmd, "list-id")

	groceryAddCmd.Flags().StringVar(&groceryListID, "list-id", "", "List ID")
	groceryAddCmd.Flags().StringSliceVar(&groceryItems, "items", nil, "Items to add (comma-separated)")
	markFlagRequired(groceryAddCmd, "list-id")
	markFlagRequired(groceryAddCmd, "items")

	groceryAddRecipeCmd.Flags().StringVar(&groceryRecipeID, "recipe-id", "", "Recipe ID")
	markFlagRequired(groceryAddRecipeCmd, "recipe-id")

	groceryClearCmd.Flags().StringVar(&groceryListID, "list-id", "", "List ID")
	markFlagRequired(groceryClearCmd, "list-id")

	groceryOrganizeCmd.Flags().StringVar(&groceryListID, "list-id", "", "List ID")
	markFlagRequired(groceryOrganizeCmd, "list-id")

	groceryOrderCmd.Flags().StringVar(&groceryListID, "list-id", "", "List ID")
	groceryOrderCmd.Flags().StringVar(&groceryRetailer, "retailer", "", "Retailer slug (e.g. costco)")
	markFlagRequired(groceryOrderCmd, "list-id")

	groceryDeleteCmd.Flags().StringVar(&groceryListID, "list-id", "", "Grocery list ID to delete")
	groceryDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	groceryDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(groceryDeleteCmd, "list-id")

	groceryDeleteItemCmd.Flags().StringVar(&groceryListID, "list-id", "", "Grocery list ID")
	groceryDeleteItemCmd.Flags().StringVar(&groceryItemID, "item-id", "", "Item ID to delete")
	groceryDeleteItemCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	groceryDeleteItemCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(groceryDeleteItemCmd, "list-id")
	markFlagRequired(groceryDeleteItemCmd, "item-id")

	groceryUpdateItemCmd.Flags().StringVar(&groceryListID, "list-id", "", "Grocery list ID")
	groceryUpdateItemCmd.Flags().StringVar(&groceryItemID, "item-id", "", "Item ID to update")
	groceryUpdateItemCmd.Flags().StringVar(&groceryItemTitle, subTitle, "", "New item title")
	groceryUpdateItemCmd.Flags().BoolVar(&groceryCompleted, "completed", false, "Mark item completed")
	markFlagRequired(groceryUpdateItemCmd, "list-id")
	markFlagRequired(groceryUpdateItemCmd, "item-id")
}
