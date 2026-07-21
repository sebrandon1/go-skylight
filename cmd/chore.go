package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	choreDate        string
	choreStatus      string
	choreAssigneeID  string
	choreID          string
	choreTitle       string
	choreDescription string
	chorePoints      int
	choreAfter       string
	choreBefore      string
	choreIncludeLate bool
	choreRecurring   bool
	choreUpForGrabs  bool
	choreWeek        string
)

var choreStatuses = []string{lib.ChoreStatusPending, lib.ChoreStatusComplete, lib.ChoreStatusSkipped}

var choreCmd = &cobra.Command{
	Use:   "chore",
	Short: "Chore management commands",
	Long: `Create, list, update, complete, delete, skip, and claim chores on a Skylight frame.

Chores can be assigned to a family member or left up-for-grabs for
anyone to claim. Prefer --after/--before over --date to filter chore
list by date range, and --status to filter by pending/complete/skipped.

  # Find today's pending chores, then complete one
  skylight chore list --after 2026-06-05 --before 2026-06-05 --status pending
  skylight chore complete --chore-id 12345678`,
}

var choreListCmd = &cobra.Command{
	Use:   "list",
	Short: "List chores",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		for _, f := range []struct {
			name string
			val  string
		}{{"date", choreDate}, {"after", choreAfter}, {"before", choreBefore}} {
			if cmd.Flags().Changed(f.name) {
				if err := validateDate(f.val); err != nil {
					return err
				}
			}
		}

		if cmd.Flags().Changed("status") {
			if err := validateEnum(choreStatus, choreStatuses); err != nil {
				return err
			}
		}

		if cmd.Flags().Changed("week") {
			monday, err := weekStart(choreWeek)
			if err != nil {
				return fmt.Errorf("computing week start: %w", err)
			}
			sunday := monday.AddDate(0, 0, 6)
			chores, err := client.ListChores(frameID, lib.ChoreListOptions{
				After:       monday.Format(lib.DateFormat),
				Before:      sunday.Format(lib.DateFormat),
				IncludeLate: true,
			})
			if err != nil {
				return fmt.Errorf("listing chores: %w", err)
			}
			days := buildWeeklyView(chores, monday)
			printOutput(days)
			return nil
		}

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
			return fmt.Errorf("listing chores: %w", err)
		}

		maybeLoadCatNames(client)
		printOutput(chores)
		return nil
	},
}

var choreCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a chore",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if err := validateDate(choreDate); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.ChoreData{
			Title:       choreTitle,
			Description: choreDescription,
			DueDate:     choreDate,
			Points:      chorePoints,
		}
		var chore *lib.Chore
		if choreUpForGrabs {
			chore, err = client.CreateUpForGrabsChore(frameID, data)
		} else {
			data.AssigneeID = choreAssigneeID
			data.Recurring = choreRecurring
			chore, err = client.CreateChore(frameID, data)
		}
		if err != nil {
			return fmt.Errorf("creating chore: %w", err)
		}

		printJSON(chore)
		return nil
	},
}

var choreDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a chore",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete chore %s", choreID)
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteChore(frameID, choreID); err != nil {
			return fmt.Errorf("deleting chore: %w", err)
		}

		printSuccess("Chore deleted successfully")
		return nil
	},
}

var choreCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Mark a chore as completed",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.CompleteChore(frameID, choreID); err != nil {
			return fmt.Errorf("completing chore: %w", err)
		}

		printSuccess("Chore completed successfully")
		return nil
	},
}

var choreUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a chore",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if cmd.Flags().Changed("date") {
			if err := validateDate(choreDate); err != nil {
				return err
			}
		}

		if cmd.Flags().Changed("status") {
			if err := validateEnum(choreStatus, choreStatuses); err != nil {
				return err
			}
		}

		client, err := getClient()
		if err != nil {
			return err
		}

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
		if cmd.Flags().Changed("description") {
			data.Description = choreDescription
		}

		chore, err := client.UpdateChore(frameID, choreID, data)
		if err != nil {
			return fmt.Errorf("updating chore: %w", err)
		}

		printJSON(chore)
		return nil
	},
}

var choreSkipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Skip a recurring chore instance",
	Long: `Skip a single instance of a recurring chore without deleting it.

The chore continues to repeat on its normal schedule — only the specified
instance is marked skipped. The API requires the chore to be both assigned
and recurring; up-for-grabs or non-recurring chores cannot be skipped.

  # Skip today's instance of an assigned recurring chore
  skylight chore list --status pending      # find the chore ID
  skylight chore skip --chore-id 12345678-2026-06-25

The date suffix in the chore ID identifies the specific instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.SkipChore(frameID, choreID); err != nil {
			return fmt.Errorf("skipping chore: %w", err)
		}

		printSuccess("Chore skipped successfully")
		return nil
	},
}

var choreClaimCmd = &cobra.Command{
	Use:   "claim",
	Short: "Claim an up-for-grabs chore",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		chore, err := client.ClaimChore(frameID, choreID, choreAssigneeID)
		if err != nil {
			return fmt.Errorf("claiming chore: %w", err)
		}

		printJSON(chore)
		return nil
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
	choreListCmd.Flags().StringVar(&choreStatus, "status", "", "Status filter: pending, complete, skipped")
	choreListCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID filter")
	registerEnumFlagCompletion(choreListCmd, "status", choreStatuses...)

	choreListCmd.Flags().StringVar(&choreAfter, "after", "", "After date filter")
	choreListCmd.Flags().StringVar(&choreBefore, "before", "", "Before date filter")
	choreListCmd.Flags().BoolVar(&choreIncludeLate, "include-late", false, "Include late chores")
	choreListCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Only show up-for-grabs chores")
	choreListCmd.Flags().StringVar(&choreWeek, "week", "", "Show weekly calendar view; optionally specify YYYY-MM-DD to select the week")
	choreListCmd.Flags().Lookup("week").NoOptDefVal = "current"

	choreCreateCmd.Flags().StringVar(&choreTitle, "title", "", "Chore title")
	choreCreateCmd.Flags().StringVar(&choreDescription, "description", "", "Chore description")
	choreCreateCmd.Flags().StringVar(&choreDate, "date", "", "Due date")
	choreCreateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreCreateCmd.Flags().IntVar(&chorePoints, "points", 0, "Points value")
	choreCreateCmd.Flags().BoolVar(&choreRecurring, "recurring", false, "Make chore recurring")
	choreCreateCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Make chore claimable by anyone")
	markFlagRequired(choreCreateCmd, "title")

	choreUpdateCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to update")
	choreUpdateCmd.Flags().StringVar(&choreTitle, "title", "", "Chore title")
	choreUpdateCmd.Flags().StringVar(&choreDescription, "description", "", "Chore description")
	choreUpdateCmd.Flags().StringVar(&choreStatus, "status", "", "Chore status: pending, complete, skipped")
	choreUpdateCmd.Flags().IntVar(&chorePoints, "points", 0, "Points value")
	choreUpdateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreUpdateCmd.Flags().StringVar(&choreDate, "date", "", "Due date")
	registerEnumFlagCompletion(choreUpdateCmd, "status", choreStatuses...)
	markFlagRequired(choreUpdateCmd, "chore-id")

	choreDeleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to delete")
	choreDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	markFlagRequired(choreDeleteCmd, "chore-id")

	choreCompleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to complete")
	markFlagRequired(choreCompleteCmd, "chore-id")

	choreSkipCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to skip")
	markFlagRequired(choreSkipCmd, "chore-id")

	choreClaimCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to claim")
	choreClaimCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Family member ID claiming the chore")
	markFlagRequired(choreClaimCmd, "chore-id")
	markFlagRequired(choreClaimCmd, "assignee-id")
}
