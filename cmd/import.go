package cmd

import (
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

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Restore frame data from an export file",
	Long: `Restore frame data from a JSON file produced by the export command.

Each resource type is created in the target frame. IDs from the source frame are
ignored — new IDs are assigned by the API. Use --resources to import only specific
types. Use --dry-run to preview what would be created without making API calls.`,
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		raw, err := os.ReadFile(importFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", importFile, err)
			os.Exit(1)
		}

		var data ExportData
		if err := json.Unmarshal(raw, &data); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing export file: %v\n", err)
			os.Exit(1)
		}

		resources := parseResourceList(importResources, allExportResources)
		want := toWantMap(resources)

		if importDryRun {
			runImportDryRun(data, want)
			return
		}

		client := getClient()
		runImport(client, data, want)
	},
}

func runImportDryRun(data ExportData, want map[string]bool) {
	counts := map[string]int{
		exportResourceChores:   len(data.Chores),
		exportResourceRewards:  len(data.Rewards),
		exportResourceLists:    len(data.Lists),
		exportResourceRecipes:  len(data.Recipes),
		exportResourceSittings: len(data.MealSittings),
		exportResourceCalendar: len(data.CalendarEvents),
	}
	fmt.Printf("Dry run — would import into frame %s:\n", frameID)
	for _, r := range allExportResources {
		if want[r] {
			fmt.Printf("  %-10s %d items\n", r, counts[r])
		}
	}
}

func runImport(client *lib.Client, data ExportData, want map[string]bool) {
	var total, failed int
	add := func(t, f int) { total += t; failed += f }

	if want[exportResourceRewards] {
		add(importRewards(client, data.Rewards))
	}
	if want[exportResourceChores] {
		add(importChores(client, data.Chores))
	}
	if want[exportResourceLists] {
		add(importLists(client, data.Lists))
	}
	if want[exportResourceRecipes] {
		add(importRecipes(client, data.Recipes))
	}
	if want[exportResourceSittings] {
		add(importSittings(client, data.MealSittings))
	}
	if want[exportResourceCalendar] {
		add(importCalendarEvents(client, data.CalendarEvents))
	}

	fmt.Printf("Imported %d/%d items successfully.\n", total-failed, total)
	if failed > 0 {
		os.Exit(1)
	}
}

func importRewards(client *lib.Client, rewards []lib.Reward) (total, failed int) {
	for _, r := range rewards {
		total++
		if _, err := client.CreateReward(frameID, lib.RewardData{Title: r.Title, Points: r.Points, EmojiIcon: r.EmojiIcon}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating reward %q: %v\n", r.Title, err)
			failed++
		}
	}
	return
}

func importChores(client *lib.Client, chores []lib.Chore) (total, failed int) {
	for _, c := range chores {
		total++
		if _, err := client.CreateChore(frameID, lib.ChoreData{Title: c.Title, DueDate: c.DueDate, Points: c.Points, AssigneeID: c.AssigneeID}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating chore %q: %v\n", c.Title, err)
			failed++
		}
	}
	return
}

func importLists(client *lib.Client, lists []lib.List) (total, failed int) {
	for _, l := range lists {
		total++
		created, err := client.CreateList(frameID, lib.ListData{Title: l.Title, Color: l.Color, Kind: l.Kind})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating list %q: %v\n", l.Title, err)
			failed++
			continue
		}
		for _, item := range l.Items {
			total++
			if _, err := client.AddListItem(frameID, created.ID, lib.ListItemData{Title: item.Title}); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding item %q to list %q: %v\n", item.Title, l.Title, err)
				failed++
			}
		}
	}
	return
}

func importRecipes(client *lib.Client, recipes []lib.Recipe) (total, failed int) {
	for _, r := range recipes {
		total++
		if _, err := client.CreateRecipe(frameID, lib.RecipeData{Title: r.Title, Description: r.Description, Ingredients: r.Ingredients, URL: r.URL}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating recipe %q: %v\n", r.Title, err)
			failed++
		}
	}
	return
}

func importSittings(client *lib.Client, sittings []lib.MealSitting) (total, failed int) {
	for _, s := range sittings {
		total++
		if _, err := client.CreateMealSitting(frameID, lib.MealSittingData{Summary: s.Summary, Date: s.Date}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating meal sitting %q: %v\n", s.Summary, err)
			failed++
		}
	}
	return
}

func importCalendarEvents(client *lib.Client, events []lib.CalendarEvent) (total, failed int) {
	for _, e := range events {
		total++
		if _, err := client.CreateCalendarEvent(frameID, lib.CalendarEventData{Title: e.Title, StartAt: e.StartAt, EndAt: e.EndAt, AllDay: e.AllDay}); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating calendar event %q: %v\n", e.Title, err)
			failed++
		}
	}
	return
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&importFile, "file", "", "Path to export JSON file")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview what would be imported without making API calls")
	importCmd.Flags().StringVar(&importResources, "resources", resourceAll, "Comma-separated resource types to import")
	markFlagRequired(importCmd, "file")
}
