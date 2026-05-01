package cmd

import (
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestBuildCalendarWeeklyView_SlotCount(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	days := buildCalendarWeeklyView(nil, monday)
	if len(days) != 7 {
		t.Errorf("want 7 days, got %d", len(days))
	}
	if days[0].Day != "Mon" {
		t.Errorf("first slot should be Mon, got %s", days[0].Day)
	}
	if days[6].Day != "Sun" {
		t.Errorf("last slot should be Sun, got %s", days[6].Day)
	}
}

func TestBuildCalendarWeeklyView_EventsGroupedByDate(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	events := []lib.CalendarEvent{
		{ID: "1", Title: "Early meeting", StartAt: "2026-04-27T08:00:00Z"},
		{ID: "2", Title: "Late meeting", StartAt: "2026-04-27T10:00:00Z"},
		{ID: "3", Title: "Wednesday lunch", StartAt: "2026-04-29T12:00:00Z"},
	}
	days := buildCalendarWeeklyView(events, monday)

	if len(days[0].Events) != 2 {
		t.Errorf("Mon: want 2 events, got %d", len(days[0].Events))
	}
	// Events sorted by StartAt ascending: 08:00 before 10:00
	if days[0].Events[0].Title != "Early meeting" {
		t.Errorf("Mon first event: want 'Early meeting', got %s", days[0].Events[0].Title)
	}
	if len(days[2].Events) != 1 {
		t.Errorf("Wed: want 1 event, got %d", len(days[2].Events))
	}
}

func TestBuildCalendarWeeklyView_EmptyDays(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	days := buildCalendarWeeklyView(nil, monday)
	for _, d := range days {
		if len(d.Events) != 0 {
			t.Errorf("day %s should have 0 events, got %d", d.Day, len(d.Events))
		}
	}
}

func TestBuildCalendarWeeklyView_OutOfRangeEventsIgnored(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	events := []lib.CalendarEvent{
		{ID: "99", Title: "Old event", StartAt: "2026-04-20T10:00:00Z"},
	}
	days := buildCalendarWeeklyView(events, monday)
	for _, d := range days {
		if len(d.Events) != 0 {
			t.Errorf("day %s should have 0 events, got %d", d.Day, len(d.Events))
		}
	}
}

func TestBuildCalendarWeeklyView_AllDayEvent(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	events := []lib.CalendarEvent{
		{ID: "1", Title: "Birthday", StartAt: "2026-04-28T00:00:00Z", AllDay: true},
	}
	days := buildCalendarWeeklyView(events, monday)
	if len(days[1].Events) != 1 {
		t.Errorf("Tue: want 1 all-day event, got %d", len(days[1].Events))
	}
	if !days[1].Events[0].AllDay {
		t.Error("expected AllDay to be true")
	}
}

func TestBuildCalendarWeeklyView_ShortStartAt(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	events := []lib.CalendarEvent{
		{ID: "1", Title: "Date-only event", StartAt: "2026-04-27"},
	}
	days := buildCalendarWeeklyView(events, monday)
	if len(days[0].Events) != 1 {
		t.Errorf("Mon: want 1 event for date-only StartAt, got %d", len(days[0].Events))
	}
}
