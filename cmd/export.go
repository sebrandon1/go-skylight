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
	exportResourceChores     = "chores"
	exportResourceRewards    = "rewards"
	exportResourceLists      = "lists"
	exportResourceRecipes    = "recipes"
	exportResourceSittings   = "sittings"
	exportResourceCalendar   = "calendar"
	exportResourceRoutines   = "routines"
	exportResourceBounties   = "bounties"
	exportResourceCategories = "categories"

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
	exportResourceRoutines,
	exportResourceBounties,
	exportResourceCategories,
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
	Routines       []lib.Routine       `json:"routines,omitempty"`
	Bounties       []lib.Bounty        `json:"bounties,omitempty"`
	Categories     []lib.Category      `json:"categories,omitempty"`
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

Resources exported: chores, rewards, lists, recipes, sittings, calendar, routines, bounties, categories.
Time-bounded resources (chores, sittings, calendar) use --days to set the window
centered on today. Use --resources to limit which resource types are included.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		resources, err := parseResourceFilter(exportResources, allExportResources)
		if err != nil {
			return err
		}
		now := time.Now()
		start := now.AddDate(0, 0, -exportDays).Format(lib.DateFormat)
		end := now.AddDate(0, 0, exportDays).Format(lib.DateFormat)

		ctx := cmd.Context()
		frame, err := getFrameOrFail(ctx, client, frameID)
		if err != nil {
			return err
		}

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
				chores, err := client.ListChores(ctx, frameID, lib.ChoreListOptions{After: start, Before: end, IncludeLate: true})
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
				rewards, err := client.ListRewards(ctx, frameID)
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
				lists, err := client.ListLists(ctx, frameID)
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
				recipes, err := client.ListRecipes(ctx, frameID)
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
				sittings, err := client.ListMealSittings(ctx, frameID, lib.MealSittingListOptions{DateMin: start, DateMax: end})
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
				events, err := client.ListCalendarEvents(ctx, frameID, start, end, frame.TimeZone)
				if err == nil {
					mu.Lock()
					data.CalendarEvents = events
					mu.Unlock()
				}
				return err
			})
		}
		if want[exportResourceRoutines] {
			launch(exportResourceRoutines, func() error {
				routines, err := client.ListRoutines(ctx, frameID)
				if err == nil {
					mu.Lock()
					data.Routines = routines
					mu.Unlock()
				}
				return err
			})
		}
		if want[exportResourceBounties] {
			launch(exportResourceBounties, func() error {
				bounties, err := client.ListBounties(ctx, frameID)
				if err == nil {
					mu.Lock()
					data.Bounties = bounties
					mu.Unlock()
				}
				return err
			})
		}
		if want[exportResourceCategories] {
			launch(exportResourceCategories, func() error {
				categories, err := client.ListCategories(ctx, frameID)
				if err == nil {
					mu.Lock()
					data.Categories = categories
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
			return fmt.Errorf("export failed: %s", strings.Join(errs, "; "))
		}

		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding export: %w", err)
		}

		if exportOutputFile == "" || exportOutputFile == "-" {
			fmt.Println(string(out))
			return nil
		}
		if err := os.WriteFile(exportOutputFile, out, 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", exportOutputFile, err)
		}
		fmt.Printf("Exported to %s\n", exportOutputFile)
		return nil
	},
}

func toWantMap(resources []string) map[string]bool {
	m := make(map[string]bool, len(resources))
	for _, r := range resources {
		m[r] = true
	}
	return m
}

func parseExportResources(s string) ([]string, error) {
	return parseResourceFilter(s, allExportResources)
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportOutputFile, "output-file", "", "Output file path (default: stdout)")
	exportCmd.Flags().StringVar(&exportResources, "resources", resourceAll, "Comma-separated resource types: chores,rewards,lists,recipes,sittings,calendar,routines,bounties,categories")
	exportCmd.Flags().IntVar(&exportDays, "days", 90, "Window (in days before/after today) for time-bounded resources")
}
