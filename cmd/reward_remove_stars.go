package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	removeStarsAssigneeID int
	removeStarsPoints     int
)

var rewardRemoveStarsCmd = &cobra.Command{
	Use:   "remove-stars",
	Short: "Remove stars from a user balance (admin only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if removeStarsAssigneeID <= 0 {
			return fmt.Errorf("--assignee-id must be a positive integer")
		}

		if removeStarsPoints <= 0 {
			return fmt.Errorf("--points must be a positive integer")
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		if err := client.RemoveStars(ctx, frameID, removeStarsAssigneeID, removeStarsPoints); err != nil {
			return fmt.Errorf("removing stars: %w", err)
		}

		points, err := client.GetRewardPoints(ctx, frameID)
		if err != nil {
			return fmt.Errorf("fetching updated reward points: %w", err)
		}
		printJSON(points)
		return nil
	},
}

func init() {
	rewardRemoveStarsCmd.Flags().IntVar(&removeStarsAssigneeID, "assignee-id", 0, "Profile/category ID to deduct from")
	rewardRemoveStarsCmd.Flags().IntVar(&removeStarsPoints, subPoints, 0, "Number of stars to remove")
	markFlagRequired(rewardRemoveStarsCmd, "assignee-id")
	markFlagRequired(rewardRemoveStarsCmd, subPoints)
}
