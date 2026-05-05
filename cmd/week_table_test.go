package cmd

import (
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestPrintCalendarWeekTable_NoEvents(t *testing.T) {
	days := []WeeklyCalendarDay{
		{Day: "Mon", Date: "2026-04-27", Display: "Mon Apr 27"},
	}
	output := captureStdout(func() { printCalendarWeekTable(days) })
	if !strings.Contains(output, "(no events)") {
		t.Errorf("expected '(no events)' for empty day, got: %s", output)
	}
	if !strings.Contains(output, "Mon") {
		t.Errorf("expected day name in output, got: %s", output)
	}
}

func TestPrintCalendarWeekTable_WithEvents(t *testing.T) {
	days := []WeeklyCalendarDay{
		{
			Day:     "Tue",
			Date:    "2026-04-28",
			Display: "Tue Apr 28",
			Events: []lib.CalendarEvent{
				{Title: "Soccer practice", AllDay: false, StartAt: "2026-04-28T16:30:00Z"},
				{Title: "Birthday", AllDay: true},
			},
		},
	}
	output := captureStdout(func() { printCalendarWeekTable(days) })
	if !strings.Contains(output, "Soccer practice") {
		t.Errorf("expected event title in output, got: %s", output)
	}
	if !strings.Contains(output, "16:30") {
		t.Errorf("expected time in output, got: %s", output)
	}
	if !strings.Contains(output, "All day") {
		t.Errorf("expected 'All day' for all-day event, got: %s", output)
	}
	if !strings.Contains(output, "Birthday") {
		t.Errorf("expected birthday event in output, got: %s", output)
	}
}

func TestPrintCalendarWeekTable_MultiDaySecondRowBlanksColumns(t *testing.T) {
	days := []WeeklyCalendarDay{
		{
			Day:     "Wed",
			Date:    "2026-04-29",
			Display: "Wed Apr 29",
			Events: []lib.CalendarEvent{
				{Title: "Event A"},
				{Title: "Event B"},
			},
		},
	}
	output := captureStdout(func() { printCalendarWeekTable(days) })
	// Both events should appear
	if !strings.Contains(output, "Event A") || !strings.Contains(output, "Event B") {
		t.Errorf("expected both events in output, got: %s", output)
	}
}

func TestPrintChoreWeekTable_NoChores(t *testing.T) {
	days := []WeeklyChoreDay{
		{Day: "Mon", Date: "2026-04-27", Display: "Mon Apr 27"},
	}
	output := captureStdout(func() { printChoreWeekTable(days) })
	if !strings.Contains(output, "(no chores)") {
		t.Errorf("expected '(no chores)' for empty day, got: %s", output)
	}
}

func TestPrintChoreWeekTable_WithChores(t *testing.T) {
	days := []WeeklyChoreDay{
		{
			Day:     "Tue",
			Date:    "2026-04-28",
			Display: "Tue Apr 28",
			Chores: []lib.Chore{
				{Title: "Vacuum living room", Status: lib.ChoreStatusComplete},
				{Title: "Do dishes", Status: lib.ChoreStatusPending},
			},
		},
	}
	output := captureStdout(func() { printChoreWeekTable(days) })
	if !strings.Contains(output, "Vacuum living room") {
		t.Errorf("expected chore title in output, got: %s", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("expected checkmark for completed chore, got: %s", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("expected cross for pending chore, got: %s", output)
	}
}

func TestPrintChoreStreakTable(t *testing.T) {
	stats := []ChoreStreakStats{
		{AssigneeName: "Alice", CurrentStreak: 5, LongestStreak: 10, CompletedChores: 20, TotalChores: 25, CompletionRate: 80.0},
	}
	output := captureStdout(func() { printChoreStreakTable(stats) })
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected name in output, got: %s", output)
	}
	if !strings.Contains(output, "5 days") {
		t.Errorf("expected current streak in output, got: %s", output)
	}
	if !strings.Contains(output, "80.0%") {
		t.Errorf("expected completion rate in output, got: %s", output)
	}
	if !strings.Contains(output, "CURRENT STREAK") {
		t.Errorf("expected header in output, got: %s", output)
	}
}
