package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	choreDate        string
	choreStatus      string
	choreAssigneeID  string
	choreID          string
	choreTitle       string
	chorePoints      int
	choreAfter       string
	choreBefore      string
	choreIncludeLate bool
	choreRecurring   bool
)

var choreCmd = &cobra.Command{
	Use:   "chore",
	Short: "Chore management commands",
}

var choreListCmd = &cobra.Command{
	Use:   "list",
	Short: "List chores",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		chores, err := client.ListChores(frameID, lib.ChoreListOptions{
			Date:        choreDate,
			Status:      choreStatus,
			AssigneeID:  choreAssigneeID,
			After:       choreAfter,
			Before:      choreBefore,
			IncludeLate: choreIncludeLate,
		})
		if err != nil {
			fmt.Printf("Error listing chores: %v\n", err)
			os.Exit(1)
		}

		printJSON(chores)
	},
}

var choreCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a chore",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		chore, err := client.CreateChore(frameID, lib.ChoreData{
			Title:      choreTitle,
			DueDate:    choreDate,
			AssigneeID: choreAssigneeID,
			Points:     chorePoints,
			Recurring:  choreRecurring,
		})
		if err != nil {
			fmt.Printf("Error creating chore: %v\n", err)
			os.Exit(1)
		}

		printJSON(chore)
	},
}

var choreDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a chore",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.DeleteChore(frameID, choreID)
		if err != nil {
			fmt.Printf("Error deleting chore: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Chore deleted successfully")
	},
}

func init() {
	choreCmd.AddCommand(choreListCmd)
	choreCmd.AddCommand(choreCreateCmd)
	choreCmd.AddCommand(choreDeleteCmd)

	choreListCmd.Flags().StringVar(&choreDate, "date", "", "Date filter")
	choreListCmd.Flags().StringVar(&choreStatus, "status", "", "Status filter")
	choreListCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID filter")

	choreListCmd.Flags().StringVar(&choreAfter, "after", "", "After date filter")
	choreListCmd.Flags().StringVar(&choreBefore, "before", "", "Before date filter")
	choreListCmd.Flags().BoolVar(&choreIncludeLate, "include-late", false, "Include late chores")

	choreCreateCmd.Flags().StringVar(&choreTitle, "title", "", "Chore title")
	choreCreateCmd.Flags().StringVar(&choreDate, "date", "", "Due date")
	choreCreateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreCreateCmd.Flags().IntVar(&chorePoints, "points", 0, "Points value")
	choreCreateCmd.Flags().BoolVar(&choreRecurring, "recurring", false, "Make chore recurring")

	choreDeleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to delete")
}
