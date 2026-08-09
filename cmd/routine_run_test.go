package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func routineMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/create_multiple") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":[{"id":"routine1","attributes":{"summary":"Morning","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"a1","type":"category"}}}}]}`)
		case strings.HasSuffix(r.URL.Path, "/chores/routine1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[{"id":"routine1","attributes":{"summary":"Morning","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"a1","type":"category"}}}}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestRoutineListCmd(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())

	out := captureStdout(func() {
		if err := routineListCmd.RunE(routineListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Morning") {
		t.Errorf("expected routine in output, got: %s", out)
	}
}

func TestRoutineCreateCmd(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())
	origTitle, origTimeOfDay, origCategoryIDs, origStartDate := routineTitle, routineTimeOfDay, routineCategoryIDs, routineStartDate
	routineTitle, routineTimeOfDay, routineCategoryIDs, routineStartDate = "Morning", "morning", []string{"a1"}, "2026-08-10"
	t.Cleanup(func() {
		routineTitle, routineTimeOfDay, routineCategoryIDs, routineStartDate = origTitle, origTimeOfDay, origCategoryIDs, origStartDate
	})

	out := captureStdout(func() {
		if err := routineCreateCmd.RunE(routineCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Morning") {
		t.Errorf("expected created routine in output, got: %s", out)
	}
}

func TestRoutineCreateCmd_InvalidTimeOfDay(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())
	origTitle, origTimeOfDay := routineTitle, routineTimeOfDay
	routineTitle, routineTimeOfDay = "Morning", "noon"
	t.Cleanup(func() { routineTitle, routineTimeOfDay = origTitle, origTimeOfDay })

	if err := routineCreateCmd.RunE(routineCreateCmd, nil); err == nil {
		t.Error("expected error for invalid --time-of-day, got nil")
	}
}

func TestRoutineDeleteCmd(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())
	origID, origYes := routineID, yes
	routineID, yes = "routine1", true
	t.Cleanup(func() { routineID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := routineDeleteCmd.RunE(routineDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestRoutineDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := routineID, dryRun
	routineID, dryRun = "routine1", true
	t.Cleanup(func() { routineID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := routineDeleteCmd.RunE(routineDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestRoutineDeleteCmd_ConfirmationDeclined(t *testing.T) {
	origID, origYes := routineID, yes
	routineID, yes = "routine1", false
	t.Cleanup(func() { routineID, yes = origID, origYes })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	mockStdin(t, "n\n")

	out := captureStdout(func() {
		if err := routineDeleteCmd.RunE(routineDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "deleted successfully") {
		t.Errorf("expected no deletion when confirmation declined, got: %s", out)
	}
}

func TestRoutineListCmd_TableMode(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/categories") {
			fmt.Fprint(w, `{"data":[{"id":"a1","attributes":{"label":"Bob","color":"#00FF00"}}]}`)
			return
		}
		fmt.Fprint(w, `{"data":[{"id":"routine1","attributes":{"summary":"Morning","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"a1","type":"category"}}}}]}`)
	})

	origFmt := outputFormat
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = origFmt })

	out := captureStdout(func() {
		if err := routineListCmd.RunE(routineListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Morning") {
		t.Errorf("expected routine title in table output, got: %s", out)
	}
	if !strings.Contains(out, "Bob") {
		t.Errorf("expected resolved category name in ASSIGNEE column, got: %s", out)
	}
}

func TestRoutineCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "routine")
}

func TestRoutineCmd_UpdateAndReorderRemoved(t *testing.T) {
	for _, use := range []string{"update", "reorder"} {
		for _, sub := range routineCmd.Commands() {
			if sub.Use == use || strings.HasPrefix(sub.Use, use+" ") {
				t.Errorf("expected %q subcommand to be removed from routine, but it's still registered", use)
			}
		}
	}
}
