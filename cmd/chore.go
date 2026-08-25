package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	choreDate           string
	choreStatus         string
	choreAssigneeID     string
	choreID             string
	choreTitle          string
	choreDescription    string
	chorePoints         int
	choreAfter          string
	choreBefore         string
	choreIncludeLate    bool
	choreRecurring      bool
	choreUpForGrabs     bool
	choreWeek           string
	choreFrequency      string
	choreInterval       int
	choreRecurrenceDays []string
	choreEndDate        string
	choreRecurFrom      string
	choreSearchQuery    string
)

var choreStatuses = []string{lib.ChoreStatusPending, lib.ChoreStatusComplete, lib.ChoreStatusSkipped}

// requireWindowPair enforces that --after/--before are provided together:
// the Skylight API rejects chore queries missing either bound.
func requireWindowPair(cmd *cobra.Command) error {
	hasAfter := cmd.Flags().Changed("after")
	hasBefore := cmd.Flags().Changed("before")
	switch {
	case hasAfter && !hasBefore:
		return fmt.Errorf("--after requires --before: the API needs a full date window")
	case hasBefore && !hasAfter:
		return fmt.Errorf("--before requires --after: the API needs a full date window")
	}
	return nil
}

var choreCmd = &cobra.Command{
	Use:   subChore,
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
	Use:   subList,
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
		}{{subDate, choreDate}, {"after", choreAfter}, {"before", choreBefore}} {
			if cmd.Flags().Changed(f.name) {
				if err := validateDate(f.val); err != nil {
					return err
				}
			}
		}

		if err := requireWindowPair(cmd); err != nil {
			return err
		}

		if cmd.Flags().Changed("status") {
			if err := validateEnum(choreStatus, choreStatuses); err != nil {
				return err
			}
		}

		ctx := cmd.Context()

		if cmd.Flags().Changed("week") {
			monday, err := weekStart(choreWeek)
			if err != nil {
				return fmt.Errorf("computing week start: %w", err)
			}
			sunday := monday.AddDate(0, 0, 6)
			chores, err := client.ListChores(ctx, frameID, lib.ChoreListOptions{
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

		chores, err := client.ListChores(ctx, frameID, lib.ChoreListOptions{
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

		maybeLoadCatNames(ctx, client)
		printOutput(chores)
		return nil
	},
}

var choreGetCmd = &cobra.Command{
	Use:   subGet,
	Short: "Get a single chore by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		chore, err := client.GetChore(cmd.Context(), frameID, choreID)
		if err != nil {
			return fmt.Errorf("getting chore: %w", err)
		}

		maybeLoadCatNames(cmd.Context(), client)
		printOutput([]lib.Chore{*chore})
		return nil
	},
}

var choreCreateCmd = &cobra.Command{
	Use:   subCreate,
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
		ctx := cmd.Context()
		if choreUpForGrabs {
			chore, err = client.CreateUpForGrabsChore(ctx, frameID, data)
		} else {
			data.AssigneeID = choreAssigneeID
			data.Recurring = choreRecurring
			if cmd.Flags().Changed("frequency") {
				data.Frequency = choreFrequency
				data.Recurring = true
			}
			if cmd.Flags().Changed("interval") {
				data.Interval = choreInterval
			}
			if cmd.Flags().Changed("recurrence-days") {
				data.RecurrenceDays = choreRecurrenceDays
			}
			if cmd.Flags().Changed("end-date") {
				data.EndDate = choreEndDate
			}
			if cmd.Flags().Changed("recur-from") {
				data.RecurFrom = choreRecurFrom
			}
			chore, err = client.CreateChore(ctx, frameID, data)
		}
		if err != nil {
			return fmt.Errorf("creating chore: %w", err)
		}

		maybeLoadCatNames(ctx, client)
		printOutput([]lib.Chore{*chore})
		return nil
	},
}

var choreDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a chore",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete chore %s", choreID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete chore %s?", choreID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteChore(cmd.Context(), frameID, choreID); err != nil {
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

		if err := client.CompleteChore(cmd.Context(), frameID, choreID); err != nil {
			return fmt.Errorf("completing chore: %w", err)
		}

		printSuccess("Chore completed successfully")
		return nil
	},
}

var choreUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update a chore",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if cmd.Flags().Changed(subDate) {
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
		if cmd.Flags().Changed(subTitle) {
			data.Title = choreTitle
		}
		if cmd.Flags().Changed("status") {
			data.Status = choreStatus
		}
		if cmd.Flags().Changed(subPoints) {
			data.Points = chorePoints
		}
		if cmd.Flags().Changed("assignee-id") {
			data.AssigneeID = choreAssigneeID
		}
		if cmd.Flags().Changed(subDate) {
			data.DueDate = choreDate
		}
		if cmd.Flags().Changed("description") {
			data.Description = choreDescription
		}
		if cmd.Flags().Changed("frequency") {
			data.Frequency = choreFrequency
		}
		if cmd.Flags().Changed("interval") {
			data.Interval = choreInterval
		}
		if cmd.Flags().Changed("recurrence-days") {
			data.RecurrenceDays = choreRecurrenceDays
		}
		if cmd.Flags().Changed("end-date") {
			data.EndDate = choreEndDate
		}
		if cmd.Flags().Changed("recur-from") {
			data.RecurFrom = choreRecurFrom
		}
		if cmd.Flags().Changed("up-for-grabs") {
			data.UpForGrabs = choreUpForGrabs
		}

		ctx := cmd.Context()
		chore, err := client.UpdateChore(ctx, frameID, choreID, data)
		if err != nil {
			return fmt.Errorf("updating chore: %w", err)
		}

		maybeLoadCatNames(ctx, client)
		printOutput([]lib.Chore{*chore})
		return nil
	},
}

var choreSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search chores by name or description",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		for _, f := range []struct {
			name string
			val  string
		}{{"after", choreAfter}, {"before", choreBefore}} {
			if cmd.Flags().Changed(f.name) {
				if err := validateDate(f.val); err != nil {
					return err
				}
			}
		}

		if err := requireWindowPair(cmd); err != nil {
			return err
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

		chores, err := client.ListChores(cmd.Context(), frameID, lib.ChoreListOptions{
			Search:     choreSearchQuery,
			AssigneeID: choreAssigneeID,
			Status:     choreStatus,
			After:      choreAfter,
			Before:     choreBefore,
		})
		if err != nil {
			return fmt.Errorf("searching chores: %w", err)
		}

		printOutput(chores)
		return nil
	},
}

var choreSkipDeferUntil string

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

  # Skip and defer to a later date
  skylight chore skip --chore-id 12345678-2026-06-25 --defer-until 2026-07-01

The date suffix in the chore ID identifies the specific instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if err := validateDate(choreSkipDeferUntil); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.SkipChore(cmd.Context(), frameID, choreID, choreSkipDeferUntil); err != nil {
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

		ctx := cmd.Context()
		chore, err := client.ClaimChore(ctx, frameID, choreID, choreAssigneeID)
		if err != nil {
			return fmt.Errorf("claiming chore: %w", err)
		}

		maybeLoadCatNames(ctx, client)
		printOutput([]lib.Chore{*chore})
		return nil
	},
}

func init() {
	choreCmd.AddCommand(choreListCmd)
	choreCmd.AddCommand(choreGetCmd)
	choreCmd.AddCommand(choreCreateCmd)
	choreCmd.AddCommand(choreUpdateCmd)
	choreCmd.AddCommand(choreDeleteCmd)
	choreCmd.AddCommand(choreCompleteCmd)
	choreCmd.AddCommand(choreSkipCmd)
	choreCmd.AddCommand(choreClaimCmd)
	choreCmd.AddCommand(choreSearchCmd)

	choreListCmd.Flags().StringVar(&choreDate, subDate, "", "Date filter")
	choreListCmd.Flags().StringVar(&choreStatus, "status", "", "Status filter: pending, complete, skipped")
	choreListCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID filter")
	registerEnumFlagCompletion(choreListCmd, "status", choreStatuses...)

	choreListCmd.Flags().StringVar(&choreAfter, "after", "", "Start of date window (YYYY-MM-DD); defaults to the current month")
	choreListCmd.Flags().StringVar(&choreBefore, "before", "", "End of date window (YYYY-MM-DD); defaults to the current month")
	choreListCmd.Flags().BoolVar(&choreIncludeLate, "include-late", false, "Include late chores")
	choreListCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Only show up-for-grabs chores")
	choreListCmd.Flags().StringVar(&choreWeek, "week", "", "Show weekly calendar view; optionally specify YYYY-MM-DD to select the week")
	choreListCmd.Flags().Lookup("week").NoOptDefVal = "current"

	choreCreateCmd.Flags().StringVar(&choreTitle, subTitle, "", "Chore title")
	choreCreateCmd.Flags().StringVar(&choreDescription, "description", "", "Chore description")
	choreCreateCmd.Flags().StringVar(&choreDate, subDate, "", "Due date")
	choreCreateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreCreateCmd.Flags().IntVar(&chorePoints, subPoints, 0, "Points value")
	choreCreateCmd.Flags().BoolVar(&choreRecurring, "recurring", false, "Make chore recurring")
	choreCreateCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Make chore claimable by anyone")
	choreCreateCmd.Flags().StringVar(&choreFrequency, "frequency", "", "Recurrence frequency: daily, weekly, monthly")
	choreCreateCmd.Flags().IntVar(&choreInterval, "interval", 0, "Recurrence interval (every N periods)")
	choreCreateCmd.Flags().StringSliceVar(&choreRecurrenceDays, "recurrence-days", nil, "Days of week for weekly recurrence (e.g., mon,wed,fri)")
	choreCreateCmd.Flags().StringVar(&choreEndDate, "end-date", "", "End date for recurring chore (YYYY-MM-DD)")
	choreCreateCmd.Flags().StringVar(&choreRecurFrom, "recur-from", "", "When to anchor recurrence: scheduled or completed")
	markFlagRequired(choreCreateCmd, subTitle)

	choreUpdateCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to update")
	choreUpdateCmd.Flags().StringVar(&choreTitle, subTitle, "", "Chore title")
	choreUpdateCmd.Flags().StringVar(&choreDescription, "description", "", "Chore description")
	choreUpdateCmd.Flags().StringVar(&choreStatus, "status", "", "Chore status: pending, complete, skipped")
	choreUpdateCmd.Flags().IntVar(&chorePoints, subPoints, 0, "Points value")
	choreUpdateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreUpdateCmd.Flags().StringVar(&choreDate, subDate, "", "Due date")
	choreUpdateCmd.Flags().StringVar(&choreFrequency, "frequency", "", "Recurrence frequency: daily, weekly, monthly")
	choreUpdateCmd.Flags().IntVar(&choreInterval, "interval", 0, "Recurrence interval (every N periods)")
	choreUpdateCmd.Flags().StringSliceVar(&choreRecurrenceDays, "recurrence-days", nil, "Days of week for weekly recurrence (e.g., mon,wed,fri)")
	choreUpdateCmd.Flags().StringVar(&choreEndDate, "end-date", "", "End date for recurring chore (YYYY-MM-DD)")
	choreUpdateCmd.Flags().StringVar(&choreRecurFrom, "recur-from", "", "When to anchor recurrence: scheduled or completed")
	choreUpdateCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Make chore claimable by anyone")
	registerEnumFlagCompletion(choreUpdateCmd, "status", choreStatuses...)
	markFlagRequired(choreUpdateCmd, "chore-id")

	choreGetCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to retrieve")
	markFlagRequired(choreGetCmd, "chore-id")

	choreDeleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to delete")
	choreDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	choreDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(choreDeleteCmd, "chore-id")

	choreCompleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to complete")
	markFlagRequired(choreCompleteCmd, "chore-id")

	choreSkipCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to skip")
	choreSkipCmd.Flags().StringVar(&choreSkipDeferUntil, "defer-until", "", "Reschedule skipped instance to DATE (YYYY-MM-DD)")
	markFlagRequired(choreSkipCmd, "chore-id")

	choreClaimCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to claim")
	choreClaimCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Family member ID claiming the chore")
	markFlagRequired(choreClaimCmd, "chore-id")
	markFlagRequired(choreClaimCmd, "assignee-id")

	choreSearchCmd.Flags().StringVar(&choreSearchQuery, "query", "", "Search term to match against chore title or description")
	choreSearchCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID filter")
	choreSearchCmd.Flags().StringVar(&choreStatus, "status", "", "Status filter: pending, complete, skipped")
	choreSearchCmd.Flags().StringVar(&choreAfter, "after", "", "Start of date window (YYYY-MM-DD); defaults to the current month")
	choreSearchCmd.Flags().StringVar(&choreBefore, "before", "", "End of date window (YYYY-MM-DD); defaults to the current month")
	markFlagRequired(choreSearchCmd, "query")
	registerEnumFlagCompletion(choreSearchCmd, "status", choreStatuses...)
}
