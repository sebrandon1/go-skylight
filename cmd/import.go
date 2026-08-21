package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	importFile      string
	importDryRun    bool
	importResources string
)

// importWorkerCount bounds how many items within a single resource type are
// created concurrently during import.
const importWorkerCount = 5

// parallelImport runs fn for each item in items using a bounded pool of
// importWorkerCount goroutines, aggregating each call's (total, failed)
// counts. Item order in stderr output is not preserved.
func parallelImport[T any](items []T, fn func(T) (total, failed int)) (total, failed int) {
	type result struct{ total, failed int }
	results := make(chan result, len(items))
	sem := make(chan struct{}, importWorkerCount)
	for _, item := range items {
		sem <- struct{}{}
		go func(item T) {
			var t, f int
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "panic in import goroutine: %v\n", r)
					t, f = 0, 1
				}
				results <- result{t, f}
				<-sem
			}()
			t, f = fn(item)
		}(item)
	}
	for range items {
		r := <-results
		total += r.total
		failed += r.failed
	}
	return total, failed
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Restore frame data from an export file",
	Long: `Restore frame data from a JSON file produced by the export command.

Each resource type is created in the target frame. IDs from the source frame are
ignored — new IDs are assigned by the API. Use --resources to import only specific
types. Use --dry-run to preview what would be created without making API calls.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		raw, err := os.ReadFile(importFile)
		if err != nil {
			return fmt.Errorf("reading %s: %w", importFile, err)
		}

		var data ExportData
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing export file: %w", err)
		}

		resources, err := parseResourceFilter(importResources, allExportResources)
		if err != nil {
			return err
		}
		want := toWantMap(resources)

		if importDryRun {
			runImportDryRun(data, want)
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}
		return runImport(cmd.Context(), client, data, want)
	},
}

// nonRoutineChoreCount excludes Routine==true chores, matching importChores'
// behavior: those are imported separately via importRoutines, so counting
// them here would overstate how many plain chores will actually be created.
func nonRoutineChoreCount(chores []lib.Chore) int {
	n := 0
	for _, c := range chores {
		if !c.Routine {
			n++
		}
	}
	return n
}

func runImportDryRun(data ExportData, want map[string]bool) {
	counts := map[string]int{
		exportResourceChores:     nonRoutineChoreCount(data.Chores),
		exportResourceRewards:    len(data.Rewards),
		exportResourceLists:      len(data.Lists),
		exportResourceRecipes:    len(data.Recipes),
		exportResourceSittings:   len(data.MealSittings),
		exportResourceCalendar:   len(data.CalendarEvents),
		exportResourceRoutines:   len(data.Routines),
		exportResourceBounties:   len(data.Bounties),
		exportResourceCategories: len(data.Categories),
		exportResourcePhotos:     len(data.Photos),
	}
	fmt.Printf("Dry run — would import into frame %s:\n", frameID)
	for _, r := range allExportResources {
		if want[r] && r != exportResourceCategories {
			fmt.Printf("  %-10s %d items\n", r, counts[r])
		}
	}
}

func runImport(ctx context.Context, client *lib.Client, data ExportData, want map[string]bool) error {
	type importFn = func() (int, int)
	var tasks []importFn

	if want[exportResourceRewards] {
		tasks = append(tasks, func() (int, int) { return importRewards(ctx, client, data.Rewards) })
	}
	if want[exportResourceChores] {
		tasks = append(tasks, func() (int, int) { return importChores(ctx, client, data.Chores) })
	}
	if want[exportResourceLists] {
		tasks = append(tasks, func() (int, int) { return importLists(ctx, client, data.Lists) })
	}
	if want[exportResourceRecipes] {
		tasks = append(tasks, func() (int, int) { return importRecipes(ctx, client, data.Recipes) })
	}
	if want[exportResourceSittings] {
		tasks = append(tasks, func() (int, int) { return importSittings(ctx, client, data.MealSittings) })
	}
	if want[exportResourceCalendar] {
		tasks = append(tasks, func() (int, int) { return importCalendarEvents(ctx, client, data.CalendarEvents) })
	}
	if want[exportResourceRoutines] {
		tasks = append(tasks, func() (int, int) { return importRoutines(ctx, client, data.Routines) })
	}
	if want[exportResourceBounties] {
		tasks = append(tasks, func() (int, int) { return importBounties(ctx, client, data.Bounties) })
	}
	if want[exportResourcePhotos] {
		tasks = append(tasks, func() (int, int) { return importPhotos(ctx, client, data.Photos) })
	}

	total, failed := parallelImport(tasks, func(fn importFn) (int, int) { return fn() })

	printSuccessf("Imported %d/%d items successfully.\n", total-failed, total)
	if failed > 0 {
		return fmt.Errorf("%d/%d items failed to import", failed, total)
	}
	return nil
}

func importRewards(ctx context.Context, client *lib.Client, rewards []lib.Reward) (total, failed int) {
	return parallelImport(rewards, func(r lib.Reward) (int, int) {
		if _, err := client.CreateReward(ctx, frameID, lib.RewardData{Title: r.Title, Points: r.Points, EmojiIcon: r.EmojiIcon}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating reward %q: %v\n", r.Title, err)
			return 1, 1
		}
		return 1, 0
	})
}

// importChores skips chores with Routine == true: a routine is a chore, so
// it appears in both the plain chore export and the routine export, and
// importRoutines already handles it separately. Without this, a round-trip
// export/import would create each routine twice -- once as a plain
// non-recurring chore, once as the correct routine.
func importChores(ctx context.Context, client *lib.Client, chores []lib.Chore) (total, failed int) {
	return parallelImport(chores, func(c lib.Chore) (int, int) {
		if c.Routine {
			fmt.Fprintf(os.Stderr, "Skipping routine chore %q (import routines separately with --resources routines)\n", c.Title)
			return 0, 0
		}
		if _, err := client.CreateChore(ctx, frameID, lib.ChoreData{Title: c.Title, DueDate: c.DueDate, Points: c.Points, AssigneeID: c.AssigneeID}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating chore %q: %v\n", c.Title, err)
			return 1, 1
		}
		return 1, 0
	})
}

// importLists parallelizes across lists, but each list's own items are
// created sequentially after it (AddListItem depends on the parent list's
// freshly assigned ID), so items are never parallelized against each other.
func importLists(ctx context.Context, client *lib.Client, lists []lib.List) (total, failed int) {
	return parallelImport(lists, func(l lib.List) (int, int) {
		t, f := 1, 0
		created, err := client.CreateList(ctx, frameID, lib.ListData{Title: l.Title, Color: l.Color, Kind: l.Kind})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating list %q: %v\n", l.Title, err)
			return t, 1
		}
		for _, item := range l.Items {
			t++
			if _, err := client.AddListItem(ctx, frameID, created.ID, lib.ListItemData{Title: item.Title}); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding item %q to list %q: %v\n", item.Title, l.Title, err)
				f++
			}
		}
		return t, f
	})
}

func importRecipes(ctx context.Context, client *lib.Client, recipes []lib.Recipe) (total, failed int) {
	return parallelImport(recipes, func(r lib.Recipe) (int, int) {
		if _, err := client.CreateRecipe(ctx, frameID, lib.RecipeData{Title: r.Title, Description: r.Description, Ingredients: r.Ingredients, URL: r.URL}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating recipe %q: %v\n", r.Title, err)
			return 1, 1
		}
		return 1, 0
	})
}

func importSittings(ctx context.Context, client *lib.Client, sittings []lib.MealSitting) (total, failed int) {
	return parallelImport(sittings, func(s lib.MealSitting) (int, int) {
		if _, err := client.CreateMealSitting(ctx, frameID, lib.MealSittingData{Summary: s.Summary, Date: s.Date}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating meal sitting %q: %v\n", s.Summary, err)
			return 1, 1
		}
		return 1, 0
	})
}

func importCalendarEvents(ctx context.Context, client *lib.Client, events []lib.CalendarEvent) (total, failed int) {
	return parallelImport(events, func(e lib.CalendarEvent) (int, int) {
		allDay := e.AllDay
		if _, err := client.CreateCalendarEvent(ctx, frameID, lib.CalendarEventData{Title: e.Title, StartAt: e.StartAt, EndAt: e.EndAt, AllDay: &allDay}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating calendar event %q: %v\n", e.Title, err)
			return 1, 1
		}
		return 1, 0
	})
}

func importRoutines(ctx context.Context, client *lib.Client, routines []lib.Routine) (total, failed int) {
	return parallelImport(routines, func(r lib.Routine) (int, int) {
		data := lib.RoutineData{Title: r.Title, TimeOfDay: r.TimeOfDay, CategoryID: r.AssigneeID, StartDate: r.NextOccurrenceDate}
		if _, err := client.CreateRoutine(ctx, frameID, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating routine %q: %v\n", r.Title, err)
			return 1, 1
		}
		return 1, 0
	})
}

func importBounties(ctx context.Context, client *lib.Client, bounties []lib.Bounty) (total, failed int) {
	return parallelImport(bounties, func(b lib.Bounty) (int, int) {
		if _, err := client.CreateBounty(ctx, frameID, lib.BountyData{
			Title:       b.Chore.Title,
			Points:      b.Chore.Points,
			DueDate:     b.Chore.DueDate,
			RewardTitle: b.Reward.Title,
			EmojiIcon:   b.Reward.EmojiIcon,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating bounty %q: %v\n", b.Chore.Title, err)
			return 1, 1
		}
		return 1, 0
	})
}

func importPhotos(ctx context.Context, client *lib.Client, photos []PhotoExport) (total, failed int) {
	return parallelImport(photos, func(p PhotoExport) (int, int) {
		raw, err := base64.StdEncoding.DecodeString(p.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding photo data: %v\n", err)
			return 1, 1
		}
		if _, err := client.UploadPhoto(ctx, frameID, p.Ext, raw, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading photo: %v\n", err)
			return 1, 1
		}
		return 1, 0
	})
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&importFile, "file", "", "Path to export JSON file")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview what would be imported without making API calls")
	importCmd.Flags().StringVar(&importResources, "resources", resourceAll, "Comma-separated resource types to import")
	markFlagRequired(importCmd, "file")
}
