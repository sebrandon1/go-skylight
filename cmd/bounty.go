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
	bountyChoreID     string
	bountyRewardID    string
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

		if err := validateDate(bountyDueDate); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

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
			fatal("creating bounty", err)
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
			fatal("listing bounties", err)
		}

		printOutput(bounties)
	},
}

var bountyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a bounty (removes both the chore and the paired reward)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.DeleteBounty(frameID, bountyChoreID, bountyRewardID); err != nil {
			fatal("deleting bounty", err)
		}

		printSuccess("Bounty deleted.")
	},
}

var bountyUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a bounty's chore and paired reward",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if cmd.Flags().Changed("due-date") {
			if err := validateDate(bountyDueDate); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		client := getClient()

		data := lib.BountyData{}
		if cmd.Flags().Changed("title") {
			data.Title = bountyTitle
		}
		if cmd.Flags().Changed("reward-title") {
			data.RewardTitle = bountyRewardTitle
		}
		if cmd.Flags().Changed("points") {
			data.Points = bountyPoints
		}
		if cmd.Flags().Changed("due-date") {
			data.DueDate = bountyDueDate
		}
		if cmd.Flags().Changed("emoji-icon") {
			data.EmojiIcon = bountyEmojiIcon
		}

		bounty, err := client.UpdateBounty(frameID, bountyChoreID, bountyRewardID, data)
		if err != nil {
			fatal("updating bounty", err)
		}

		printJSON(bounty)
	},
}

func init() {
	bountyCmd.AddCommand(bountyCreateCmd)
	bountyCmd.AddCommand(bountyListCmd)
	bountyCmd.AddCommand(bountyDeleteCmd)
	bountyCmd.AddCommand(bountyUpdateCmd)

	bountyCreateCmd.Flags().StringVar(&bountyTitle, "title", "", "Chore title")
	bountyCreateCmd.Flags().IntVar(&bountyPoints, "points", 0, "Point value for chore and reward")
	bountyCreateCmd.Flags().StringVar(&bountyAssigneeID, "assignee-id", "", "Assignee ID")
	bountyCreateCmd.Flags().StringVar(&bountyDueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	bountyCreateCmd.Flags().StringVar(&bountyRewardTitle, "reward-title", "", "Reward title")
	bountyCreateCmd.Flags().StringVar(&bountyEmojiIcon, "emoji-icon", "", "Reward emoji icon")
	bountyCreateCmd.Flags().BoolVar(&bountyRecurring, "recurring", false, "Make chore recurring")
	markFlagRequired(bountyCreateCmd, "title")
	markFlagRequired(bountyCreateCmd, "points")
	markFlagRequired(bountyCreateCmd, "reward-title")

	bountyDeleteCmd.Flags().StringVar(&bountyChoreID, "chore-id", "", "Chore ID of the bounty")
	bountyDeleteCmd.Flags().StringVar(&bountyRewardID, "reward-id", "", "Reward ID of the bounty")
	markFlagRequired(bountyDeleteCmd, "chore-id")
	markFlagRequired(bountyDeleteCmd, "reward-id")

	bountyUpdateCmd.Flags().StringVar(&bountyChoreID, "chore-id", "", "Chore ID of the bounty")
	bountyUpdateCmd.Flags().StringVar(&bountyRewardID, "reward-id", "", "Reward ID of the bounty")
	bountyUpdateCmd.Flags().StringVar(&bountyTitle, "title", "", "New chore title")
	bountyUpdateCmd.Flags().StringVar(&bountyRewardTitle, "reward-title", "", "New reward title")
	bountyUpdateCmd.Flags().IntVar(&bountyPoints, "points", 0, "New point value for chore and reward")
	bountyUpdateCmd.Flags().StringVar(&bountyDueDate, "due-date", "", "New due date (YYYY-MM-DD)")
	bountyUpdateCmd.Flags().StringVar(&bountyEmojiIcon, "emoji-icon", "", "New reward emoji icon")
	markFlagRequired(bountyUpdateCmd, "chore-id")
	markFlagRequired(bountyUpdateCmd, "reward-id")
}
