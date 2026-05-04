package cmd

import (
	"fmt"
	"os"

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
			fmt.Fprintf(os.Stderr, "Error listing routines: %v\n", err)
			os.Exit(1)
		}

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
			fmt.Fprintf(os.Stderr, "Error creating routine: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error updating routine: %v\n", err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error deleting routine: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Routine deleted successfully")
	},
}

var routineReorderCmd = &cobra.Command{
	Use:   "reorder",
	Short: "Set the display order of routines",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.ReorderRoutines(frameID, routineIDs); err != nil {
			fmt.Fprintf(os.Stderr, "Error reordering routines: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Routines reordered successfully")
	},
}

func filterEmptyStrings(ss []string) []string {
	out := ss[:0:0]
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
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
	routineCreateCmd.MarkFlagRequired("title") //nolint:errcheck

	routineUpdateCmd.Flags().StringVar(&routineID, "routine-id", "", "Routine ID")
	routineUpdateCmd.Flags().StringVar(&routineTitle, "title", "", "Routine title")
	routineUpdateCmd.Flags().StringVar(&routineAssignee, "assignee-id", "", "Assignee ID")
	routineUpdateCmd.Flags().StringSliceVar(&routineSteps, "steps", nil, "Step titles (comma-separated)")
	routineUpdateCmd.MarkFlagRequired("routine-id") //nolint:errcheck

	routineDeleteCmd.Flags().StringVar(&routineID, "routine-id", "", "Routine ID")
	routineDeleteCmd.MarkFlagRequired("routine-id") //nolint:errcheck

	routineReorderCmd.Flags().StringSliceVar(&routineIDs, "routine-ids", nil, "Routine IDs in desired order (comma-separated)")
	routineReorderCmd.MarkFlagRequired("routine-ids") //nolint:errcheck
}
