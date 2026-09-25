package cmd

import (
	"testing"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestBuildCalendarScheduleView_SlotCount(t *testing.T) {
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	days := buildCalendarScheduleView(nil, start, 5)
	if len(days) != 5 {
		t.Errorf("want 5 days, got %d", len(days))
	}
	if days[0].Day != "Fri" {
		t.Errorf("first slot should be Fri, got %s", days[0].Day)
	}
	if days[4].Day != "Tue" {
		t.Errorf("last slot should be Tue, got %s", days[4].Day)
	}
}

func TestBuildCalendarScheduleView_DatesAreConsecutive(t *testing.T) {
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	days := buildCalendarScheduleView(nil, start, 3)
	want := []string{"2026-09-25", "2026-09-26", "2026-09-27"}
	for i, w := range want {
		if days[i].Date != w {
			t.Errorf("slot %d: want date %s, got %s", i, w, days[i].Date)
		}
	}
}

func TestBuildCalendarScheduleView_EventsGroupedByDate(t *testing.T) {
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	events := []lib.CalendarEvent{
		{ID: "1", Title: "Morning standup", StartAt: "2026-09-25T09:00:00Z"},
		{ID: "2", Title: "Afternoon sync", StartAt: "2026-09-25T14:00:00Z"},
		{ID: "3", Title: "Birthday party", StartAt: "2026-09-27T12:00:00Z"},
	}
	days := buildCalendarScheduleView(events, start, 3)

	if len(days[0].Events) != 2 {
		t.Errorf("day 0 (Sep 25): want 2 events, got %d", len(days[0].Events))
	}
	// Events sorted ascending: 09:00 before 14:00
	if days[0].Events[0].Title != "Morning standup" {
		t.Errorf("day 0 first event: want 'Morning standup', got %s", days[0].Events[0].Title)
	}
	if len(days[2].Events) != 1 {
		t.Errorf("day 2 (Sep 27): want 1 event, got %d", len(days[2].Events))
	}
}

func TestBuildCalendarScheduleView_EmptyDays(t *testing.T) {
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	days := buildCalendarScheduleView(nil, start, 3)
	for _, d := range days {
		if len(d.Events) != 0 {
			t.Errorf("day %s should have 0 events, got %d", d.Date, len(d.Events))
		}
	}
}

func TestBuildCalendarScheduleView_OutOfRangeEventsIgnored(t *testing.T) {
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	events := []lib.CalendarEvent{
		{ID: "1", Title: "Old event", StartAt: "2026-09-20T10:00:00Z"},
		{ID: "2", Title: "Future event", StartAt: "2026-10-10T10:00:00Z"},
	}
	days := buildCalendarScheduleView(events, start, 3)
	for _, d := range days {
		if len(d.Events) != 0 {
			t.Errorf("day %s should have 0 events, got %d", d.Date, len(d.Events))
		}
	}
}

func TestBuildCalendarScheduleView_ShortStartAtDropped(t *testing.T) {
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	events := []lib.CalendarEvent{
		{ID: "1", Title: "Malformed", StartAt: "bad"},
	}
	days := buildCalendarScheduleView(events, start, 1)
	if len(days[0].Events) != 0 {
		t.Errorf("event with StartAt < 10 chars should be dropped, got %d events", len(days[0].Events))
	}
}

func TestBuildCalendarScheduleView_OneDayWindow(t *testing.T) {
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	events := []lib.CalendarEvent{
		{ID: "1", Title: "Only event", StartAt: "2026-09-25T10:00:00Z"},
	}
	days := buildCalendarScheduleView(events, start, 1)
	if len(days) != 1 {
		t.Errorf("want 1 day, got %d", len(days))
	}
	if len(days[0].Events) != 1 {
		t.Errorf("want 1 event, got %d", len(days[0].Events))
	}
}
