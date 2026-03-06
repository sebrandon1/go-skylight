package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	choreDate       string
	choreStatus     string
	choreAssigneeID string
	choreID         string
	choreTitle      string
	chorePoints     int
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

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		chores, err := client.ListChores(frameID, choreDate, choreStatus, choreAssigneeID)
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

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		chore, err := client.CreateChore(frameID, lib.ChoreData{
			Title:      choreTitle,
			DueDate:    choreDate,
			AssigneeID: choreAssigneeID,
			Points:     chorePoints,
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

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.DeleteChore(frameID, choreID)
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

	choreCreateCmd.Flags().StringVar(&choreTitle, "title", "", "Chore title")
	choreCreateCmd.Flags().StringVar(&choreDate, "date", "", "Due date")
	choreCreateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreCreateCmd.Flags().IntVar(&chorePoints, "points", 0, "Points value")

	choreDeleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to delete")
}
