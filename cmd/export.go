package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

const (
	exportResourceChores   = "chores"
	exportResourceRewards  = "rewards"
	exportResourceLists    = "lists"
	exportResourceRecipes  = "recipes"
	exportResourceSittings = "sittings"
	exportResourceCalendar = "calendar"

	// resourceAll is the documented default/wildcard value for the various
	// --resources flags across commands, meaning "no filter, include all".
	resourceAll = "all"
)

var allExportResources = []string{
	exportResourceChores,
	exportResourceRewards,
	exportResourceLists,
	exportResourceRecipes,
	exportResourceSittings,
	exportResourceCalendar,
}

type ExportData struct {
	ExportedAt     string              `json:"exported_at"`
	FrameID        string              `json:"frame_id"`
	Chores         []lib.Chore         `json:"chores,omitempty"`
	Rewards        []lib.Reward        `json:"rewards,omitempty"`
	Lists          []lib.List          `json:"lists,omitempty"`
	Recipes        []lib.Recipe        `json:"recipes,omitempty"`
	MealSittings   []lib.MealSitting   `json:"meal_sittings,omitempty"`
	CalendarEvents []lib.CalendarEvent `json:"calendar_events,omitempty"`
}

var (
	exportOutputFile string
	exportResources  string
	exportDays       int
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Dump frame data to a JSON file",
	Long: `Export frame data to a JSON file for backup or migration.

Resources exported: chores, rewards, lists, recipes, sittings, calendar.
Time-bounded resources (chores, sittings, calendar) use --days to set the window
centered on today. Use --resources to limit which resource types are included.`,
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()
		client := getClient()

		resources := parseResourceList(exportResources, allExportResources)
		now := time.Now()
		start := now.AddDate(0, 0, -exportDays).Format(lib.DateFormat)
		end := now.AddDate(0, 0, exportDays).Format(lib.DateFormat)

		frame := getFrameOrFail(client, frameID)

		data := ExportData{
			ExportedAt: now.Format(time.RFC3339),
			FrameID:    frameID,
		}

		type result struct {
			name string
			err  error
		}
		results := make(chan result, len(resources))

		var mu sync.Mutex
		var wg sync.WaitGroup

		// launch runs fn in a goroutine, recovering any panic as an error.
		// It sends the result before decrementing wg so the closer never
		// closes the channel before all sends complete.
		launch := func(name string, fn func() error) {
			wg.Add(1)
			go func() {
				var err error
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic: %v", r)
					}
					results <- result{name, err}
					wg.Done()
				}()
				err = fn()
			}()
		}

		want := toWantMap(resources)

		if want[exportResourceChores] {
			launch(exportResourceChores, func() error {
				chores, err := client.ListChores(frameID, lib.ChoreListOptions{After: start, Before: end, IncludeLate: true})
				if err == nil {
					mu.Lock()
					data.Chores = chores
					mu.Unlock()
				}
				return err
			})
		}
		if want[exportResourceRewards] {
			launch(exportResourceRewards, func() error {
				rewards, err := client.ListRewards(frameID)
				if err == nil {
					mu.Lock()
					data.Rewards = rewards
					mu.Unlock()
				}
				return err
			})
		}
		if want[exportResourceLists] {
			launch(exportResourceLists, func() error {
				lists, err := client.ListLists(frameID)
				if err == nil {
					mu.Lock()
					data.Lists = lists
					mu.Unlock()
				}
				return err
			})
		}
		if want[exportResourceRecipes] {
			launch(exportResourceRecipes, func() error {
				recipes, err := client.ListRecipes(frameID)
				if err == nil {
					mu.Lock()
					data.Recipes = recipes
					mu.Unlock()
				}
				return err
			})
		}
		if want[exportResourceSittings] {
			launch(exportResourceSittings, func() error {
				sittings, err := client.ListMealSittings(frameID, lib.MealSittingListOptions{DateMin: start, DateMax: end})
				if err == nil {
					mu.Lock()
					data.MealSittings = sittings
					mu.Unlock()
				}
				return err
			})
		}
		if want[exportResourceCalendar] {
			launch(exportResourceCalendar, func() error {
				events, err := client.ListCalendarEvents(frameID, start, end, frame.TimeZone)
				if err == nil {
					mu.Lock()
					data.CalendarEvents = events
					mu.Unlock()
				}
				return err
			})
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		var errs []string
		for r := range results {
			if r.err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", r.name, r.err))
			}
		}
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "Error exporting %s\n", e)
			}
			os.Exit(1)
		}

		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding export: %v\n", err)
			os.Exit(1)
		}

		if exportOutputFile == "" || exportOutputFile == "-" {
			fmt.Println(string(out))
			return
		}
		if err := os.WriteFile(exportOutputFile, out, 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", exportOutputFile, err)
			os.Exit(1)
		}
		fmt.Printf("Exported to %s\n", exportOutputFile)
	},
}

func toWantMap(resources []string) map[string]bool {
	m := make(map[string]bool, len(resources))
	for _, r := range resources {
		m[r] = true
	}
	return m
}

func parseExportResources(s string) []string {
	return parseResourceList(s, allExportResources)
}

func parseResourceList(s string, all []string) []string {
	if s == "" || s == resourceAll {
		return all
	}
	var out []string
	for _, r := range strings.Split(s, ",") {
		r = strings.TrimSpace(r)
		if r != "" {
			out = append(out, r)
		}
	}
	return out
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportOutputFile, "output-file", "", "Output file path (default: stdout)")
	exportCmd.Flags().StringVar(&exportResources, "resources", resourceAll, "Comma-separated resource types: chores,rewards,lists,recipes,sittings,calendar")
	exportCmd.Flags().IntVar(&exportDays, "days", 90, "Window (in days before/after today) for time-bounded resources")
}
