package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	groceryListID   string
	groceryTitle    string
	groceryRetailer string
	groceryItems    []string
	groceryRecipeID string
)

var groceryCmd = &cobra.Command{
	Use:   "grocery",
	Short: "Grocery list management commands",
}

var groceryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List grocery lists",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		all, err := client.ListLists(frameID)
		if err != nil {
			fatal("listing lists", err)
		}

		grocery := []lib.List{}
		for _, l := range all {
			if l.Kind == lib.ListKindGrocery {
				grocery = append(grocery, l)
			}
		}

		printOutput(grocery)
	},
}

var groceryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a grocery list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		list, err := client.CreateList(frameID, lib.ListData{
			Title: groceryTitle,
			Kind:  lib.ListKindGrocery,
		})
		if err != nil {
			fatal("creating grocery list", err)
		}

		printJSON(list)
	},
}

var groceryOrganizeCmd = &cobra.Command{
	Use:   "organize",
	Short: "Deduplicate and sort a grocery list by aisle",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.OrganizeGroceryList(frameID, groceryListID); err != nil {
			fatal("organizing grocery list", err)
		}

		fmt.Println("Grocery list organized successfully")
	},
}

var groceryOrderCmd = &cobra.Command{
	Use:   "order",
	Short: "Send grocery list to Instacart",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		url, err := client.OrderGroceryList(frameID, groceryListID, groceryRetailer)
		if err != nil {
			fatal("ordering grocery list", err)
		}

		if url != "" {
			fmt.Printf("Order URL: %s\n", url)
		} else {
			fmt.Println("Order submitted successfully")
		}
	},
}

var groceryShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display the current grocery list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		list, err := client.GetList(frameID, groceryListID)
		if err != nil {
			fatal("getting grocery list", err)
		}

		printOutput(list)
	},
}

var groceryAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add items to a grocery list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		for _, item := range groceryItems {
			if _, err := client.AddListItem(frameID, groceryListID, lib.ListItemData{Title: item}); err != nil {
				fatal(fmt.Sprintf("adding item %q", item), err)
			}
		}

		fmt.Printf("Added %d item(s) to grocery list\n", len(groceryItems))
	},
}

var groceryAddRecipeCmd = &cobra.Command{
	Use:   "add-recipe",
	Short: "Add all ingredients from a recipe to the grocery list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.AddRecipeToGroceryList(frameID, groceryRecipeID); err != nil {
			fatal("adding recipe to grocery list", err)
		}

		fmt.Println("Recipe added to grocery list successfully")
	},
}

var groceryClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear completed items from a grocery list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		n, err := client.ClearCompletedListItems(frameID, groceryListID)
		if err != nil {
			fatal("clearing grocery list", err)
		}

		fmt.Printf("Cleared %d completed item(s) from grocery list\n", n)
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

	groceryCreateCmd.Flags().StringVar(&groceryTitle, "title", "", "Grocery list title")
	markFlagRequired(groceryCreateCmd, "title")

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
}
