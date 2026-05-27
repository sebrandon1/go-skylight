package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	rewardID          string
	rewardTitle       string
	rewardPoints      int
	rewardEmojiIcon   string
	rewardNoRespawn   bool
	rewardCategoryIDs []int
)

var rewardCmd = &cobra.Command{
	Use:   "reward",
	Short: "Reward management commands",
}

var rewardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rewards",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		rewards, err := client.ListRewards(frameID)
		if err != nil {
			fatal("listing rewards", err)
		}

		printOutput(rewards)
	},
}

var rewardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.RewardData{
			Title:  rewardTitle,
			Points: rewardPoints,
		}
		if rewardEmojiIcon != "" {
			data.EmojiIcon = rewardEmojiIcon
		}
		if rewardNoRespawn {
			noRespawn := false
			data.RespawnOnRedemption = &noRespawn
		}
		if len(rewardCategoryIDs) > 0 {
			data.CategoryIDs = rewardCategoryIDs
		}
		reward, err := client.CreateReward(frameID, data)
		if err != nil {
			fatal("creating reward", err)
		}

		printJSON(reward)
	},
}

var rewardDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.DeleteReward(frameID, rewardID)
		if err != nil {
			fatal("deleting reward", err)
		}

		fmt.Println("Reward deleted successfully")
	},
}

var rewardRedeemCmd = &cobra.Command{
	Use:   "redeem",
	Short: "Redeem a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.RedeemReward(frameID, rewardID)
		if err != nil {
			fatal("redeeming reward", err)
		}

		fmt.Println("Reward redeemed successfully")
	},
}

var rewardUnredeemCmd = &cobra.Command{
	Use:   "unredeem",
	Short: "Unredeem a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.UnredeemReward(frameID, rewardID)
		if err != nil {
			fatal("unredeeming reward", err)
		}

		fmt.Println("Reward unredeemed successfully")
	},
}

var rewardPointsCmd = &cobra.Command{
	Use:   "points",
	Short: "Get reward points",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		points, err := client.GetRewardPoints(frameID)
		if err != nil {
			fatal("getting reward points", err)
		}

		printJSON(points)
	},
}

var rewardUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.RewardData{}
		if cmd.Flags().Changed("title") {
			data.Title = rewardTitle
		}
		if cmd.Flags().Changed("points") {
			data.Points = rewardPoints
		}
		if cmd.Flags().Changed("emoji-icon") {
			data.EmojiIcon = rewardEmojiIcon
		}

		reward, err := client.UpdateReward(frameID, rewardID, data)
		if err != nil {
			fatal("updating reward", err)
		}

		printJSON(reward)
	},
}

func init() {
	rewardCmd.AddCommand(rewardListCmd)
	rewardCmd.AddCommand(rewardCreateCmd)
	rewardCmd.AddCommand(rewardUpdateCmd)
	rewardCmd.AddCommand(rewardDeleteCmd)
	rewardCmd.AddCommand(rewardRedeemCmd)
	rewardCmd.AddCommand(rewardUnredeemCmd)
	rewardCmd.AddCommand(rewardPointsCmd)
	rewardCmd.AddCommand(rewardRemoveStarsCmd)

	rewardCreateCmd.Flags().StringVar(&rewardTitle, "title", "", "Reward title")
	rewardCreateCmd.Flags().IntVar(&rewardPoints, "points", 0, "Points cost")
	rewardCreateCmd.Flags().StringVar(&rewardEmojiIcon, "emoji-icon", "", "Emoji icon for the reward")
	rewardCreateCmd.Flags().BoolVar(&rewardNoRespawn, "no-respawn", false, "Disable respawn on redemption")
	rewardCreateCmd.Flags().IntSliceVar(&rewardCategoryIDs, "category-ids", nil, "Category IDs to assign reward to")
	rewardCreateCmd.MarkFlagRequired("title")  //nolint:errcheck
	rewardCreateCmd.MarkFlagRequired("points") //nolint:errcheck

	rewardUpdateCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID to update")
	rewardUpdateCmd.Flags().StringVar(&rewardTitle, "title", "", "Reward title")
	rewardUpdateCmd.Flags().IntVar(&rewardPoints, "points", 0, "Points cost")
	rewardUpdateCmd.Flags().StringVar(&rewardEmojiIcon, "emoji-icon", "", "Emoji icon for the reward")
	rewardUpdateCmd.MarkFlagRequired("reward-id") //nolint:errcheck

	rewardDeleteCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
	rewardDeleteCmd.MarkFlagRequired("reward-id") //nolint:errcheck

	rewardRedeemCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
	rewardRedeemCmd.MarkFlagRequired("reward-id") //nolint:errcheck

	rewardUnredeemCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
	rewardUnredeemCmd.MarkFlagRequired("reward-id") //nolint:errcheck
}
