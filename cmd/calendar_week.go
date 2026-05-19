package cmd

import (
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
