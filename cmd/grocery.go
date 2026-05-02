package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	groceryListID   string
	groceryTitle    string
	groceryRetailer string
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
			fmt.Fprintf(os.Stderr, "Error listing lists: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error creating grocery list: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error organizing grocery list: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error ordering grocery list: %v\n", err)
			os.Exit(1)
		}

		if url != "" {
			fmt.Printf("Order URL: %s\n", url)
		} else {
			fmt.Println("Order submitted successfully")
		}
	},
}

func init() {
	rootCmd.AddCommand(groceryCmd)

	groceryCmd.AddCommand(groceryListCmd)
	groceryCmd.AddCommand(groceryCreateCmd)
	groceryCmd.AddCommand(groceryOrganizeCmd)
	groceryCmd.AddCommand(groceryOrderCmd)

	groceryCreateCmd.Flags().StringVar(&groceryTitle, "title", "", "Grocery list title")
	groceryCreateCmd.MarkFlagRequired("title") //nolint:errcheck

	groceryOrganizeCmd.Flags().StringVar(&groceryListID, "list-id", "", "List ID")
	groceryOrganizeCmd.MarkFlagRequired("list-id") //nolint:errcheck

	groceryOrderCmd.Flags().StringVar(&groceryListID, "list-id", "", "List ID")
	groceryOrderCmd.Flags().StringVar(&groceryRetailer, "retailer", "", "Retailer slug (e.g. costco)")
	groceryOrderCmd.MarkFlagRequired("list-id") //nolint:errcheck
}
