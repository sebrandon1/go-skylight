package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func choreMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/completions"):
			fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Dishes"}}}`)
		case strings.HasSuffix(r.URL.Path, "/create_multiple"):
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","up_for_grabs":true}}]}`)
		case strings.HasSuffix(r.URL.Path, "/c1") && r.Method == http.MethodPut:
			fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Updated"}}}`)
		case strings.HasSuffix(r.URL.Path, "/c1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/chores") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Dishes"}}}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"pending"}}]}`)
		}
	}
}

func TestChoreListCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())

	out := captureStdout(func() { choreListCmd.Run(choreListCmd, nil) })
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected chore in output, got: %s", out)
	}
}

func TestChoreListCmd_WeekView(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := choreListCmd.Flags().Set("week", "current"); err != nil {
		t.Fatalf("setting week flag: %v", err)
	}

	out := captureStdout(func() { choreListCmd.Run(choreListCmd, nil) })
	if out == "" {
		t.Error("expected non-empty weekly view output")
	}
}

func TestChoreCreateCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origTitle, origUpForGrabs := choreTitle, choreUpForGrabs
	choreTitle, choreUpForGrabs = "Dishes", false
	t.Cleanup(func() { choreTitle, choreUpForGrabs = origTitle, origUpForGrabs })

	out := captureStdout(func() { choreCreateCmd.Run(choreCreateCmd, nil) })
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected created chore in output, got: %s", out)
	}
}

func TestChoreCreateCmd_UpForGrabs(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origTitle, origUpForGrabs := choreTitle, choreUpForGrabs
	choreTitle, choreUpForGrabs = "Dishes", true
	t.Cleanup(func() { choreTitle, choreUpForGrabs = origTitle, origUpForGrabs })

	out := captureStdout(func() { choreCreateCmd.Run(choreCreateCmd, nil) })
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected created up-for-grabs chore in output, got: %s", out)
	}
}

func TestChoreDeleteCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID := choreID
	choreID = "c1"
	t.Cleanup(func() { choreID = origID })

	out := captureStdout(func() { choreDeleteCmd.Run(choreDeleteCmd, nil) })
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestChoreDeleteCmd_Quiet(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID := choreID
	choreID = "c1"
	t.Cleanup(func() { choreID = origID })

	origQuiet := quiet
	quiet = true
	t.Cleanup(func() { quiet = origQuiet })

	out := captureStdout(func() { choreDeleteCmd.Run(choreDeleteCmd, nil) })
	if out != "" {
		t.Errorf("expected no output with --quiet, got: %s", out)
	}
}

func TestChoreDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := choreID, dryRun
	choreID, dryRun = "c1", true
	t.Cleanup(func() { choreID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() { choreDeleteCmd.Run(choreDeleteCmd, nil) })
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestChoreCompleteCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID := choreID
	choreID = "c1"
	t.Cleanup(func() { choreID = origID })

	out := captureStdout(func() { choreCompleteCmd.Run(choreCompleteCmd, nil) })
	if !strings.Contains(out, "completed successfully") {
		t.Errorf("expected completion confirmation, got: %s", out)
	}
}

func TestChoreUpdateCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID, origTitle := choreID, choreTitle
	choreID, choreTitle = "c1", "Updated"
	t.Cleanup(func() { choreID, choreTitle = origID, origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := choreUpdateCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() { choreUpdateCmd.Run(choreUpdateCmd, nil) })
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated chore in output, got: %s", out)
	}
}

func TestChoreSkipCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID := choreID
	choreID = "c1"
	t.Cleanup(func() { choreID = origID })

	out := captureStdout(func() { choreSkipCmd.Run(choreSkipCmd, nil) })
	if !strings.Contains(out, "skipped successfully") {
		t.Errorf("expected skip confirmation, got: %s", out)
	}
}

func TestChoreClaimCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID, origAssignee := choreID, choreAssigneeID
	choreID, choreAssigneeID = "c1", "a1"
	t.Cleanup(func() { choreID, choreAssigneeID = origID, origAssignee })

	out := captureStdout(func() { choreClaimCmd.Run(choreClaimCmd, nil) })
	if !strings.Contains(out, `"id": "c1"`) {
		t.Errorf("expected claimed chore in output, got: %s", out)
	}
}

func TestChoreCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "chore")
}
