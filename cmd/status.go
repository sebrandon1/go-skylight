package cmd

import (
	"context"
	"fmt"
	"strings"
	"sync"
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

		chores, err := client.ListChores(ctx, frameID, lib.ChoreListOptions{
			After:  today,
			Before: today,
			Status: lib.ChoreStatusPending,
		})
		if err != nil {
			return fmt.Errorf("listing chores: %w", err)
		}

		events, err := client.ListCalendarEvents(ctx, frameID, today, today, frame.TimeZone)
		if err != nil {
			return fmt.Errorf("listing calendar events: %w", err)
		}

		categories, err := client.ListCategories(ctx, frameID)
		if err != nil {
			return fmt.Errorf("listing categories: %w", err)
		}

		points, err := client.GetRewardPoints(ctx, frameID)
		if err != nil {
			return fmt.Errorf("getting reward points: %w", err)
		}

		sittings, err := client.ListMealSittings(ctx, frameID, lib.MealSittingListOptions{
			DateMin: today,
			DateMax: today,
		})
		if err != nil {
			return fmt.Errorf("listing meal sittings: %w", err)
		}

		lists, err := client.ListLists(ctx, frameID)
		if err != nil {
			return fmt.Errorf("listing lists: %w", err)
		}
		incompleteItems, listErrors := countIncompleteListItems(ctx, client, frameID, lists)

		routines, err := client.ListRoutines(ctx, frameID)
		if err != nil {
			return fmt.Errorf("listing routines: %w", err)
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
				"list_errors":           listErrors,
				watchResourceRoutines:   len(routines),
			})
			return nil
		}

		fmt.Printf("Frame:    %s\n", frame.Name)
		fmt.Printf("Chores:   %d pending today\n", len(chores))
		fmt.Printf("Events:   %d today\n", len(events))
		fmt.Printf("Meals:    %d today\n", len(sittings))
		listsLine := fmt.Sprintf("%d active, %d incomplete items", len(lists), incompleteItems)
		if listErrors > 0 {
			listsLine += fmt.Sprintf(" (%d lists unavailable)", listErrors)
		}
		fmt.Printf("Lists:    %s\n", listsLine)
		fmt.Printf("Routines: %d\n", len(routines))
		fmt.Printf("Points:   %s\n", pointsStr)
		return nil
	},
}

// statusListWorkerCount bounds concurrent GetList calls in status (same idea
// as importWorkerCount) so large frames do not open unbounded connections.
const statusListWorkerCount = 5

// countIncompleteListItems fetches each list's full detail concurrently and
// counts incomplete items. ListLists does not populate item data, so this
// requires one GetList call per list. A failed list is excluded from the
// count rather than failing the whole status command (this is supplementary
// detail on top of the primary status fields), but the number of failures is
// returned so callers can surface it instead of silently under-reporting.
// Concurrency is capped (#271).
func countIncompleteListItems(ctx context.Context, client *lib.Client, frameID string, lists []lib.List) (incomplete, errors int) {
	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		sem = make(chan struct{}, statusListWorkerCount)
	)
	wg.Add(len(lists))
	for _, l := range lists {
		go func(l lib.List) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			full, err := client.GetList(ctx, frameID, l.ID)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors++
				return
			}
			for _, item := range full.Items {
				if !item.Completed {
					incomplete++
				}
			}
		}(l)
	}
	wg.Wait()
	return incomplete, errors
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
