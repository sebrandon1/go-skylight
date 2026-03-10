package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"today"},
	Short:   "Show today's dashboard (events, chores, points, meals, lists)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		dash, err := client.GetDashboard(frameID)
		if err != nil {
			fmt.Printf("Error getting dashboard: %v\n", err)
			os.Exit(1)
		}

		printJSON(dash)
	},
}
