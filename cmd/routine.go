package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var routineTimeOfDays = []string{lib.RoutineTODMorning, lib.RoutineTODAfternoon, lib.RoutineTODEvening}

var (
	routineID              string
	routineTitle           string
	routineTimeOfDay       string
	routineCategoryID      string
	routineStartDate       string
	routineListAssigneeID  string
	routineUpdateTitle     string
	routineUpdateTimeOfDay string
)

var routineCmd = &cobra.Command{
	Use:   "routine",
	Short: "Routine management commands",
	Long: `Create, list, and delete routines on a Skylight frame.

A routine is a recurring chore with a fixed time-of-day slot (morning,
afternoon, or evening). There is no separate routines resource on the
Skylight API -- routines are chores with routine:true.

  # Create a morning routine for a family member, then list all routines
  skylight routine create --title "Brush teeth" --time-of-day morning --category-id 12345678 --start-date 2026-01-01
  skylight routine list`,
}

var routineListCmd = &cobra.Command{
	Use:   subList,
	Short: "List routines",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		routines, err := client.ListRoutines(ctx, frameID)
		if err != nil {
			return fmt.Errorf("listing routines: %w", err)
		}

		if routineListAssigneeID != "" {
			filtered := routines[:0]
			for _, r := range routines {
				if r.AssigneeID == routineListAssigneeID {
					filtered = append(filtered, r)
				}
			}
			routines = filtered
		}

		maybeLoadCatNames(ctx, client)
		printOutput(routines)
		return nil
	},
}

var routineCreateCmd = &cobra.Command{
	Use:   subCreate,
	Short: "Create a routine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}
		if err := validateEnum(routineTimeOfDay, routineTimeOfDays); err != nil {
			return err
		}
		if err := validateDate(routineStartDate); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		routine, err := client.CreateRoutine(ctx, frameID, lib.RoutineData{
			Title:      routineTitle,
			TimeOfDay:  routineTimeOfDay,
			CategoryID: routineCategoryID,
			StartDate:  routineStartDate,
		})
		if err != nil {
			return fmt.Errorf("creating routine: %w", err)
		}

		maybeLoadCatNames(ctx, client)
		printOutput([]lib.Routine{*routine})
		return nil
	},
}

var routineUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update a routine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		upd := lib.RoutineUpdateData{}
		if cmd.Flags().Changed("title") {
			upd.Title = routineUpdateTitle
		}
		if cmd.Flags().Changed("time-of-day") {
			if err := validateEnum(routineUpdateTimeOfDay, routineTimeOfDays); err != nil {
				return err
			}
			upd.TimeOfDay = routineUpdateTimeOfDay
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		routine, err := client.UpdateRoutine(cmd.Context(), frameID, routineID, upd)
		if err != nil {
			return fmt.Errorf("updating routine: %w", err)
		}

		maybeLoadCatNames(cmd.Context(), client)
		printOutput([]lib.Routine{*routine})
		return nil
	},
}

var routineDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a routine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete routine %s", routineID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete routine %s?", routineID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteRoutine(cmd.Context(), frameID, routineID); err != nil {
			return fmt.Errorf("deleting routine: %w", err)
		}

		printSuccess("Routine deleted successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(routineCmd)

	routineCmd.AddCommand(routineListCmd)
	routineListCmd.Flags().StringVar(&routineListAssigneeID, "assignee-id", "", "Filter routines by assignee ID")
	routineCmd.AddCommand(routineCreateCmd)
	routineCmd.AddCommand(routineUpdateCmd)
	routineCmd.AddCommand(routineDeleteCmd)

	routineCreateCmd.Flags().StringVar(&routineTitle, subTitle, "", "Routine title")
	routineCreateCmd.Flags().StringVar(&routineTimeOfDay, "time-of-day", "", "Time of day: morning, afternoon, or evening")
	routineCreateCmd.Flags().StringVar(&routineCategoryID, "category-id", "", "Assignee category ID")
	routineCreateCmd.Flags().StringVar(&routineStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	markFlagRequired(routineCreateCmd, subTitle)
	markFlagRequired(routineCreateCmd, "time-of-day")
	markFlagRequired(routineCreateCmd, "category-id")
	markFlagRequired(routineCreateCmd, "start-date")

	routineUpdateCmd.Flags().StringVar(&routineID, "routine-id", "", "Routine ID")
	routineUpdateCmd.Flags().StringVar(&routineUpdateTitle, subTitle, "", "New title")
	routineUpdateCmd.Flags().StringVar(&routineUpdateTimeOfDay, "time-of-day", "", "New time of day: morning, afternoon, or evening")
	markFlagRequired(routineUpdateCmd, "routine-id")

	routineDeleteCmd.Flags().StringVar(&routineID, "routine-id", "", "Routine ID")
	routineDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	routineDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(routineDeleteCmd, "routine-id")
}
