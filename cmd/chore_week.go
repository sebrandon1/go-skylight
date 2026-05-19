package cmd

import (
	"fmt"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

// WeeklyChoreDay represents one day's chores in the weekly view.
type WeeklyChoreDay struct {
	Day     string      `json:"day"`
	Date    string      `json:"date"`
	Display string      `json:"-"`
	Chores  []lib.Chore `json:"chores"`
}

// buildWeeklyView groups chores into Mon–Sun slots for the week starting on monday.
func buildWeeklyView(chores []lib.Chore, monday time.Time) []WeeklyChoreDay {
	slots := buildWeekSlots(chores, monday,
		func(c lib.Chore) string { return c.DueDate },
		func(a, b lib.Chore) bool { return a.Title < b.Title },
	)
	days := make([]WeeklyChoreDay, len(slots))
	for i, s := range slots {
		days[i] = WeeklyChoreDay{Day: s.day, Date: s.date, Display: s.display, Chores: s.items}
	}
	return days
}

func printChoreWeekTable(days []WeeklyChoreDay) {
	w := newTableWriter()
	fmt.Fprintln(w, "DAY\tDATE\tTITLE\tSTATUS")
	for _, d := range days {
		if len(d.Chores) == 0 {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.Day, d.Display, "(no chores)", "—")
			continue
		}
		for i, c := range d.Chores {
			dayCol, dateCol := d.Day, d.Display
			if i > 0 {
				dayCol, dateCol = "", ""
			}
			status := "✗"
			if c.Status == lib.ChoreStatusComplete {
				status = "✓"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", dayCol, dateCol, c.Title, status)
		}
	}
	w.Flush()
}
