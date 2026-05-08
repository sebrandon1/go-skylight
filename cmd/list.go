package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	listID            string
	listTitle         string
	listColor         string
	listItemID        string
	listItemTitle     string
	listItemCompleted bool
	listHideFromFrame bool
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

		client := getClient()

		lists, err := client.ListLists(frameID)
		if err != nil {
			fatal("listing lists", err)
		}

		printOutput(lists)
	},
}

var listGetCmd = &cobra.Command{
	Use:   "info",
	Short: "Get a specific list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		list, err := client.GetList(frameID, listID)
		if err != nil {
			fatal("getting list", err)
		}

		printJSON(list)
	},
}

var listCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.ListData{
			Title: listTitle,
			Color: listColor,
		}
		if cmd.Flags().Changed("hide-from-frame") {
			data.HideFromFrame = &listHideFromFrame
		}
		list, err := client.CreateList(frameID, data)
		if err != nil {
			fatal("creating list", err)
		}

		printJSON(list)
	},
}

var listDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.DeleteList(frameID, listID)
		if err != nil {
			fatal("deleting list", err)
		}

		fmt.Println("List deleted successfully")
	},
}

var listAddItemCmd = &cobra.Command{
	Use:   "add-item",
	Short: "Add an item to a list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		item, err := client.AddListItem(frameID, listID, lib.ListItemData{
			Title: listItemTitle,
		})
		if err != nil {
			fatal("adding list item", err)
		}

		printJSON(item)
	},
}

var listDeleteItemCmd = &cobra.Command{
	Use:   "delete-item",
	Short: "Delete an item from a list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.DeleteListItem(frameID, listID, listItemID)
		if err != nil {
			fatal("deleting list item", err)
		}

		fmt.Println("List item deleted successfully")
	},
}

var listUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.ListData{}
		if cmd.Flags().Changed("title") {
			data.Title = listTitle
		}
		if cmd.Flags().Changed("color") {
			data.Color = listColor
		}
		if cmd.Flags().Changed("hide-from-frame") {
			data.HideFromFrame = &listHideFromFrame
		}

		list, err := client.UpdateList(frameID, listID, data)
		if err != nil {
			fatal("updating list", err)
		}

		printJSON(list)
	},
}

var listUpdateItemCmd = &cobra.Command{
	Use:   "update-item",
	Short: "Update a list item",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.ListItemData{}
		if cmd.Flags().Changed("title") {
			data.Title = listItemTitle
		}
		if cmd.Flags().Changed("completed") {
			data.Completed = listItemCompleted
		}

		item, err := client.UpdateListItem(frameID, listID, listItemID, data)
		if err != nil {
			fatal("updating list item", err)
		}

		printJSON(item)
	},
}

var taskBoxItemCreateCmd = &cobra.Command{
	Use:   "task-box-item",
	Short: "Create a task box item",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		item, err := client.CreateTaskBoxItem(frameID, lib.TaskBoxItemData{
			Title: listItemTitle,
		})
		if err != nil {
			fatal("creating task box item", err)
		}

		printJSON(item)
	},
}

var listClearCompletedCmd = &cobra.Command{
	Use:   "clear-completed",
	Short: "Delete all completed items from a list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		deleted, err := client.ClearCompletedListItems(frameID, listID)
		if err != nil {
			if deleted > 0 {
				fmt.Fprintf(os.Stderr, "Deleted %d item(s) before error: %v\n", deleted, err)
			} else {
				fmt.Fprintf(os.Stderr, "Error clearing completed items: %v\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("Deleted %d completed item(s)\n", deleted)
	},
}

func init() {
	listCmd.AddCommand(listListCmd)
	listCmd.AddCommand(listGetCmd)
	listCmd.AddCommand(listCreateCmd)
	listCmd.AddCommand(listUpdateCmd)
	listCmd.AddCommand(listDeleteCmd)
	listCmd.AddCommand(listAddItemCmd)
	listCmd.AddCommand(listUpdateItemCmd)
	listCmd.AddCommand(listDeleteItemCmd)
	listCmd.AddCommand(listClearCompletedCmd)
	listCmd.AddCommand(taskBoxItemCreateCmd)

	taskBoxItemCreateCmd.Flags().StringVar(&listItemTitle, "title", "", "Task box item title")
	taskBoxItemCreateCmd.MarkFlagRequired("title") //nolint:errcheck

	listGetCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listGetCmd.MarkFlagRequired("list-id") //nolint:errcheck

	listCreateCmd.Flags().StringVar(&listTitle, "title", "", "List title")
	listCreateCmd.Flags().StringVar(&listColor, "color", "", "List color")
	listCreateCmd.Flags().BoolVar(&listHideFromFrame, "hide-from-frame", false, "Hide list from calendar devices")
	listCreateCmd.MarkFlagRequired("title") //nolint:errcheck

	listDeleteCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listDeleteCmd.MarkFlagRequired("list-id") //nolint:errcheck

	listUpdateCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listUpdateCmd.Flags().StringVar(&listTitle, "title", "", "List title")
	listUpdateCmd.Flags().StringVar(&listColor, "color", "", "List color")
	listUpdateCmd.Flags().BoolVar(&listHideFromFrame, "hide-from-frame", false, "Hide list from calendar devices")
	listUpdateCmd.MarkFlagRequired("list-id") //nolint:errcheck

	listAddItemCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listAddItemCmd.Flags().StringVar(&listItemTitle, "title", "", "Item title")
	listAddItemCmd.MarkFlagRequired("list-id") //nolint:errcheck
	listAddItemCmd.MarkFlagRequired("title")   //nolint:errcheck

	listUpdateItemCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listUpdateItemCmd.Flags().StringVar(&listItemID, "item-id", "", "Item ID")
	listUpdateItemCmd.Flags().StringVar(&listItemTitle, "title", "", "Item title")
	listUpdateItemCmd.Flags().BoolVar(&listItemCompleted, "completed", false, "Mark item as completed")
	listUpdateItemCmd.MarkFlagRequired("list-id") //nolint:errcheck
	listUpdateItemCmd.MarkFlagRequired("item-id") //nolint:errcheck

	listDeleteItemCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listDeleteItemCmd.Flags().StringVar(&listItemID, "item-id", "", "Item ID")
	listDeleteItemCmd.MarkFlagRequired("list-id") //nolint:errcheck
	listDeleteItemCmd.MarkFlagRequired("item-id") //nolint:errcheck

	listClearCompletedCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listClearCompletedCmd.MarkFlagRequired("list-id") //nolint:errcheck
}
