package cmd

import (
	"fmt"

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
	listItemPosition  int
	listHideFromFrame bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List management commands",
	Long: `Manage to-do and shopping lists and their items on a Skylight frame.

Lists have a title, color, and kind (e.g. to_do, shopping). Use
"list info" to fetch a single list with its items included, or
"list clear-completed" to bulk-remove finished items.

  # Create a list, then add an item to it
  skylight list create --title "Groceries" --color "#F83922"
  skylight list add-item --list-id 12345678 --title "Milk"`,
}

var listListCmd = &cobra.Command{
	Use:   "all", //nolint:goconst // subcommand name, not the --resources sentinel value (resourceAll) elsewhere in the package.
	Short: "List all lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		lists, err := client.ListLists(frameID)
		if err != nil {
			return fmt.Errorf("listing lists: %w", err)
		}

		printOutput(lists)
		return nil
	},
}

var listGetCmd = &cobra.Command{
	Use:   "info",
	Short: "Get a specific list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		list, err := client.GetList(frameID, listID)
		if err != nil {
			return fmt.Errorf("getting list: %w", err)
		}

		printJSON(list)
		return nil
	},
}

var listCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.ListData{
			Title: listTitle,
			Color: listColor,
		}
		if cmd.Flags().Changed("hide-from-frame") {
			data.HideFromFrame = &listHideFromFrame
		}
		list, err := client.CreateList(frameID, data)
		if err != nil {
			return fmt.Errorf("creating list: %w", err)
		}

		printJSON(list)
		return nil
	},
}

var listDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete list %s", listID)
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteList(frameID, listID); err != nil {
			return fmt.Errorf("deleting list: %w", err)
		}

		printSuccess("List deleted successfully")
		return nil
	},
}

var listAddItemCmd = &cobra.Command{
	Use:   "add-item",
	Short: "Add an item to a list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		item, err := client.AddListItem(frameID, listID, lib.ListItemData{
			Title:    listItemTitle,
			Position: listItemPosition,
		})
		if err != nil {
			return fmt.Errorf("adding list item: %w", err)
		}

		printJSON(item)
		return nil
	},
}

var listDeleteItemCmd = &cobra.Command{
	Use:   "delete-item",
	Short: "Delete an item from a list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete item %s from list %s", listItemID, listID)
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteListItem(frameID, listID, listItemID); err != nil {
			return fmt.Errorf("deleting list item: %w", err)
		}

		printSuccess("List item deleted successfully")
		return nil
	},
}

var listUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

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
			return fmt.Errorf("updating list: %w", err)
		}

		printJSON(list)
		return nil
	},
}

var listUpdateItemCmd = &cobra.Command{
	Use:   "update-item",
	Short: "Update a list item",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.ListItemData{}
		if cmd.Flags().Changed("title") {
			data.Title = listItemTitle
		}
		if cmd.Flags().Changed("completed") {
			data.Completed = listItemCompleted
		}
		// only send position when flag is explicitly set — zero value is ambiguous
		if cmd.Flags().Changed("position") {
			data.Position = listItemPosition
		}

		item, err := client.UpdateListItem(frameID, listID, listItemID, data)
		if err != nil {
			return fmt.Errorf("updating list item: %w", err)
		}

		printJSON(item)
		return nil
	},
}

var taskBoxItemCreateCmd = &cobra.Command{
	Use:   "task-box-item",
	Short: "Create a task box item",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		item, err := client.CreateTaskBoxItem(frameID, lib.TaskBoxItemData{
			Title: listItemTitle,
		})
		if err != nil {
			return fmt.Errorf("creating task box item: %w", err)
		}

		printJSON(item)
		return nil
	},
}

var listClearCompletedCmd = &cobra.Command{
	Use:   "clear-completed",
	Short: "Delete all completed items from a list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		deleted, err := client.ClearCompletedListItems(frameID, listID)
		if err != nil {
			if deleted > 0 {
				return fmt.Errorf("deleted %d item(s) before error: %w", deleted, err)
			}
			return fmt.Errorf("clearing completed items: %w", err)
		}

		printSuccessf("Deleted %d completed item(s)\n", deleted)
		return nil
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
	markFlagRequired(taskBoxItemCreateCmd, "title")

	listGetCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	markFlagRequired(listGetCmd, "list-id")

	listCreateCmd.Flags().StringVar(&listTitle, "title", "", "List title")
	listCreateCmd.Flags().StringVar(&listColor, "color", "", "List color")
	listCreateCmd.Flags().BoolVar(&listHideFromFrame, "hide-from-frame", false, "Hide list from calendar devices")
	markFlagRequired(listCreateCmd, "title")

	listDeleteCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	markFlagRequired(listDeleteCmd, "list-id")

	listUpdateCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listUpdateCmd.Flags().StringVar(&listTitle, "title", "", "List title")
	listUpdateCmd.Flags().StringVar(&listColor, "color", "", "List color")
	listUpdateCmd.Flags().BoolVar(&listHideFromFrame, "hide-from-frame", false, "Hide list from calendar devices")
	markFlagRequired(listUpdateCmd, "list-id")

	listAddItemCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listAddItemCmd.Flags().StringVar(&listItemTitle, "title", "", "Item title")
	listAddItemCmd.Flags().IntVar(&listItemPosition, "position", 0, "Item position/order in the list")
	markFlagRequired(listAddItemCmd, "list-id")
	markFlagRequired(listAddItemCmd, "title")

	listUpdateItemCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listUpdateItemCmd.Flags().StringVar(&listItemID, "item-id", "", "Item ID")
	listUpdateItemCmd.Flags().StringVar(&listItemTitle, "title", "", "Item title")
	listUpdateItemCmd.Flags().BoolVar(&listItemCompleted, "completed", false, "Mark item as completed")
	listUpdateItemCmd.Flags().IntVar(&listItemPosition, "position", 0, "Item position/order in the list")
	markFlagRequired(listUpdateItemCmd, "list-id")
	markFlagRequired(listUpdateItemCmd, "item-id")

	listDeleteItemCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	listDeleteItemCmd.Flags().StringVar(&listItemID, "item-id", "", "Item ID")
	listDeleteItemCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	markFlagRequired(listDeleteItemCmd, "list-id")
	markFlagRequired(listDeleteItemCmd, "item-id")

	listClearCompletedCmd.Flags().StringVar(&listID, "list-id", "", "List ID")
	markFlagRequired(listClearCompletedCmd, "list-id")
}
