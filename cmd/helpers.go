package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/sebrandon1/go-skylight/lib"
)

func printJSON(data any) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

// printOutput prints data in the format specified by --output (json or table).
// Defaults to JSON for types that don't have a dedicated table renderer.
func printOutput(data any) {
	if outputFormat == outputTable {
		switch v := data.(type) {
		case []lib.Chore:
			printChoresTable(v)
			return
		case []lib.Reward:
			printRewardsTable(v)
			return
		case []lib.Frame:
			printFramesTable(v)
			return
		case []lib.CalendarEvent:
			printCalendarTable(v)
			return
		case []lib.Category:
			printCategoriesTable(v)
			return
		case []ChoreStreakStats:
			printChoreStreakTable(v)
			return
		case []WeeklyChoreDay:
			printChoreWeekTable(v)
			return
		}
	}
	printJSON(data)
}

func printChoresTable(chores []lib.Chore) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tDUE DATE\tPOINTS\tASSIGNEE")
	for _, c := range chores {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
			c.ID, c.Title, c.Status, c.DueDate, c.Points, c.AssigneeID)
	}
	w.Flush()
}

func printRewardsTable(rewards []lib.Reward) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tPOINTS\tEMOJI\tREDEEMED\tCATEGORY")
	for _, r := range rewards {
		redeemed := "no"
		if r.Redeemed {
			redeemed = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			r.ID, r.Title, r.Points, r.EmojiIcon, redeemed, r.CategoryID)
	}
	w.Flush()
}

func printFramesTable(frames []lib.Frame) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTIMEZONE")
	for _, f := range frames {
		fmt.Fprintf(w, "%s\t%s\t%s\n", f.ID, f.Name, f.TimeZone)
	}
	w.Flush()
}

func printCalendarTable(events []lib.CalendarEvent) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTART\tEND\tALL DAY")
	for _, e := range events {
		allDay := "no"
		if e.AllDay {
			allDay = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			e.ID, e.Title, e.StartAt, e.EndAt, allDay)
	}
	w.Flush()
}

func printCategoriesTable(cats []lib.Category) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCOLOR")
	for _, c := range cats {
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.ID, c.Name, c.Color)
	}
	w.Flush()
}

func printChoreStreakTable(stats []ChoreStreakStats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCURRENT STREAK\tLONGEST STREAK\tCOMPLETED\tTOTAL\tRATE")
	for _, s := range stats {
		fmt.Fprintf(w, "%s\t%d days\t%d days\t%d\t%d\t%.1f%%\n",
			s.AssigneeName, s.CurrentStreak, s.LongestStreak, s.CompletedChores, s.TotalChores, s.CompletionRate)
	}
	w.Flush()
}
