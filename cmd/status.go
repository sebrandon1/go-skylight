package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Quick overview of the connected frame",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		today := time.Now().Format(lib.DateFormat)

		frame, err := client.GetFrame(ctx, frameID)
		if err != nil {
			return fmt.Errorf("getting frame: %w", err)
		}

		var (
			chores          []lib.Chore
			events          []lib.CalendarEvent
			categories      []lib.Category
			points          []lib.RewardPointEntry
			sittings        []lib.MealSitting
			lists           []lib.List
			routines        []lib.Routine
			incompleteItems int
		)

		if err := runConcurrent(
			func() error {
				var err error
				chores, err = client.ListChores(ctx, frameID, lib.ChoreListOptions{
					After:  today,
					Before: today,
					Status: lib.ChoreStatusPending,
				})
				if err != nil {
					return fmt.Errorf("listing chores: %w", err)
				}
				return nil
			},
			func() error {
				var err error
				events, err = client.ListCalendarEvents(ctx, frameID, today, today, frame.TimeZone)
				if err != nil {
					return fmt.Errorf("listing calendar events: %w", err)
				}
				return nil
			},
			func() error {
				var err error
				categories, err = client.ListCategories(ctx, frameID)
				if err != nil {
					return fmt.Errorf("listing categories: %w", err)
				}
				return nil
			},
			func() error {
				var err error
				points, err = client.GetRewardPoints(ctx, frameID)
				if err != nil {
					return fmt.Errorf("getting reward points: %w", err)
				}
				return nil
			},
			func() error {
				var err error
				sittings, err = client.ListMealSittings(ctx, frameID, lib.MealSittingListOptions{
					DateMin: today,
					DateMax: today,
				})
				if err != nil {
					return fmt.Errorf("listing meal sittings: %w", err)
				}
				return nil
			},
			func() error {
				var err error
				lists, err = client.ListLists(ctx, frameID)
				if err != nil {
					return fmt.Errorf("listing lists: %w", err)
				}
				incompleteItems = countIncompleteListItems(lists)
				return nil
			},
			func() error {
				var err error
				routines, err = client.ListRoutines(ctx, frameID)
				if err != nil {
					return fmt.Errorf("listing routines: %w", err)
				}
				return nil
			},
		); err != nil {
			return err
		}

		pointEntries := resolveRewardPointNames(points, categories)

		var pointParts []string
		for _, pe := range pointEntries {
			pointParts = append(pointParts, fmt.Sprintf("%s: %d", pe.Name, pe.Balance))
		}
		pointsStr := strings.Join(pointParts, "  ")
		if pointsStr == "" {
			pointsStr = "none"
		}

		if outputFormat == outputJSON {
			printJSON(map[string]any{
				subFrame:                frame.Name,
				"pending_chores":        len(chores),
				"events_today":          len(events),
				subPoints:               pointEntries,
				"meal_sittings_today":   len(sittings),
				"active_lists":          len(lists),
				"incomplete_list_items": incompleteItems,
				watchResourceRoutines:   len(routines),
			})
			return nil
		}

		fmt.Printf("Frame:    %s\n", frame.Name)
		fmt.Printf("Chores:   %d pending today\n", len(chores))
		fmt.Printf("Events:   %d today\n", len(events))
		fmt.Printf("Meals:    %d today\n", len(sittings))
		fmt.Printf("Lists:    %d active, %d incomplete items\n", len(lists), incompleteItems)
		fmt.Printf("Routines: %d\n", len(routines))
		fmt.Printf("Points:   %s\n", pointsStr)
		return nil
	},
}

func countIncompleteListItems(lists []lib.List) (incomplete int) {
	for _, l := range lists {
		for _, item := range l.Items {
			if !item.Completed {
				incomplete++
			}
		}
	}
	return incomplete
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
