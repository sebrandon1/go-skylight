package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

type WeeklyCalendarDay struct {
	Day     string              `json:"day"`
	Date    string              `json:"date"`
	Display string              `json:"-"`
	Events  []lib.CalendarEvent `json:"events"`
}

// buildCalendarWeeklyView groups events into Mon–Sun slots for the week starting on monday.
// Events are grouped by the date of their StartAt field; out-of-range events are silently dropped.
func buildCalendarWeeklyView(events []lib.CalendarEvent, monday time.Time) []WeeklyCalendarDay {
	slots := buildWeekSlots(events, monday,
		func(e lib.CalendarEvent) string { return e.StartAt },
		func(a, b lib.CalendarEvent) bool { return a.StartAt < b.StartAt },
	)
	days := make([]WeeklyCalendarDay, len(slots))
	for i, s := range slots {
		days[i] = WeeklyCalendarDay{Day: s.day, Date: s.date, Display: s.display, Events: s.items}
	}
	return days
}

func printCalendarWeekTable(days []WeeklyCalendarDay) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DAY\tDATE\tTITLE\tTIME\tALL DAY")
	for _, d := range days {
		if len(d.Events) == 0 {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.Day, d.Display, "(no events)", "—", "—")
			continue
		}
		for i, e := range d.Events {
			dayCol, dateCol := d.Day, d.Display
			if i > 0 {
				dayCol, dateCol = "", ""
			}
			timeCol := "All day"
			if !e.AllDay {
				timeCol = "—"
				if len(e.StartAt) >= 16 {
					timeCol = e.StartAt[11:16]
				}
			}
			allDayCol := boolNo
			if e.AllDay {
				allDayCol = boolYes
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", dayCol, dateCol, e.Title, timeCol, allDayCol)
		}
	}
	w.Flush()
}
