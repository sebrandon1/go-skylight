package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var addonCmd = &cobra.Command{
	Use:   "addon",
	Short: "Manage frame add-ons",
}

var addonListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available add-ons and their enabled state",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		frame, err := client.GetFrame(frameID)
		if err != nil {
			return fmt.Errorf("listing addons: %w", err)
		}

		if !frame.Plus {
			fmt.Fprintln(os.Stderr, "Warning: Skylight Plus subscription required to access add-ons.")
		}
		if len(frame.FeatureBundle) == 0 {
			fmt.Println("No add-ons found.")
			return nil
		}

		printOutput(frame.FeatureBundle)
		return nil
	},
}

func init() {
	addonCmd.AddCommand(addonListCmd)
}
