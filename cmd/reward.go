package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	rewardID             string
	rewardTitle          string
	rewardPoints         int
	rewardEmojiIcon      string
	rewardNoRespawn      bool
	rewardCategoryIDs    []int
	rewardListAssigneeID string
	rewardListPointsMin  int
	rewardListPointsMax  int
	rewardListStatus     string
)

const (
	statusRedeemed  = "redeemed"
	statusAvailable = "available"
)

var rewardStatuses = []string{statusRedeemed, statusAvailable}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if rewardListStatus != "" {
			if err := validateEnum(rewardListStatus, rewardStatuses); err != nil {
				return err
			}
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		rewards, err := client.ListRewards(ctx, frameID)
		if err != nil {
			return fmt.Errorf("listing rewards: %w", err)
		}

		wantRedeemed := rewardListStatus == statusRedeemed
		filtered := rewards[:0:0]
		for _, r := range rewards {
			if rewardListAssigneeID != "" && r.CategoryID != rewardListAssigneeID {
				continue
			}
			if rewardListPointsMin > 0 && r.Points < rewardListPointsMin {
				continue
			}
			if rewardListPointsMax > 0 && r.Points > rewardListPointsMax {
				continue
			}
			if rewardListStatus != "" && r.Redeemed != wantRedeemed {
				continue
			}
			filtered = append(filtered, r)
		}
		rewards = filtered

		maybeLoadCatNames(ctx, client)
		printOutput(rewards)
		return nil
	},
}

var rewardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a reward",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

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
		reward, err := client.CreateReward(cmd.Context(), frameID, data)
		if err != nil {
			return fmt.Errorf("creating reward: %w", err)
		}

		printJSON(reward)
		return nil
	},
}

var rewardDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a reward",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete reward %s", rewardID)
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteReward(cmd.Context(), frameID, rewardID); err != nil {
			return fmt.Errorf("deleting reward: %w", err)
		}

		printSuccess("Reward deleted successfully")
		return nil
	},
}

var rewardRedeemCmd = &cobra.Command{
	Use:   "redeem",
	Short: "Redeem a reward",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("redeem reward %s", rewardID)
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.RedeemReward(cmd.Context(), frameID, rewardID); err != nil {
			return fmt.Errorf("redeeming reward: %w", err)
		}

		printSuccess("Reward redeemed successfully")
		return nil
	},
}

var rewardUnredeemCmd = &cobra.Command{
	Use:   "unredeem",
	Short: "Unredeem a reward",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("unredeem reward %s", rewardID)
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.UnredeemReward(cmd.Context(), frameID, rewardID); err != nil {
			return fmt.Errorf("unredeeming reward: %w", err)
		}

		printSuccess("Reward unredeemed successfully")
		return nil
	},
}

var rewardPointsCmd = &cobra.Command{
	Use:   "points",
	Short: "Get reward points",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		points, err := client.GetRewardPoints(ctx, frameID)
		if err != nil {
			return fmt.Errorf("getting reward points: %w", err)
		}

		categories, err := client.ListCategories(ctx, frameID)
		if err != nil {
			return fmt.Errorf("listing categories: %w", err)
		}

		printOutput(resolveRewardPointNames(points, categories))
		return nil
	},
}

var rewardUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a reward",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

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

		reward, err := client.UpdateReward(cmd.Context(), frameID, rewardID, data)
		if err != nil {
			return fmt.Errorf("updating reward: %w", err)
		}

		printJSON(reward)
		return nil
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

	rewardListCmd.Flags().StringVar(&rewardListAssigneeID, "assignee-id", "", "Filter by assignee/category ID")
	rewardListCmd.Flags().IntVar(&rewardListPointsMin, "points-min", 0, "Filter rewards with points >= this value")
	rewardListCmd.Flags().IntVar(&rewardListPointsMax, "points-max", 0, "Filter rewards with points <= this value")
	rewardListCmd.Flags().StringVar(&rewardListStatus, "status", "", "Filter by status: redeemed or available")
	registerEnumFlagCompletion(rewardListCmd, "status", rewardStatuses...)

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
