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
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		frame, err := client.GetFrame(frameID)
		if err != nil {
			fatal("listing addons", err)
		}

		if !frame.Plus {
			fmt.Fprintln(os.Stderr, "Warning: Skylight Plus subscription required to access add-ons.")
		}
		if len(frame.FeatureBundle) == 0 {
			fmt.Println("No add-ons found.")
			return
		}

		printOutput(frame.FeatureBundle)
	},
}

func init() {
	addonCmd.AddCommand(addonListCmd)
}
