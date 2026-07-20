package cmd

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		routines, err := client.ListRoutines(frameID)
		if err != nil {
			fatal("listing routines", err)
		}

		maybeLoadCatNames(client)
		printOutput(routines)
	},
}

var routineCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a routine",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		routine, err := client.CreateRoutine(frameID, lib.RoutineData{
			Title:      routineTitle,
			AssigneeID: routineAssignee,
			Steps:      filterEmptyStrings(routineSteps),
		})
		if err != nil {
			fatal("creating routine", err)
		}

		printJSON(routine)
	},
}

var routineUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a routine",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

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

		routine, err := client.UpdateRoutine(frameID, routineID, data)
		if err != nil {
			fatal("updating routine", err)
		}

		printJSON(routine)
	},
}

var routineDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a routine",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.DeleteRoutine(frameID, routineID); err != nil {
			fatal("deleting routine", err)
		}

		printSuccess("Routine deleted successfully")
	},
}

var routineReorderCmd = &cobra.Command{
	Use:   "reorder",
	Short: "Set the display order of routines",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.ReorderRoutines(frameID, routineIDs); err != nil {
			fatal("reordering routines", err)
		}

		printSuccess("Routines reordered successfully")
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
