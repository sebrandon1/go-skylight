package cmd

import (
	"fmt"
	"strconv"
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

// countIncompleteListItems fetches each list's full detail concurrently and
// counts incomplete items. ListLists does not populate item data, so this
// requires one GetList call per list. A failed list is excluded from the
// count rather than failing the whole status command (this is supplementary
// detail on top of the primary status fields), but the number of failures is
// returned so callers can surface it instead of silently under-reporting.
func countIncompleteListItems(client *lib.Client, frameID string, lists []lib.List) (incomplete, errors int) {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	wg.Add(len(lists))
	for _, l := range lists {
		go func(l lib.List) {
			defer wg.Done()
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
