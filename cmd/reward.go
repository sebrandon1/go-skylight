package cmd

import (
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
	Long: `Create, list, update, delete, redeem, and unredeem rewards.

Each reward has a points cost; family members spend earned chore
points to redeem it. Use "reward points" to see current point
balances per family member.

  # Check point balances, then redeem a reward
  skylight reward points
  skylight reward redeem --reward-id 12345678`,
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

		if dryRun {
			printDryRun("delete reward %s", rewardID)
			return
		}

		client := getClient()

		err := client.DeleteReward(frameID, rewardID)
		if err != nil {
			fatal("deleting reward", err)
		}

		printSuccess("Reward deleted successfully")
	},
}

var rewardRedeemCmd = &cobra.Command{
	Use:   "redeem",
	Short: "Redeem a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if dryRun {
			printDryRun("redeem reward %s", rewardID)
			return
		}

		client := getClient()

		err := client.RedeemReward(frameID, rewardID)
		if err != nil {
			fatal("redeeming reward", err)
		}

		printSuccess("Reward redeemed successfully")
	},
}

var rewardUnredeemCmd = &cobra.Command{
	Use:   "unredeem",
	Short: "Unredeem a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if dryRun {
			printDryRun("unredeem reward %s", rewardID)
			return
		}

		client := getClient()

		err := client.UnredeemReward(frameID, rewardID)
		if err != nil {
			fatal("unredeeming reward", err)
		}

		printSuccess("Reward unredeemed successfully")
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

		categories, err := client.ListCategories(frameID)
		if err != nil {
			fatal("listing categories", err)
		}

		printOutput(resolveRewardPointNames(points, categories))
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
		if cmd.Flags().Changed("no-respawn") {
			// --no-respawn sets RespawnOnRedemption=false; explicit false re-enables.
			respawn := !rewardNoRespawn
			data.RespawnOnRedemption = &respawn
		}
		if cmd.Flags().Changed("category-ids") {
			data.CategoryIDs = rewardCategoryIDs
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
	markFlagRequired(rewardCreateCmd, "title")
	markFlagRequired(rewardCreateCmd, "points")

	rewardUpdateCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID to update")
	rewardUpdateCmd.Flags().StringVar(&rewardTitle, "title", "", "Reward title")
	rewardUpdateCmd.Flags().IntVar(&rewardPoints, "points", 0, "Points cost")
	rewardUpdateCmd.Flags().StringVar(&rewardEmojiIcon, "emoji-icon", "", "Emoji icon for the reward")
	rewardUpdateCmd.Flags().BoolVar(&rewardNoRespawn, "no-respawn", false, "Disable respawn on redemption")
	rewardUpdateCmd.Flags().IntSliceVar(&rewardCategoryIDs, "category-ids", nil, "Category IDs to assign reward to")
	markFlagRequired(rewardUpdateCmd, "reward-id")

	rewardDeleteCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
	rewardDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	markFlagRequired(rewardDeleteCmd, "reward-id")

	rewardRedeemCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
	rewardRedeemCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	markFlagRequired(rewardRedeemCmd, "reward-id")

	rewardUnredeemCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
	rewardUnredeemCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	markFlagRequired(rewardUnredeemCmd, "reward-id")
}
