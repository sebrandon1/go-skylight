package cmd

import (
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
