package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	bountyTitle       string
	bountyPoints      int
	bountyAssigneeID  string
	bountyDueDate     string
	bountyRewardTitle string
	bountyEmojiIcon   string
	bountyRecurring   bool
)

var bountyCmd = &cobra.Command{
	Use:   "bounty",
	Short: "Bounty management (chore + reward pairs)",
}

var bountyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a bounty (chore + paired reward)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		bounty, err := client.CreateBounty(frameID, lib.BountyData{
			Title:       bountyTitle,
			Points:      bountyPoints,
			DueDate:     bountyDueDate,
			AssigneeID:  bountyAssigneeID,
			Recurring:   bountyRecurring,
			RewardTitle: bountyRewardTitle,
			EmojiIcon:   bountyEmojiIcon,
		})
		if err != nil {
			fmt.Printf("Error creating bounty: %v\n", err)
			os.Exit(1)
		}

		printJSON(bounty)
	},
}

var bountyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bounties (matched chore+reward pairs)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		bounties, err := client.ListBounties(frameID)
		if err != nil {
			fmt.Printf("Error listing bounties: %v\n", err)
			os.Exit(1)
		}

		printJSON(bounties)
	},
}

func init() {
	bountyCmd.AddCommand(bountyCreateCmd)
	bountyCmd.AddCommand(bountyListCmd)

	bountyCreateCmd.Flags().StringVar(&bountyTitle, "title", "", "Chore title")
	bountyCreateCmd.Flags().IntVar(&bountyPoints, "points", 0, "Point value for chore and reward")
	bountyCreateCmd.Flags().StringVar(&bountyAssigneeID, "assignee-id", "", "Assignee ID")
	bountyCreateCmd.Flags().StringVar(&bountyDueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	bountyCreateCmd.Flags().StringVar(&bountyRewardTitle, "reward-title", "", "Reward title")
	bountyCreateCmd.Flags().StringVar(&bountyEmojiIcon, "emoji-icon", "", "Reward emoji icon")
	bountyCreateCmd.Flags().BoolVar(&bountyRecurring, "recurring", false, "Make chore recurring")
}
