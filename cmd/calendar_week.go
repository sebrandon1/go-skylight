package cmd

import (
	"fmt"
	"os"
	"sort"
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
// Events are grouped by the UTC date of their StartAt field; out-of-range events are silently dropped.
func buildCalendarWeeklyView(events []lib.CalendarEvent, monday time.Time) []WeeklyCalendarDay {
	days := make([]WeeklyCalendarDay, 7)
	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		days[i] = WeeklyCalendarDay{
			Day:     d.Format("Mon"),
			Date:    d.Format("2006-01-02"),
			Display: d.Format("Jan 02"),
			Events:  []lib.CalendarEvent{},
		}
	}

	byDate := make(map[string][]lib.CalendarEvent, 7)
	for _, e := range events {
		date := e.StartAt
		if len(date) >= 10 {
			date = date[:10]
		}
		byDate[date] = append(byDate[date], e)
	}

	for i, day := range days {
		if evs, ok := byDate[day.Date]; ok {
			sort.Slice(evs, func(a, b int) bool {
				return evs[a].StartAt < evs[b].StartAt
			})
			days[i].Events = evs
		}
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
