package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	routineID       string
	routineTitle    string
	routineAssignee string
	routineSteps    []string
	routineIDs      []string
)

var routineCmd = &cobra.Command{
	Use:   "routine",
	Short: "Routine management commands",
}

var routineListCmd = &cobra.Command{
	Use:   "list",
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

		maybeLoadCatNames(ctx, client)
		printOutput(routines)
		return nil
	},
}

var routineCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a routine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		routine, err := client.CreateRoutine(cmd.Context(), frameID, lib.RoutineData{
			Title:      routineTitle,
			AssigneeID: routineAssignee,
			Steps:      filterEmptyStrings(routineSteps),
		})
		if err != nil {
			return fmt.Errorf("creating routine: %w", err)
		}

		printJSON(routine)
		return nil
	},
}

var routineUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a routine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.RoutineData{}
		if cmd.Flags().Changed("title") {
			data.Title = routineTitle
		}
		if cmd.Flags().Changed("assignee-id") {
			data.AssigneeID = routineAssignee
		}
		if cmd.Flags().Changed("steps") {
			data.Steps = filterEmptyStrings(routineSteps)
		}

		routine, err := client.UpdateRoutine(cmd.Context(), frameID, routineID, data)
		if err != nil {
			return fmt.Errorf("updating routine: %w", err)
		}

		printJSON(routine)
		return nil
	},
}

var routineDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a routine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
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

var routineReorderCmd = &cobra.Command{
	Use:   "reorder",
	Short: "Set the display order of routines",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.ReorderRoutines(cmd.Context(), frameID, routineIDs); err != nil {
			return fmt.Errorf("reordering routines: %w", err)
		}

		printSuccess("Routines reordered successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(routineCmd)

	routineCmd.AddCommand(routineListCmd)
	routineCmd.AddCommand(routineCreateCmd)
	routineCmd.AddCommand(routineUpdateCmd)
	routineCmd.AddCommand(routineDeleteCmd)
	routineCmd.AddCommand(routineReorderCmd)

	routineCreateCmd.Flags().StringVar(&routineTitle, "title", "", "Routine title")
	routineCreateCmd.Flags().StringVar(&routineAssignee, "assignee-id", "", "Assignee ID")
	routineCreateCmd.Flags().StringSliceVar(&routineSteps, "steps", nil, "Step titles (comma-separated)")
	markFlagRequired(routineCreateCmd, "title")

	routineUpdateCmd.Flags().StringVar(&routineID, "routine-id", "", "Routine ID")
	routineUpdateCmd.Flags().StringVar(&routineTitle, "title", "", "Routine title")
	routineUpdateCmd.Flags().StringVar(&routineAssignee, "assignee-id", "", "Assignee ID")
	routineUpdateCmd.Flags().StringSliceVar(&routineSteps, "steps", nil, "Step titles (comma-separated)")
	markFlagRequired(routineUpdateCmd, "routine-id")

	routineDeleteCmd.Flags().StringVar(&routineID, "routine-id", "", "Routine ID")
	markFlagRequired(routineDeleteCmd, "routine-id")

	routineReorderCmd.Flags().StringSliceVar(&routineIDs, "routine-ids", nil, "Routine IDs in desired order (comma-separated)")
	markFlagRequired(routineReorderCmd, "routine-ids")
}
