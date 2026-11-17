package cmd

import (
	"fmt"
	"strings"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

const (
	flagTaskDue             = "task-due"
	flagTaskCompleted       = "task-completed"
	flagDueCategoryID       = "due-category-id"
	flagDueUpForGrabs       = "due-up-for-grabs"
	flagCompletedCategoryID = "completed-category-id"
	flagCompletedUpForGrabs = "completed-up-for-grabs"
	notificationOn          = "on"
	notificationOff         = "off"
	onOffUsage              = "on or off"
	categoryFilterUsage     = "Only notify for tasks assigned to these family member IDs (repeatable; pass an empty value to clear). Applies to "
)

var (
	notifTaskDue             string
	notifTaskCompleted       string
	notifDueCategoryIDs      []string
	notifDueUpForGrabs       string
	notifCompletedCategoryID []string
	notifCompletedUpForGrabs string
)

var onOffValues = []string{notificationOn, notificationOff}

var choreNotificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "Show or change the frame's task notification settings",
	Long: `Show or change the frame-wide task notification settings.

There is no per-task reminder. A "task due" notification fires for tasks that
have a due time (set one with chore create/update --due-time HH:MM), and the
filters choose whose tasks notify.`,
}

var choreNotificationsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show task due and task completed notification settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		settings, err := client.GetTaskNotificationSettings(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("getting task notification settings: %w", err)
		}

		maybeLoadCatNames(cmd.Context(), client)
		printOutput(settings)
		return nil
	},
}

// onOffFlag parses an on|off flag value, returning nil when the flag was not set.
func onOffFlag(cmd *cobra.Command, name, value string) (*bool, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	if value != notificationOn && value != notificationOff {
		return nil, fmt.Errorf("--%s: invalid value %q: use on or off", name, value)
	}
	b := value == notificationOn
	return &b, nil
}

// notificationChange builds a partial change for one rule from its flags,
// returning nil when none of them were set.
func notificationChange(cmd *cobra.Command, enabledFlag, enabled, catFlag string, cats []string, grabsFlag, grabs string) (*lib.TaskNotificationChange, error) {
	var ch lib.TaskNotificationChange
	var err error
	if ch.Enabled, err = onOffFlag(cmd, enabledFlag, enabled); err != nil {
		return nil, err
	}
	if ch.UpForGrabs, err = onOffFlag(cmd, grabsFlag, grabs); err != nil {
		return nil, err
	}
	if cmd.Flags().Changed(catFlag) {
		ids := []string{}
		for _, id := range cats {
			if id != "" {
				ids = append(ids, id)
			}
		}
		ch.CategoryIDs = &ids
	}
	if ch.Enabled == nil && ch.UpForGrabs == nil && ch.CategoryIDs == nil {
		return nil, nil
	}
	return &ch, nil
}

var choreNotificationsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Change task notification settings (only the flags you pass are changed)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		var update lib.TaskNotificationUpdate
		var err error
		if update.TaskDue, err = notificationChange(cmd, flagTaskDue, notifTaskDue, flagDueCategoryID, notifDueCategoryIDs, flagDueUpForGrabs, notifDueUpForGrabs); err != nil {
			return err
		}
		if update.TaskCompleted, err = notificationChange(cmd, flagTaskCompleted, notifTaskCompleted, flagCompletedCategoryID, notifCompletedCategoryID, flagCompletedUpForGrabs, notifCompletedUpForGrabs); err != nil {
			return err
		}
		if update.TaskDue == nil && update.TaskCompleted == nil {
			return fmt.Errorf("nothing to change: pass at least one of --%s, --%s, --%s, --%s, --%s, --%s",
				flagTaskDue, flagTaskCompleted, flagDueCategoryID, flagDueUpForGrabs, flagCompletedCategoryID, flagCompletedUpForGrabs)
		}

		if dryRun {
			printDryRun("update task notification settings on frame %s", frameID)
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		settings, err := client.UpdateTaskNotificationSettings(cmd.Context(), frameID, update)
		if err != nil {
			return fmt.Errorf("updating task notification settings: %w", err)
		}

		maybeLoadCatNames(cmd.Context(), client)
		printOutput(settings)
		return nil
	},
}

func yesNo(b bool) string {
	if b {
		return boolYes
	}
	return boolNo
}

func printTaskNotificationsTable(s *lib.TaskNotificationSettings) {
	w := newTableWriter()
	fmt.Fprintln(w, "NOTIFICATION\tENABLED\tCATEGORIES\tUP FOR GRABS")
	for _, row := range []struct {
		name string
		n    lib.TaskNotification
	}{{"task due", s.TaskDue}, {"task completed", s.TaskCompleted}} {
		cats := "-"
		if len(row.n.Filters.CategoryIDs) > 0 {
			names := make([]string, len(row.n.Filters.CategoryIDs))
			for i, id := range row.n.Filters.CategoryIDs {
				names[i] = resolveCatName(id)
			}
			cats = strings.Join(names, ",")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", row.name, yesNo(row.n.Enabled), cats, yesNo(row.n.Filters.UpForGrabs))
	}
	w.Flush()
}

func init() {
	choreNotificationsCmd.AddCommand(choreNotificationsShowCmd)
	choreNotificationsCmd.AddCommand(choreNotificationsSetCmd)

	f := choreNotificationsSetCmd.Flags()
	f.StringVar(&notifTaskDue, flagTaskDue, "", "Notify when a task comes due: "+onOffUsage)
	f.StringSliceVar(&notifDueCategoryIDs, flagDueCategoryID, nil, categoryFilterUsage+"task due.")
	f.StringVar(&notifDueUpForGrabs, flagDueUpForGrabs, "", "Include up-for-grabs tasks in task due notifications: "+onOffUsage)
	f.StringVar(&notifTaskCompleted, flagTaskCompleted, "", "Notify when a task is completed: "+onOffUsage)
	f.StringSliceVar(&notifCompletedCategoryID, flagCompletedCategoryID, nil, categoryFilterUsage+"task completed.")
	f.StringVar(&notifCompletedUpForGrabs, flagCompletedUpForGrabs, "", "Include up-for-grabs tasks in task completed notifications: "+onOffUsage)
	f.BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	registerEnumFlagCompletion(choreNotificationsSetCmd, flagTaskDue, onOffValues...)
	registerEnumFlagCompletion(choreNotificationsSetCmd, flagTaskCompleted, onOffValues...)
	registerEnumFlagCompletion(choreNotificationsSetCmd, flagDueUpForGrabs, onOffValues...)
	registerEnumFlagCompletion(choreNotificationsSetCmd, flagCompletedUpForGrabs, onOffValues...)
}
