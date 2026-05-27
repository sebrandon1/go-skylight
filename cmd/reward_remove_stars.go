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
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if removeStarsPoints <= 0 {
			fatal("removing stars", fmt.Errorf("--points must be a positive integer"))
		}

		client := getClient()

		err := client.RemoveStars(frameID, removeStarsAssigneeID, removeStarsPoints)
		if err != nil {
			fatal("removing stars", err)
		}

		points, err := client.GetRewardPoints(frameID)
		if err != nil {
			fatal("fetching updated reward points", err)
		}
		printJSON(points)
	},
}

func init() {
	rewardRemoveStarsCmd.Flags().IntVar(&removeStarsAssigneeID, "assignee-id", 0, "Profile/category ID to deduct from")
	rewardRemoveStarsCmd.Flags().IntVar(&removeStarsPoints, "points", 0, "Number of stars to remove")
	rewardRemoveStarsCmd.MarkFlagRequired("assignee-id") //nolint:errcheck
	rewardRemoveStarsCmd.MarkFlagRequired("points")      //nolint:errcheck
}
