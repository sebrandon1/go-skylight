package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rewardRemoveStarsCmd = &cobra.Command{
	Use:   "remove-stars",
	Short: "Remove stars from a user balance (admin only)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		var starCount int
		_, err := fmt.Sscanf(args[1], "%d", &starCount)
		if err != nil { return fmt.Errorf("invalid star count: %w", err) }
		if starCount <= 0 { return fmt.Errorf("star count must be positive") }
		fmt.Printf("Removing %d stars from user %s\n", starCount, userID)
		return nil
	},
}

func init() { rewardCmd.AddCommand(rewardRemoveStarsCmd) }
