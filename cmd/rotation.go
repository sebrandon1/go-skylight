package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	rotationChores      []string
	rotationAssigneeIDs []string
	rotationStartDate   string
	rotationWeeks       int
	rotationPoints      int
)

var rotationCmd = &cobra.Command{
	Use:   "rotation",
	Short: "Chore rotation management",
}

var rotationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a rotating chore schedule across family members",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		result, err := client.CreateChoreRotation(cmd.Context(), frameID, lib.RotationData{
			Chores:      rotationChores,
			AssigneeIDs: rotationAssigneeIDs,
			StartDate:   rotationStartDate,
			Weeks:       rotationWeeks,
			Points:      rotationPoints,
		})
		if err != nil {
			if result != nil && len(result.Chores) > 0 {
				fmt.Fprintf(os.Stderr, "Partial result (%d chores created):\n", len(result.Chores))
				printJSON(result)
			}
			return fmt.Errorf("creating rotation: %w", err)
		}

		printJSON(result)
		return nil
	},
}

func init() {
	rotationCmd.AddCommand(rotationCreateCmd)

	rotationCreateCmd.Flags().StringSliceVar(&rotationChores, "chores", nil, "Chore titles (comma-separated)")
	rotationCreateCmd.Flags().StringSliceVar(&rotationAssigneeIDs, "assignee-ids", nil, "Assignee IDs (comma-separated)")
	rotationCreateCmd.Flags().StringVar(&rotationStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	rotationCreateCmd.Flags().IntVar(&rotationWeeks, "weeks", 4, "Number of weeks")
	rotationCreateCmd.Flags().IntVar(&rotationPoints, "points", 0, "Points per chore")

	markFlagRequired(rotationCreateCmd, "chores")
	markFlagRequired(rotationCreateCmd, "assignee-ids")
	markFlagRequired(rotationCreateCmd, "start-date")
}
