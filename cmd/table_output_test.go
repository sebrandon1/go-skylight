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
		{ID: "1", Title: "Morning", AssigneeID: "a1", Steps: []lib.RoutineStep{{}, {}}},
	}
	out := captureStdout(func() { printRoutinesTable(routines) })

	if !strings.Contains(out, "Morning") {
		t.Errorf("expected routine title in output, got: %s", out)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "STEPS") {
		t.Errorf("expected table headers in output, got: %s", out)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("expected step count in output, got: %s", out)
	}
}
