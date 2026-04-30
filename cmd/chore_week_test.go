package cmd

import (
	"testing"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestWeekStart_Current(t *testing.T) {
	monday, err := weekStart("current")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if monday.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", monday.Weekday())
	}
}

func TestWeekStart_Empty(t *testing.T) {
	monday, err := weekStart("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if monday.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", monday.Weekday())
	}
}

func TestWeekStart_KnownMonday(t *testing.T) {
	// 2026-04-27 is a Monday.
	monday, err := weekStart("2026-04-27")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if monday.Format("2006-01-02") != "2026-04-27" {
		t.Errorf("want 2026-04-27, got %s", monday.Format("2006-01-02"))
	}
}

func TestWeekStart_MidWeek(t *testing.T) {
	// 2026-04-29 is a Wednesday; its Monday is 2026-04-27.
	monday, err := weekStart("2026-04-29")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if monday.Format("2006-01-02") != "2026-04-27" {
		t.Errorf("want 2026-04-27, got %s", monday.Format("2006-01-02"))
	}
}

func TestWeekStart_Sunday(t *testing.T) {
	// 2026-05-03 is a Sunday; its Monday is 2026-04-27.
	monday, err := weekStart("2026-05-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if monday.Format("2006-01-02") != "2026-04-27" {
		t.Errorf("want 2026-04-27, got %s", monday.Format("2006-01-02"))
	}
}

func TestWeekStart_InvalidDate(t *testing.T) {
	_, err := weekStart("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date, got nil")
	}
}

func TestBuildWeeklyView_SlotCount(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	days := buildWeeklyView(nil, monday)
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

func TestBuildWeeklyView_ChoresGroupedByDate(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	chores := []lib.Chore{
		{ID: "1", Title: "Walk the dog", Status: "completed", DueDate: "2026-04-27"},
		{ID: "2", Title: "Take out trash", Status: "pending", DueDate: "2026-04-28"},
		{ID: "3", Title: "Feed the cat", Status: "completed", DueDate: "2026-04-28"},
	}
	days := buildWeeklyView(chores, monday)

	if len(days[0].Chores) != 1 {
		t.Errorf("Mon: want 1 chore, got %d", len(days[0].Chores))
	}
	if days[0].Chores[0].Title != "Walk the dog" {
		t.Errorf("Mon chore: want 'Walk the dog', got %s", days[0].Chores[0].Title)
	}
	if len(days[1].Chores) != 2 {
		t.Errorf("Tue: want 2 chores, got %d", len(days[1].Chores))
	}
	// Chores sorted alphabetically: "Feed the cat" < "Take out trash"
	if days[1].Chores[0].Title != "Feed the cat" {
		t.Errorf("Tue first chore: want 'Feed the cat', got %s", days[1].Chores[0].Title)
	}
}

func TestBuildWeeklyView_EmptyDays(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	days := buildWeeklyView(nil, monday)
	for _, d := range days {
		if len(d.Chores) != 0 {
			t.Errorf("day %s should have 0 chores, got %d", d.Day, len(d.Chores))
		}
	}
}

func TestBuildWeeklyView_OutOfRangeChoresIgnored(t *testing.T) {
	monday, _ := weekStart("2026-04-27")
	chores := []lib.Chore{
		{ID: "99", Title: "Old chore", Status: "completed", DueDate: "2026-04-20"},
	}
	days := buildWeeklyView(chores, monday)
	for _, d := range days {
		if len(d.Chores) != 0 {
			t.Errorf("day %s should have 0 chores, got %d", d.Day, len(d.Chores))
		}
	}
}
