package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	choreDate            string
	choreStatus          string
	choreAssigneeID      string
	choreID              string
	choreTitle           string
	choreDescription     string
	chorePoints          int
	choreAfter           string
	choreBefore          string
	choreIncludeLate     bool
	choreRecurring       bool
	choreUpForGrabs      bool
	choreWeek            string
	choreFrequency       string
	choreInterval        int
	choreRecurrenceDays  []string
	choreEndDate         string
	choreRecurFrom       string
	choreSearchQuery     string
	choreDeleteRecurring bool
	choreAllLimit        int
	choreAllUpForGrabs   bool
	choreDueTime         string
	choreAssigneeIDs     []string
	choreApplyToProfiles string
)

const (
	flagDueTime  = "due-time"
	dueTimeUsage = "Due time (HH:MM, 24-hour); due notifications only fire for tasks with a due time (see chore notifications)"
)

var choreApplyScopes = []string{lib.ApplyToProfilesAll, lib.ApplyToProfilesOne}

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

const choreAllLong = `Show the "All Tasks" view: chores and routines, one entry per recurring
series, bucketed into late, today, today_timed, any_day and future.

  skylight chore all --limit 50
  skylight task all --include-up-for-grabs=false -o table

--limit caps the total across all buckets (0 uses the server default).`

func runChoreAll(cmd *cobra.Command, _ []string) error {
	if err := requireFrameID(); err != nil {
		return err
	}
	if choreAllLimit < 0 {
		return fmt.Errorf("--limit must be >= 0")
	}
	client, err := getClient()
	if err != nil {
		return err
	}
	tasks, err := client.ListAllTasks(cmd.Context(), frameID, lib.AllTasksOptions{
		Limit:             choreAllLimit,
		IncludeUpForGrabs: choreAllUpForGrabs,
	})
	if err != nil {
		return fmt.Errorf("listing all tasks: %w", err)
	}
	maybeLoadCatNames(cmd.Context(), client)
	printOutput(tasks)
	return nil
}

var choreAllCmd = &cobra.Command{
	Use:     resourceAll,
	Aliases: []string{"all-tasks"},
	Short:   "Show the consolidated All Tasks view (one entry per series)",
	Long:    choreAllLong + "\nAlso available as `skylight task all`.",
	RunE:    runChoreAll,
}

// taskCmd is a top-level alias group so `skylight task all` works.
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task views (alias for chore all)",
}

var taskAllCmd = &cobra.Command{
	Use:   resourceAll,
	Short: choreAllCmd.Short,
	Long:  choreAllLong,
	RunE:  runChoreAll,
}

var choreListCmd = &cobra.Command{
	Use:   subList,
	Short: "List chores",
	Long: `List chores for a Skylight frame.

NOTE: up-for-grabs (unassigned) chores are excluded from results by default.
Pass --up-for-grabs to include them:

  skylight chore list --up-for-grabs --after 2026-06-01 --before 2026-06-30

Use --status to filter by pending/complete/skipped, and --after/--before for
a date window (both flags are required together when either is supplied).`,
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
		if !choreUpForGrabs && !quiet {
			fmt.Fprintln(os.Stderr, "hint: up-for-grabs (unassigned) chores are excluded; use --up-for-grabs to include them")
		}
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
		if err := lib.ValidateStartTime(choreDueTime); err != nil {
			return err
		}

		if cmd.Flags().Changed("assignee-ids") {
			if cmd.Flags().Changed("assignee-id") {
				return fmt.Errorf("--assignee-id and --assignee-ids are mutually exclusive")
			}
			if choreUpForGrabs {
				return fmt.Errorf("--assignee-ids cannot be combined with --up-for-grabs")
			}
			if len(choreAssigneeIDs) < 2 {
				return fmt.Errorf("--assignee-ids needs at least two profile IDs; use --assignee-id for one")
			}
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.ChoreData{
			StartTime:   choreDueTime,
			Title:       choreTitle,
			Description: choreDescription,
			DueDate:     choreDate,
			Points:      chorePoints,
		}
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
		var chore *lib.Chore
		ctx := cmd.Context()
		if cmd.Flags().Changed("assignee-ids") {
			linked, err := createLinkedChores(cmd, client, data)
			if err != nil {
				return err
			}
			maybeLoadCatNames(ctx, client)
			printOutput(linked)
			return nil
		}
		if choreUpForGrabs {
			chore, err = client.CreateUpForGrabsChore(ctx, frameID, data)
		} else {
			data.AssigneeID = choreAssigneeID
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

		if cmd.Flags().Changed("apply-to-profiles") {
			if err := validateEnum(choreApplyToProfiles, choreApplyScopes); err != nil {
				return err
			}
			if choreDeleteRecurring {
				return fmt.Errorf("--apply-to-profiles and --recurring are mutually exclusive")
			}
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

		var delErr error
		if cmd.Flags().Changed("apply-to-profiles") {
			delErr = client.DeleteLinkedChore(cmd.Context(), frameID, choreID, choreApplyToProfiles)
		} else if choreDeleteRecurring {
			delErr = client.DeleteRecurringChore(cmd.Context(), frameID, choreID)
		} else {
			delErr = client.DeleteChore(cmd.Context(), frameID, choreID)
		}
		if delErr != nil {
			return fmt.Errorf("deleting chore: %w", delErr)
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

		if cmd.Flags().Changed(flagDueTime) {
			if err := lib.ValidateStartTime(choreDueTime); err != nil {
				return err
			}
		}
		if err := validateLinkedUpdateFlags(cmd); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.ChoreData{}
		if cmd.Flags().Changed(flagDueTime) {
			data.StartTime = choreDueTime
		}
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
		if cmd.Flags().Changed("apply-to-profiles") {
			data.ApplyToProfiles = choreApplyToProfiles
			data.After, data.Before = choreDate, choreDate
		}
		if cmd.Flags().Changed("assignee-ids") {
			data.CategoryIDs = choreAssigneeIDs
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

// validateLinkedUpdateFlags checks the linked-task flags of chore update: the API
// needs an instance date with --apply-to-profiles, and a profile list only makes
// sense when editing every linked copy.
func validateLinkedUpdateFlags(cmd *cobra.Command) error {
	hasScope := cmd.Flags().Changed("apply-to-profiles")
	hasIDs := cmd.Flags().Changed("assignee-ids")
	if hasScope {
		if err := validateEnum(choreApplyToProfiles, choreApplyScopes); err != nil {
			return err
		}
		if !cmd.Flags().Changed(subDate) {
			return fmt.Errorf("--apply-to-profiles requires --date (the instance date)")
		}
	}
	if hasIDs {
		if cmd.Flags().Changed("assignee-id") {
			return fmt.Errorf("--assignee-id and --assignee-ids are mutually exclusive")
		}
		if !hasScope || choreApplyToProfiles != lib.ApplyToProfilesAll {
			return fmt.Errorf("--assignee-ids requires --apply-to-profiles all")
		}
		if len(choreAssigneeIDs) == 0 {
			return fmt.Errorf("--assignee-ids needs at least one profile ID")
		}
	}
	return nil
}

// createLinkedChores creates one chore linked across the --assignee-ids profiles.
func createLinkedChores(cmd *cobra.Command, client *lib.Client, data lib.ChoreData) ([]lib.Chore, error) {
	data.CategoryIDs = choreAssigneeIDs
	chores, err := client.CreateLinkedChore(cmd.Context(), frameID, data)
	if err != nil {
		return nil, fmt.Errorf("creating linked chore: %w", err)
	}
	return chores, nil
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
	choreCmd.AddCommand(choreAllCmd)
	taskCmd.AddCommand(taskAllCmd)
	rootCmd.AddCommand(taskCmd)
	for _, c := range []*cobra.Command{choreAllCmd, taskAllCmd} {
		c.Flags().IntVar(&choreAllLimit, "limit", 0, "Maximum total entries across all buckets (0 = server default)")
		c.Flags().BoolVar(&choreAllUpForGrabs, "include-up-for-grabs", true, "Include up-for-grabs (unassigned) tasks")
	}
	choreCmd.AddCommand(choreGetCmd)
	choreCmd.AddCommand(choreCreateCmd)
	choreCmd.AddCommand(choreUpdateCmd)
	choreCmd.AddCommand(choreDeleteCmd)
	choreCmd.AddCommand(choreCompleteCmd)
	choreCmd.AddCommand(choreSkipCmd)
	choreCmd.AddCommand(choreClaimCmd)
	choreCmd.AddCommand(choreSearchCmd)
	choreCmd.AddCommand(choreNotificationsCmd)

	choreListCmd.Flags().StringVar(&choreDate, subDate, "", "Date filter")
	choreListCmd.Flags().StringVar(&choreStatus, "status", "", "Status filter: pending, complete, skipped")
	choreListCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID filter")
	registerEnumFlagCompletion(choreListCmd, "status", choreStatuses...)

	choreListCmd.Flags().StringVar(&choreAfter, "after", "", "Start of date window (YYYY-MM-DD); defaults to the current month")
	choreListCmd.Flags().StringVar(&choreBefore, "before", "", "End of date window (YYYY-MM-DD); defaults to the current month")
	choreListCmd.Flags().BoolVar(&choreIncludeLate, "include-late", false, "Include late chores")
	choreListCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Include up-for-grabs (unassigned) chores; excluded from results by default")
	choreListCmd.Flags().StringVar(&choreWeek, "week", "", "Show weekly calendar view; optionally specify YYYY-MM-DD to select the week")
	choreListCmd.Flags().Lookup("week").NoOptDefVal = "current"

	choreCreateCmd.Flags().StringVar(&choreTitle, subTitle, "", "Chore title")
	choreCreateCmd.Flags().StringVar(&choreDescription, "description", "", "Chore description")
	choreCreateCmd.Flags().StringVar(&choreDate, subDate, "", "Due date")
	choreCreateCmd.Flags().StringVar(&choreDueTime, flagDueTime, "", dueTimeUsage)
	choreCreateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreCreateCmd.Flags().StringSliceVar(&choreAssigneeIDs, "assignee-ids", nil, "Comma-separated profile IDs (2+) to create one linked task across")
	choreCreateCmd.Flags().IntVar(&chorePoints, subPoints, 0, "Points value")
	choreCreateCmd.Flags().BoolVar(&choreRecurring, "recurring", false, "Make chore recurring")
	choreCreateCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Make chore claimable by anyone")
	choreCreateCmd.Flags().StringVar(&choreFrequency, "frequency", "", "Recurrence frequency: daily, weekly, monthly")
	choreCreateCmd.Flags().IntVar(&choreInterval, "interval", 0, "Recurrence interval (every N periods)")
	choreCreateCmd.Flags().StringSliceVar(&choreRecurrenceDays, "recurrence-days", nil, "Days of week for weekly recurrence (e.g., mon,wed,fri)")
	choreCreateCmd.Flags().StringVar(&choreEndDate, "end-date", "", "End date for recurring chore (YYYY-MM-DD)")
	choreCreateCmd.Flags().StringVar(&choreRecurFrom, "recur-from", "", "Has no effect (not stored by the API)")
	markFlagRequired(choreCreateCmd, subTitle)

	choreUpdateCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to update")
	choreUpdateCmd.Flags().StringVar(&choreTitle, subTitle, "", "Chore title")
	choreUpdateCmd.Flags().StringVar(&choreDescription, "description", "", "Chore description")
	choreUpdateCmd.Flags().StringVar(&choreStatus, "status", "", "Chore status: pending, complete, skipped")
	choreUpdateCmd.Flags().IntVar(&chorePoints, subPoints, 0, "Points value")
	choreUpdateCmd.Flags().StringVar(&choreDueTime, flagDueTime, "", dueTimeUsage)
	choreUpdateCmd.Flags().StringVar(&choreAssigneeID, "assignee-id", "", "Assignee ID")
	choreUpdateCmd.Flags().StringSliceVar(&choreAssigneeIDs, "assignee-ids", nil, "Full list of profile IDs for a linked task (adds/removes profiles; requires --apply-to-profiles all)")
	choreUpdateCmd.Flags().StringVar(&choreApplyToProfiles, "apply-to-profiles", "", "Linked task scope: all (every profile's copy) or one (this copy only, detaches it); requires --date")
	registerEnumFlagCompletion(choreUpdateCmd, "apply-to-profiles", choreApplyScopes...)
	choreUpdateCmd.Flags().StringVar(&choreDate, subDate, "", "Due date (with --apply-to-profiles: the instance date)")
	choreUpdateCmd.Flags().StringVar(&choreFrequency, "frequency", "", "Recurrence frequency: daily, weekly, monthly (replaces the schedule)")
	choreUpdateCmd.Flags().IntVar(&choreInterval, "interval", 0, "Recurrence interval (every N periods)")
	choreUpdateCmd.Flags().StringSliceVar(&choreRecurrenceDays, "recurrence-days", nil, "Days of week for weekly recurrence (e.g., mon,wed,fri; replaces the schedule)")
	choreUpdateCmd.Flags().StringVar(&choreEndDate, "end-date", "", "End date for recurring chore (YYYY-MM-DD)")
	choreUpdateCmd.Flags().StringVar(&choreRecurFrom, "recur-from", "", "Has no effect (not stored by the API)")
	choreUpdateCmd.Flags().BoolVar(&choreUpForGrabs, "up-for-grabs", false, "Make chore claimable by anyone")
	registerEnumFlagCompletion(choreUpdateCmd, "status", choreStatuses...)
	markFlagRequired(choreUpdateCmd, "chore-id")

	choreGetCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to retrieve")
	markFlagRequired(choreGetCmd, "chore-id")

	choreDeleteCmd.Flags().StringVar(&choreID, "chore-id", "", "Chore ID to delete")
	choreDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	choreDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	choreDeleteCmd.Flags().BoolVar(&choreDeleteRecurring, "recurring", false, "Delete all instances of a recurring chore")
	choreDeleteCmd.Flags().StringVar(&choreApplyToProfiles, "apply-to-profiles", "", "Linked task scope: all (every profile) or one (this profile's copy only)")
	registerEnumFlagCompletion(choreDeleteCmd, "apply-to-profiles", choreApplyScopes...)
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
