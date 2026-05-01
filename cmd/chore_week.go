package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
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

// weekStart returns the Monday of the week containing the given date string.
// "current" or "" resolves to the current week.
func weekStart(date string) (time.Time, error) {
	var t time.Time
	if date == "" || date == "current" {
		t = time.Now()
	} else {
		var err error
		t, err = time.Parse(lib.DateFormat, date)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date %q: use YYYY-MM-DD format", date)
		}
	}
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7 // Sunday = 7 in ISO week
	}
	monday := t.AddDate(0, 0, -(wd - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC), nil
}

// buildWeeklyView groups chores into Mon–Sun slots for the week starting on monday.
func buildWeeklyView(chores []lib.Chore, monday time.Time) []WeeklyChoreDay {
	days := make([]WeeklyChoreDay, 7)
	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		days[i] = WeeklyChoreDay{
			Day:     d.Format("Mon"),
			Date:    d.Format(lib.DateFormat),
			Display: d.Format("Jan 02"),
			Chores:  []lib.Chore{},
		}
	}

	byDate := make(map[string][]lib.Chore, 7)
	for _, c := range chores {
		date := c.DueDate
		if len(date) >= 10 {
			date = date[:10]
		}
		byDate[date] = append(byDate[date], c)
	}

	for i, day := range days {
		if cs, ok := byDate[day.Date]; ok {
			sort.Slice(cs, func(a, b int) bool { return cs[a].Title < cs[b].Title })
			days[i].Chores = cs
		}
	}

	return days
}

func printChoreWeekTable(days []WeeklyChoreDay) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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
