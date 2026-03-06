package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	listID        string
	listTitle     string
	listColor     string
	listItemID    string
	listItemTitle string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List management commands",
}

var listListCmd = &cobra.Command{
	Use:   "all",
	Short: "List all lists",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		lists, err := client.ListLists(frameID)
		if err != nil {
			fmt.Printf("Error listing lists: %v\n", err)
			os.Exit(1)
		}

		printJSON(lists)
	},
}

var listGetCmd = &cobra.Command{
	Use:   "info",
	Short: "Get a specific list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		list, err := client.GetList(frameID, listID)
		if err != nil {
			fmt.Printf("Error getting list: %v\n", err)
			os.Exit(1)
		}

		printJSON(list)
	},
}

var listCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		list, err := client.CreateList(frameID, lib.ListData{
			Title: listTitle,
			Color: listColor,
		})
		if err != nil {
			fmt.Printf("Error creating list: %v\n", err)
			os.Exit(1)
		}

		printJSON(list)
	},
}

var listDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.DeleteList(frameID, listID)
		if err != nil {
			fmt.Printf("Error deleting list: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("List deleted successfully")
	},
}

var listAddItemCmd = &cobra.Command{
	Use:   "add-item",
	Short: "Add an item to a list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		item, err := client.AddListItem(frameID, listID, lib.ListItemData{
			Title: listItemTitle,
		})
		if err != nil {
			fmt.Printf("Error adding list item: %v\n", err)
			os.Exit(1)
		}

		printJSON(item)
	},
}

var listDeleteItemCmd = &cobra.Command{
	Use:   "delete-item",
	Short: "Delete an item from a list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.DeleteListItem(frameID, listID, listItemID)
		if err != nil {
			fmt.Printf("Error deleting list item: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("List item deleted successfully")
	},
}

func init() {
	listCmd.AddCommand(listListCmd)
	listCmd.AddCommand(listGetCmd)
	listCmd.AddCommand(listCreateCmd)
	listCmd.AddCommand(listDeleteCmd)
	listCmd.AddCommand(listAddItemCmd)
	listCmd.AddCommand(listDeleteItemCmd)

	listGetCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listCreateCmd.Flags().StringVar(&listTitle, "title", "", "List title")
	listCreateCmd.Flags().StringVar(&listColor, "color", "", "List color")
	listDeleteCmd.Flags().StringVar(&listID, "list-id", "", "List ID")

	listAddItemCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listAddItemCmd.Flags().StringVar(&listItemTitle, "title", "", "Item title")

	listDeleteItemCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listDeleteItemCmd.Flags().StringVar(&listItemID, "item-id", "", "Item ID")
}
