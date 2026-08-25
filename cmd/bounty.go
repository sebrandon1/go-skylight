package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/sebrandon1/go-skylight/lib"
)

var bountyCmd = &cobra.Command{
	Use:   subBounty,
	Short: "Manage bounties",
}

var bountyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new bounty",
	RunE: func(cmd *cobra.Command, args []string) error {
		bounty, err := client.CreateBounty(args[0], args[1])
		if err!= nil {
			return err
		}
		return printOutput(bounty)
	},
}

var bountyUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing bounty",
	RunE: func(cmd *cobra.Command, args []string) error {
		bounty, err := client.UpdateBounty(args[0], args[1], args[2])
		if err!= nil {
			return err
		}
		return printOutput(bounty)
	},
}

var bountyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bounties",
	RunE: func(cmd *cobra.Command, args []string) error {
		bounties, err := client.ListBounties(args[0])
		if err!= nil {
			return err
		}
		return printOutput(bounties)
	},
}

var bountyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a bounty",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteBounty(args[0]); err!= nil {
			return err
		}
		return nil
	},
}

func init() {
	bountyCmd.AddCommand(bountyCreateCmd, bountyUpdateCmd, bountyListCmd, bountyDeleteCmd)
	rootCmd.AddCommand(bountyCmd)
}