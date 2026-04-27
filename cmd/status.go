package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Quick overview of the connected frame",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()
		client := getClient()
		today := time.Now().Format("2006-01-02")

		frame, err := client.GetFrame(frameID)
		if err != nil {
			fmt.Printf("Error getting frame: %v\n", err)
			os.Exit(1)
		}

		chores, err := client.ListChores(frameID, lib.ChoreListOptions{
			Date:   today,
			Status: "pending",
		})
		if err != nil {
			fmt.Printf("Error listing chores: %v\n", err)
			os.Exit(1)
		}

		events, err := client.ListCalendarEvents(frameID, today, today)
		if err != nil {
			fmt.Printf("Error listing calendar events: %v\n", err)
			os.Exit(1)
		}

		categories, err := client.ListCategories(frameID)
		if err != nil {
			fmt.Printf("Error listing categories: %v\n", err)
			os.Exit(1)
		}

		points, err := client.GetRewardPoints(frameID)
		if err != nil {
			fmt.Printf("Error getting reward points: %v\n", err)
			os.Exit(1)
		}

		catNames := make(map[int]string, len(categories))
		for _, c := range categories {
			id, convErr := strconv.Atoi(c.ID)
			if convErr == nil {
				catNames[id] = c.Name
			}
		}

		var pointParts []string
		for _, p := range points {
			name := catNames[p.CategoryID]
			if name == "" {
				name = strconv.Itoa(p.CategoryID)
			}
			pointParts = append(pointParts, fmt.Sprintf("%s: %d", name, p.CurrentPointBalance))
		}
		pointsStr := strings.Join(pointParts, "  ")
		if pointsStr == "" {
			pointsStr = "none"
		}

		fmt.Printf("Frame:   %s\n", frame.Name)
		fmt.Printf("Chores:  %d pending today\n", len(chores))
		fmt.Printf("Events:  %d today\n", len(events))
		fmt.Printf("Points:  %s\n", pointsStr)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
