package cmd

import (
	"fmt"

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
	Long: `Create and manage bounties: a chore and a reward created together,
matched by point value so completing the chore earns exactly enough
points to redeem the paired reward.

  # Create a $10-point bounty due next Friday
  skylight bounty create --title "Wash the car" --points 10 --reward-title "Ice cream" --due-date 2026-06-05`,
}

var bountyCreateCmd = &cobra.Command{
	Use:   subCreate,
	Short: "Create a bounty (chore + paired reward)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if err := validateDate(bountyDueDate); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		bounty, err := client.CreateBounty(cmd.Context(), frameID, lib.BountyData{
			Title:       bountyTitle,
			Points:      bountyPoints,
			DueDate:     bountyDueDate,
			AssigneeID:  bountyAssigneeID,
			Recurring:   bountyRecurring,
			RewardTitle: bountyRewardTitle,
			EmojiIcon:   bountyEmojiIcon,
		})
		if err != nil {
			return fmt.Errorf("creating bounty: %w", err)
		}

		printOutput(*bounty)
		return nil
	},
}

var bountyListCmd = &cobra.Command{
	Use:   subList,
	Short: "List bounties (matched chore+reward pairs)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		bounties, err := client.ListBounties(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing bounties: %w", err)
		}

		printOutput(bounties)
		return nil
	},
}

var bountyDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a bounty (removes both the chore and the paired reward)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete bounty (chore %s + reward %s)", bountyChoreID, bountyRewardID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete bounty (chore %s + reward %s)?", bountyChoreID, bountyRewardID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteBounty(cmd.Context(), frameID, bountyChoreID, bountyRewardID); err != nil {
			return fmt.Errorf("deleting bounty: %w", err)
		}

		printSuccess("Bounty deleted successfully")
		return nil
	},
}

var bountyUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update a bounty's chore and paired reward",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if cmd.Flags().Changed("due-date") {
			if err := validateDate(bountyDueDate); err != nil {
				return err
			}
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.BountyData{}
		if cmd.Flags().Changed(subTitle) {
			data.Title = bountyTitle
		}
		if cmd.Flags().Changed("reward-title") {
			data.RewardTitle = bountyRewardTitle
		}
		if cmd.Flags().Changed(subPoints) {
			data.Points = bountyPoints
		}
		if cmd.Flags().Changed("due-date") {
			data.DueDate = bountyDueDate
		}
		if cmd.Flags().Changed("emoji-icon") {
			data.EmojiIcon = bountyEmojiIcon
		}

		bounty, err := client.UpdateBounty(cmd.Context(), frameID, bountyChoreID, bountyRewardID, data)
		if err != nil {
			return fmt.Errorf("updating bounty: %w", err)
		}

		printOutput(*bounty)
		return nil
	},
}

func init() {
	bountyCmd.AddCommand(bountyCreateCmd)
	bountyCmd.AddCommand(bountyListCmd)
	bountyCmd.AddCommand(bountyDeleteCmd)
	bountyCmd.AddCommand(bountyUpdateCmd)

	bountyCreateCmd.Flags().StringVar(&bountyTitle, subTitle, "", "Chore title")
	bountyCreateCmd.Flags().IntVar(&bountyPoints, subPoints, 0, "Point value for chore and reward")
	bountyCreateCmd.Flags().StringVar(&bountyAssigneeID, "assignee-id", "", "Assignee ID")
	bountyCreateCmd.Flags().StringVar(&bountyDueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	bountyCreateCmd.Flags().StringVar(&bountyRewardTitle, "reward-title", "", "Reward title")
	bountyCreateCmd.Flags().StringVar(&bountyEmojiIcon, "emoji-icon", "", "Reward emoji icon")
	bountyCreateCmd.Flags().BoolVar(&bountyRecurring, "recurring", false, "Make chore recurring")
	markFlagRequired(bountyCreateCmd, subTitle)
	markFlagRequired(bountyCreateCmd, subPoints)
	markFlagRequired(bountyCreateCmd, "reward-title")

	bountyDeleteCmd.Flags().StringVar(&bountyChoreID, "chore-id", "", "Chore ID of the bounty")
	bountyDeleteCmd.Flags().StringVar(&bountyRewardID, "reward-id", "", "Reward ID of the bounty")
	bountyDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	bountyDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(bountyDeleteCmd, "chore-id")
	markFlagRequired(bountyDeleteCmd, "reward-id")

	bountyUpdateCmd.Flags().StringVar(&bountyChoreID, "chore-id", "", "Chore ID of the bounty")
	bountyUpdateCmd.Flags().StringVar(&bountyRewardID, "reward-id", "", "Reward ID of the bounty")
	bountyUpdateCmd.Flags().StringVar(&bountyTitle, subTitle, "", "New chore title")
	bountyUpdateCmd.Flags().StringVar(&bountyRewardTitle, "reward-title", "", "New reward title")
	bountyUpdateCmd.Flags().IntVar(&bountyPoints, subPoints, 0, "New point value for chore and reward")
	bountyUpdateCmd.Flags().StringVar(&bountyDueDate, "due-date", "", "New due date (YYYY-MM-DD)")
	bountyUpdateCmd.Flags().StringVar(&bountyEmojiIcon, "emoji-icon", "", "New reward emoji icon")
	markFlagRequired(bountyUpdateCmd, "chore-id")
	markFlagRequired(bountyUpdateCmd, "reward-id")
}
