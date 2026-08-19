package cmd

import (
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestPrintCalendarTable_AllDay(t *testing.T) {
	events := []lib.CalendarEvent{
		{ID: "e1", Title: "Holiday", StartAt: "2026-07-14", AllDay: true},
	}
	out := captureStdout(func() { printCalendarTable(events) })

	if !strings.Contains(out, "Holiday") {
		t.Errorf("expected event title in output, got: %s", out)
	}
	if !strings.Contains(out, boolYes) {
		t.Errorf("expected %q for all-day event, got: %s", boolYes, out)
	}
}

func TestPrintRoutinesTable(t *testing.T) {
	routines := []lib.Routine{
		{ID: "1", Title: "Morning", TimeOfDay: "morning", AssigneeID: "a1"},
	}
	out := captureStdout(func() { printRoutinesTable(routines) })

	if !strings.Contains(out, "Morning") {
		t.Errorf("expected routine title in output, got: %s", out)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "TIME OF DAY") {
		t.Errorf("expected table headers in output, got: %s", out)
	}
	if !strings.Contains(out, "morning") {
		t.Errorf("expected time-of-day in output, got: %s", out)
	}
}

func TestPrintChoresTable_NonRecurringShowsDashes(t *testing.T) {
	chores := []lib.Chore{{ID: "c1", Title: "Dishes", Recurring: false}}
	out := captureStdout(func() { printChoresTable(chores) })
	if !strings.Contains(out, "FREQUENCY") {
		t.Errorf("expected FREQUENCY column header, got: %s", out)
	}
	// Non-recurring chores show "-" for all four recurrence fields.
	if strings.Count(out, "-") < 4 {
		t.Errorf("expected at least 4 dashes for non-recurring recurrence fields, got: %s", out)
	}
}

func TestPrintChoresTable_RecurringShowsFields(t *testing.T) {
	chores := []lib.Chore{{
		ID:             "c1",
		Title:          "Exercise",
		Recurring:      true,
		Frequency:      "weekly",
		Interval:       2,
		RecurrenceDays: []string{"mon", "wed"},
		EndDate:        "2026-12-31",
	}}
	out := captureStdout(func() { printChoresTable(chores) })
	if !strings.Contains(out, "weekly") {
		t.Errorf("expected frequency in output, got: %s", out)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("expected interval in output, got: %s", out)
	}
	if !strings.Contains(out, "mon,wed") {
		t.Errorf("expected recurrence days in output, got: %s", out)
	}
	if !strings.Contains(out, "2026-12-31") {
		t.Errorf("expected end date in output, got: %s", out)
	}
}

func TestPrintChoresTable_ResolvesCatName(t *testing.T) {
	orig := activeCatNames
	activeCatNames = map[string]string{"cat1": "Alice"}
	t.Cleanup(func() { activeCatNames = orig })

	chores := []lib.Chore{{ID: "c1", Title: "Dishes", AssigneeID: "cat1"}}
	out := captureStdout(func() { printChoresTable(chores) })
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected resolved name in ASSIGNEE column, got: %s", out)
	}
	if strings.Contains(out, "cat1") {
		t.Errorf("expected raw ID to be replaced by name, got: %s", out)
	}
}

func TestPrintRewardsTable_ResolvesCatName(t *testing.T) {
	orig := activeCatNames
	activeCatNames = map[string]string{"cat2": "Bob"}
	t.Cleanup(func() { activeCatNames = orig })

	rewards := []lib.Reward{{ID: "r1", Title: "Candy", CategoryID: "cat2"}}
	out := captureStdout(func() { printRewardsTable(rewards) })
	if !strings.Contains(out, "Bob") {
		t.Errorf("expected resolved name in CATEGORY column, got: %s", out)
	}
	if strings.Contains(out, "cat2") {
		t.Errorf("expected raw ID to be replaced by name, got: %s", out)
	}
}

func TestPrintRoutinesTable_ResolvesCatName(t *testing.T) {
	orig := activeCatNames
	activeCatNames = map[string]string{"a1": "Charlie"}
	t.Cleanup(func() { activeCatNames = orig })

	routines := []lib.Routine{{ID: "r1", Title: "Morning", AssigneeID: "a1"}}
	out := captureStdout(func() { printRoutinesTable(routines) })
	if !strings.Contains(out, "Charlie") {
		t.Errorf("expected resolved name in ASSIGNEE column, got: %s", out)
	}
	if strings.Contains(out, "a1") {
		t.Errorf("expected raw ID to be replaced by name, got: %s", out)
	}
}
