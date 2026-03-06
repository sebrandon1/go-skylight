package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	rewardID     string
	rewardTitle  string
	rewardPoints int
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

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		rewards, err := client.ListRewards(frameID)
		if err != nil {
			fmt.Printf("Error listing rewards: %v\n", err)
			os.Exit(1)
		}

		printJSON(rewards)
	},
}

var rewardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		reward, err := client.CreateReward(frameID, lib.RewardData{
			Title:  rewardTitle,
			Points: rewardPoints,
		})
		if err != nil {
			fmt.Printf("Error creating reward: %v\n", err)
			os.Exit(1)
		}

		printJSON(reward)
	},
}

var rewardDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.DeleteReward(frameID, rewardID)
		if err != nil {
			fmt.Printf("Error deleting reward: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Reward deleted successfully")
	},
}

var rewardRedeemCmd = &cobra.Command{
	Use:   "redeem",
	Short: "Redeem a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.RedeemReward(frameID, rewardID)
		if err != nil {
			fmt.Printf("Error redeeming reward: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Reward redeemed successfully")
	},
}

var rewardUnredeemCmd = &cobra.Command{
	Use:   "unredeem",
	Short: "Unredeem a reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.UnredeemReward(frameID, rewardID)
		if err != nil {
			fmt.Printf("Error unredeeming reward: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Reward unredeemed successfully")
	},
}

var rewardPointsCmd = &cobra.Command{
	Use:   "points",
	Short: "Get reward points",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		points, err := client.GetRewardPoints(frameID)
		if err != nil {
			fmt.Printf("Error getting reward points: %v\n", err)
			os.Exit(1)
		}

		printJSON(points)
	},
}

func init() {
	rewardCmd.AddCommand(rewardListCmd)
	rewardCmd.AddCommand(rewardCreateCmd)
	rewardCmd.AddCommand(rewardDeleteCmd)
	rewardCmd.AddCommand(rewardRedeemCmd)
	rewardCmd.AddCommand(rewardUnredeemCmd)
	rewardCmd.AddCommand(rewardPointsCmd)

	rewardCreateCmd.Flags().StringVar(&rewardTitle, "title", "", "Reward title")
	rewardCreateCmd.Flags().IntVar(&rewardPoints, "points", 0, "Points cost")

	rewardDeleteCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
	rewardRedeemCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
	rewardUnredeemCmd.Flags().StringVar(&rewardID, "reward-id", "", "Reward ID")
}
