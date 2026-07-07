package cmd

import (
	"fmt"
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
		today := time.Now().Format(lib.DateFormat)

		frame, err := client.GetFrame(frameID)
		if err != nil {
			fatal("getting frame", err)
		}

		chores, err := client.ListChores(frameID, lib.ChoreListOptions{
			After:  today,
			Before: today,
			Status: lib.ChoreStatusPending,
		})
		if err != nil {
			fatal("listing chores", err)
		}

		events, err := client.ListCalendarEvents(frameID, today, today, frame.TimeZone)
		if err != nil {
			fatal("listing calendar events", err)
		}

		categories, err := client.ListCategories(frameID)
		if err != nil {
			fatal("listing categories", err)
		}

		points, err := client.GetRewardPoints(frameID)
		if err != nil {
			fatal("getting reward points", err)
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

		if outputFormat == outputJSON {
			type pointEntry struct {
				Name    string `json:"name"`
				Balance int    `json:"balance"`
			}
			var pointEntries []pointEntry
			for _, p := range points {
				name := catNames[p.CategoryID]
				if name == "" {
					name = strconv.Itoa(p.CategoryID)
				}
				pointEntries = append(pointEntries, pointEntry{Name: name, Balance: p.CurrentPointBalance})
			}
			printJSON(map[string]any{
				"frame":          frame.Name,
				"pending_chores": len(chores),
				"events_today":   len(events),
				"points":         pointEntries,
			})
			return
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
