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
		case strings.HasSuffix(r.URL.Path, "/reorder"):
			fmt.Fprint(w, `{}`)
		case strings.HasSuffix(r.URL.Path, "/routine1") && r.Method == http.MethodPut:
			fmt.Fprint(w, `{"data":{"id":"routine1","attributes":{"title":"Updated","assignee_id":"a1","steps":[]}}}`)
		case strings.HasSuffix(r.URL.Path, "/routine1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/routines") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"routine1","attributes":{"title":"Morning","assignee_id":"a1","steps":[]}}}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"routine1","attributes":{"title":"Morning","assignee_id":"a1","steps":[]}}]}`)
		}
	}
}

func TestRoutineListCmd(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())

	out := captureStdout(func() { routineListCmd.Run(routineListCmd, nil) })
	if !strings.Contains(out, "Morning") {
		t.Errorf("expected routine in output, got: %s", out)
	}
}

func TestRoutineCreateCmd(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())
	origTitle, origSteps := routineTitle, routineSteps
	routineTitle, routineSteps = "Morning", []string{"Brush teeth", ""}
	t.Cleanup(func() { routineTitle, routineSteps = origTitle, origSteps })

	out := captureStdout(func() { routineCreateCmd.Run(routineCreateCmd, nil) })
	if !strings.Contains(out, "Morning") {
		t.Errorf("expected created routine in output, got: %s", out)
	}
}

func TestRoutineUpdateCmd(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())
	origID, origTitle := routineID, routineTitle
	routineID, routineTitle = "routine1", "Updated"
	t.Cleanup(func() { routineID, routineTitle = origID, origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := routineUpdateCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() { routineUpdateCmd.Run(routineUpdateCmd, nil) })
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated routine in output, got: %s", out)
	}
}

func TestRoutineDeleteCmd(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())
	origID := routineID
	routineID = "routine1"
	t.Cleanup(func() { routineID = origID })

	out := captureStdout(func() { routineDeleteCmd.Run(routineDeleteCmd, nil) })
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestRoutineReorderCmd(t *testing.T) {
	newCmdTestClient(t, routineMockHandler())
	origIDs := routineIDs
	routineIDs = []string{"routine1", "routine2"}
	t.Cleanup(func() { routineIDs = origIDs })

	out := captureStdout(func() { routineReorderCmd.Run(routineReorderCmd, nil) })
	if !strings.Contains(out, "reordered successfully") {
		t.Errorf("expected reorder confirmation, got: %s", out)
	}
}

func TestRoutineCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "routine")
}
