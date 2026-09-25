package cmd

import (
	"sort"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

type ScheduleDay struct {
	Day     string              `json:"day"`
	Date    string              `json:"date"`
	Display string              `json:"-"`
	Events  []lib.CalendarEvent `json:"events"`
}

func buildCalendarScheduleView(events []lib.CalendarEvent, start time.Time, days int) []ScheduleDay {
	byDate := make(map[string][]lib.CalendarEvent, days)
	for _, e := range events {
		if len(e.StartAt) < 10 {
			continue
		}
		key := e.StartAt[:10]
		byDate[key] = append(byDate[key], e)
	}

	slots := make([]ScheduleDay, days)
	for i := range days {
		d := start.AddDate(0, 0, i)
		key := d.Format(lib.DateFormat)
		evts := byDate[key]
		if evts == nil {
			evts = []lib.CalendarEvent{}
		} else {
			sort.Slice(evts, func(a, b int) bool { return evts[a].StartAt < evts[b].StartAt })
		}
		slots[i] = ScheduleDay{
			Day:     d.Format("Mon"),
			Date:    key,
			Display: d.Format("Jan 02"),
			Events:  evts,
		}
	}
	return slots
}
