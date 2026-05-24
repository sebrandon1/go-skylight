package cmd

import (
	"github.com/spf13/cobra"
)

var (
	removeStarsAssigneeID string
	removeStarsPoints     int
)

var rewardRemoveStarsCmd = &cobra.Command{
	Use:   "remove-stars",
	Short: "Remove stars from a user balance (admin only)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.RemoveStars(frameID, removeStarsAssigneeID, removeStarsPoints)
		if err != nil {
			fatal("removing stars", err)
		}

		printJSON(map[string]any{
			"assignee_id": removeStarsAssigneeID,
			"points":      removeStarsPoints,
			"status":      "removed",
		})
	},
}

func init() {
	rewardRemoveStarsCmd.Flags().StringVar(&removeStarsAssigneeID, "assignee-id", "", "Profile/category ID to deduct from")
	rewardRemoveStarsCmd.Flags().IntVar(&removeStarsPoints, "points", 0, "Number of stars to remove")
	rewardRemoveStarsCmd.MarkFlagRequired("assignee-id") //nolint:errcheck
	rewardRemoveStarsCmd.MarkFlagRequired("points")      //nolint:errcheck
}
