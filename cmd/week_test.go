package cmd

import (
	"testing"
)

type item struct {
	date  string
	label string
}

func TestBuildWeekSlots_SlotCount(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	slots := buildWeekSlots([]item{}, monday,
		func(i item) string { return i.date },
		func(a, b item) bool { return a.label < b.label },
	)
	if len(slots) != 7 {
		t.Fatalf("expected 7 slots, got %d", len(slots))
	}
	if slots[0].day != "Mon" {
		t.Errorf("slot[0]: want Mon, got %s", slots[0].day)
	}
	if slots[6].day != "Sun" {
		t.Errorf("slot[6]: want Sun, got %s", slots[6].day)
	}
}

func TestBuildWeekSlots_ItemsBinnedByDate(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	items := []item{
		{date: "2026-04-27", label: "A"},
		{date: "2026-04-29", label: "B"},
		{date: "2026-04-29", label: "C"},
	}
	slots := buildWeekSlots(items, monday,
		func(i item) string { return i.date },
		func(a, b item) bool { return a.label < b.label },
	)
	if len(slots[0].items) != 1 || slots[0].items[0].label != "A" {
		t.Errorf("Mon: want [A], got %v", slots[0].items)
	}
	if len(slots[2].items) != 2 {
		t.Errorf("Wed: want 2 items, got %d", len(slots[2].items))
	}
}

func TestBuildWeekSlots_DatetimeStringTruncated(t *testing.T) {
	// Items with a full RFC3339 timestamp must still be binned by the date prefix.
	monday, _ := weekStart("2026-04-27")
	items := []item{
		{date: "2026-04-28T08:00:00Z", label: "morning"},
		{date: "2026-04-28T20:00:00Z", label: "evening"},
	}
	slots := buildWeekSlots(items, monday,
		func(i item) string { return i.date },
		func(a, b item) bool { return a.label < b.label },
	)
	if len(slots[1].items) != 2 {
		t.Errorf("Tue: expected 2 items from datetime strings, got %d", len(slots[1].items))
	}
}

func TestBuildWeekSlots_SortedWithinSlot(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	items := []item{
		{date: "2026-04-27", label: "Zebra"},
		{date: "2026-04-27", label: "Apple"},
		{date: "2026-04-27", label: "Mango"},
	}
	slots := buildWeekSlots(items, monday,
		func(i item) string { return i.date },
		func(a, b item) bool { return a.label < b.label },
	)
	got := slots[0].items
	if len(got) != 3 || got[0].label != "Apple" || got[1].label != "Mango" || got[2].label != "Zebra" {
		t.Errorf("Mon: want [Apple Mango Zebra], got %v", got)
	}
}

func TestBuildWeekSlots_OutOfRangeDropped(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	items := []item{
		{date: "2026-04-20", label: "past"},   // before the week
		{date: "2026-05-10", label: "future"}, // after the week
	}
	slots := buildWeekSlots(items, monday,
		func(i item) string { return i.date },
		func(a, b item) bool { return a.label < b.label },
	)
	for _, s := range slots {
		if len(s.items) != 0 {
			t.Errorf("slot %s: expected empty, got %v", s.day, s.items)
		}
	}
}

func TestBuildWeekSlots_SlotDatesCorrect(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	slots := buildWeekSlots([]item{}, monday,
		func(i item) string { return i.date },
		func(a, b item) bool { return false },
	)
	wantDates := []string{
		"2026-04-27", "2026-04-28", "2026-04-29",
		"2026-04-30", "2026-05-01", "2026-05-02", "2026-05-03",
	}
	for i, want := range wantDates {
		if slots[i].date != want {
			t.Errorf("slot[%d].date: want %s, got %s", i, want, slots[i].date)
		}
	}
}

func TestBuildWeekSlots_DisplayFormat(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	slots := buildWeekSlots([]item{}, monday,
		func(i item) string { return i.date },
		func(a, b item) bool { return false },
	)
	if slots[0].display != "Apr 27" {
		t.Errorf("slot[0].display: want Apr 27, got %s", slots[0].display)
	}
}
