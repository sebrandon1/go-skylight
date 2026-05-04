package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

type weekSlot[T any] struct {
	day     string
	date    string
	display string
	items   []T
}

// weekStart returns the Monday of the week containing date.
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

// buildWeekSlots creates 7 Mon–Sun slots, groups items by date, and sorts each
// slot. dateOf extracts the ISO date from an item (truncated to 10 chars when
// longer). less is the sort comparator applied within each slot.
func buildWeekSlots[T any](items []T, monday time.Time, dateOf func(T) string, less func(a, b T) bool) []weekSlot[T] {
	slots := make([]weekSlot[T], 7)
	for i := range slots {
		d := monday.AddDate(0, 0, i)
		slots[i] = weekSlot[T]{
			day:     d.Format("Mon"),
			date:    d.Format(lib.DateFormat),
			display: d.Format("Jan 02"),
		}
	}

	byDate := make(map[string][]T, 7)
	for _, item := range items {
		date := dateOf(item)
		if len(date) >= 10 {
			date = date[:10]
		}
		byDate[date] = append(byDate[date], item)
	}

	for i, s := range slots {
		if its, ok := byDate[s.date]; ok {
			sort.Slice(its, func(a, b int) bool { return less(its[a], its[b]) })
			slots[i].items = its
		} else {
			slots[i].items = []T{}
		}
	}

	return slots
}
