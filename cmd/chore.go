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
	choreUpForGrabs  bool
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
			UpForGrabs:  choreUpForGrabs,
		})
		if err != nil {
			fmt.Printf("Error listing chores: %v\n", err)
			os.Exit(1)
		}

		printOutput(chores)
	},
}

var choreCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a chore",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		var chore *lib.Chore
		var err error
		if choreUpForGrabs {
			chore, err = client.CreateUpForGrabsChore(frameID, lib.ChoreData{
				Title:   choreTitle,
				DueDate: choreDate,
				Points:  chorePoints,
			})
		} else {
			chore, err = client.CreateChore(frameID, lib.ChoreData{
				Title:      choreTitle,
				DueDate:    choreDate,
				AssigneeID: choreAssigneeID,
				Points:     chorePoints,
				Recurring:  choreRecurring,
			})
		}
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

var choreCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Mark a chore as completed",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.CompleteChore(frameID, choreID); err != nil {
			fmt.Printf("Error completing chore: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Chore completed successfully")
	},
}

var choreUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a chore",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.ChoreData{}
		if cmd.Flags().Changed("title") {
			data.Title = choreTitle
		}
		if cmd.Flags().Changed("status") {
			data.Status = choreStatus
		}
		if cmd.Flags().Changed("points") {
			data.Points = chorePoints
		}
		if cmd.Flags().Changed("assignee-id") {
			data.AssigneeID = choreAssigneeID
		}
		if cmd.Flags().Changed("date") {
			data.DueDate = choreDate
		}

		chore, err := client.UpdateChore(frameID, choreID, data)
		if err != nil {
			fmt.Printf("Error updating chore: %v\n", err)
			os.Exit(1)
		}

		printJSON(chore)
	},
}

var choreSkipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Skip a recurring chore instance",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.SkipChore(frameID, choreID); err != nil {
			fmt.Printf("Error skipping chore: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Chore skipped successfully")
	},
}

var choreClaimCmd = &cobra.Command{
	Use:   "claim",
	Short: "Claim an up-for-grabs chore",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		chore, err := client.ClaimChore(frameID, choreID, choreAssigneeID)
		if err != nil {
			fmt.Printf("Error claiming chore: %v\n", err)
			os.Exit(1)
		}

		printJSON(chore)
	},
}

func init() {
	choreCmd.AddCommand(choreListCmd)
	choreCmd.AddCommand(choreCreateCmd)
	choreCmd.AddCommand(choreUpdateCmd)
	choreCmd.AddCommand(choreDeleteCmd)
	choreCmd.AddCommand(choreCompleteCmd)
	choreCmd.AddCommand(choreSkipCmd)
	choreCmd.AddCommand(choreClaimCmd)

	choreListCmd.Flags().StringVar(&choreDate, "date", "", "Date filter")
	choreListCmd.Flags().StringVar(&choreStatus, "status", "", "Status filter")
	choreListCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID filter")

	choreListCmd.Flags().StringVar(&choreAfter, "after", "", "After date filter")
	choreListCmd.Flags().StringVar(&choreBefore, "before", "", "Before date filter")
	choreListCmd.Flags().BoolVar(&choreIncludeLate, "include-late", false, "Include late chores")
	choreListCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Only show up-for-grabs chores")

	choreCreateCmd.Flags().StringVar(&choreTitle, "title", "", "Chore title")
	choreCreateCmd.Flags().StringVar(&choreDate, "date", "", "Due date")
	choreCreateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreCreateCmd.Flags().IntVar(&chorePoints, "points", 0, "Points value")
	choreCreateCmd.Flags().BoolVar(&choreRecurring, "recurring", false, "Make chore recurring")
	choreCreateCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Make chore claimable by anyone")
	choreCreateCmd.MarkFlagRequired("title") //nolint:errcheck

	choreUpdateCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to update")
	choreUpdateCmd.Flags().StringVar(&choreTitle, "title", "", "Chore title")
	choreUpdateCmd.Flags().StringVar(&choreStatus, "status", "", "Chore status")
	choreUpdateCmd.Flags().IntVar(&chorePoints, "points", 0, "Points value")
	choreUpdateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreUpdateCmd.Flags().StringVar(&choreDate, "date", "", "Due date")

	choreDeleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to delete")

	choreCompleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to complete")
	choreCompleteCmd.MarkFlagRequired("chore-id") //nolint:errcheck

	choreSkipCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to skip")
	choreSkipCmd.MarkFlagRequired("chore-id") //nolint:errcheck

	choreClaimCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to claim")
	choreClaimCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Family member ID claiming the chore")
	choreClaimCmd.MarkFlagRequired("chore-id")    //nolint:errcheck
	choreClaimCmd.MarkFlagRequired("assignee-id") //nolint:errcheck
}
