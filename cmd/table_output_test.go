package cmd

import (
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

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
