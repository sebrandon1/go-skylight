package cmd

import (
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

		sittings, err := client.ListMealSittings(frameID, lib.MealSittingListOptions{
			DateMin: today,
			DateMax: today,
		})
		if err != nil {
			fatal("listing meal sittings", err)
		}

		lists, err := client.ListLists(frameID)
		if err != nil {
			fatal("listing lists", err)
		}
		incompleteItems, listErrors := countIncompleteListItems(client, frameID, lists)

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
				"frame":                 frame.Name,
				"pending_chores":        len(chores),
				"events_today":          len(events),
				"points":                pointEntries,
				"meal_sittings_today":   len(sittings),
				"active_lists":          len(lists),
				"incomplete_list_items": incompleteItems,
				"list_errors":           listErrors,
			})
			return
		}

		fmt.Printf("Frame:   %s\n", frame.Name)
		fmt.Printf("Chores:  %d pending today\n", len(chores))
		fmt.Printf("Events:  %d today\n", len(events))
		fmt.Printf("Meals:   %d today\n", len(sittings))
		listsLine := fmt.Sprintf("%d active, %d incomplete items", len(lists), incompleteItems)
		if listErrors > 0 {
			listsLine += fmt.Sprintf(" (%d lists unavailable)", listErrors)
		}
		fmt.Printf("Lists:   %s\n", listsLine)
		fmt.Printf("Points:  %s\n", pointsStr)
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
func countIncompleteListItems(client *lib.Client, frameID string, lists []lib.List) (incomplete, errors int) {
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

			full, err := client.GetList(frameID, l.ID)

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
