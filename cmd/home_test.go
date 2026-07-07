package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestPrintHomeTable(t *testing.T) {
	monday, _ := weekStart("")

	t.Run("shows tasks, lists, and meals when present", func(t *testing.T) {
		chores := []lib.Chore{{ID: "c1", Title: "Dishes", Status: lib.ChoreStatusPending}}
		lists := []lib.List{{ID: "l1", Title: "Groceries"}}
		meals := []lib.MealSitting{{ID: "s1", Summary: "Dinner"}}

		out := captureStdout(func() { printHomeTable(nil, chores, lists, meals, monday) })

		if !strings.Contains(out, "EVENTS THIS WEEK") {
			t.Errorf("expected events section header, got: %s", out)
		}
		if !strings.Contains(out, "PENDING TASKS TODAY") || !strings.Contains(out, "Dishes") {
			t.Errorf("expected pending tasks section with chore, got: %s", out)
		}
		if !strings.Contains(out, "LISTS") || !strings.Contains(out, "Groceries") {
			t.Errorf("expected lists section with list, got: %s", out)
		}
		if !strings.Contains(out, "MEALS THIS WEEK") || !strings.Contains(out, "Dinner") {
			t.Errorf("expected meals section with sitting, got: %s", out)
		}
	})

	t.Run("omits tasks, lists, and meals sections when empty", func(t *testing.T) {
		out := captureStdout(func() { printHomeTable(nil, nil, nil, nil, monday) })

		if strings.Contains(out, "PENDING TASKS TODAY") {
			t.Errorf("expected no tasks section when chores empty, got: %s", out)
		}
		if strings.Contains(out, "=== LISTS ===") {
			t.Errorf("expected no lists section when lists empty, got: %s", out)
		}
		if strings.Contains(out, "=== MEALS THIS WEEK ===") {
			t.Errorf("expected no meals section when meals empty, got: %s", out)
		}
	})

	t.Run("includes events in the weekly view", func(t *testing.T) {
		events := []lib.CalendarEvent{
			{ID: "e1", Title: "Standup", StartAt: monday.Format(time.RFC3339)},
		}
		out := captureStdout(func() { printHomeTable(events, nil, nil, nil, monday) })
		if !strings.Contains(out, "Standup") {
			t.Errorf("expected event title in weekly view, got: %s", out)
		}
	})
}
