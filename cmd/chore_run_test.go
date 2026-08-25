package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

	out := captureStdout(func() {
		if err := choreListCmd.RunE(choreListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
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

	out := captureStdout(func() {
		if err := choreListCmd.RunE(choreListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if out == "" {
		t.Error("expected non-empty weekly view output")
	}
}

func TestChoreCreateCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origTitle, origUpForGrabs := choreTitle, choreUpForGrabs
	choreTitle, choreUpForGrabs = "Dishes", false
	t.Cleanup(func() { choreTitle, choreUpForGrabs = origTitle, origUpForGrabs })

	out := captureStdout(func() {
		if err := choreCreateCmd.RunE(choreCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected created chore in output, got: %s", out)
	}
}

func TestChoreCreateCmd_UpForGrabs(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origTitle, origUpForGrabs := choreTitle, choreUpForGrabs
	choreTitle, choreUpForGrabs = "Dishes", true
	t.Cleanup(func() { choreTitle, choreUpForGrabs = origTitle, origUpForGrabs })

	out := captureStdout(func() {
		if err := choreCreateCmd.RunE(choreCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected created up-for-grabs chore in output, got: %s", out)
	}
}

func TestChoreDeleteCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID, origYes := choreID, yes
	choreID, yes = "c1", true
	t.Cleanup(func() { choreID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := choreDeleteCmd.RunE(choreDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestChoreDeleteCmd_Quiet(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID, origYes, origQuiet := choreID, yes, quiet
	choreID, yes, quiet = "c1", true, true
	t.Cleanup(func() { choreID, yes, quiet = origID, origYes, origQuiet })

	out := captureStdout(func() {
		if err := choreDeleteCmd.RunE(choreDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
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

	out := captureStdout(func() {
		if err := choreDeleteCmd.RunE(choreDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestChoreCompleteCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID := choreID
	choreID = "c1"
	t.Cleanup(func() { choreID = origID })

	out := captureStdout(func() {
		if err := choreCompleteCmd.RunE(choreCompleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
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

	out := captureStdout(func() {
		if err := choreUpdateCmd.RunE(choreUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated chore in output, got: %s", out)
	}
}

func TestChoreSkipCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID := choreID
	choreID = "c1"
	t.Cleanup(func() { choreID = origID })

	out := captureStdout(func() {
		if err := choreSkipCmd.RunE(choreSkipCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "skipped successfully") {
		t.Errorf("expected skip confirmation, got: %s", out)
	}
}

func TestChoreSkipCmd_DeferUntil(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID, origDefer := choreID, choreSkipDeferUntil
	choreID = "c1"
	choreSkipDeferUntil = "2026-08-10"
	t.Cleanup(func() { choreID, choreSkipDeferUntil = origID, origDefer })

	out := captureStdout(func() {
		if err := choreSkipCmd.RunE(choreSkipCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "skipped successfully") {
		t.Errorf("expected skip confirmation with defer-until, got: %s", out)
	}
}

func TestChoreClaimCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID, origAssignee := choreID, choreAssigneeID
	choreID, choreAssigneeID = "c1", "a1"
	t.Cleanup(func() { choreID, choreAssigneeID = origID, origAssignee })

	out := captureStdout(func() {
		if err := choreClaimCmd.RunE(choreClaimCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, `"id": "c1"`) {
		t.Errorf("expected claimed chore in output, got: %s", out)
	}
}

func TestChoreCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "chore")
}

func TestChoreCreateCmd_Recurring(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())

	origTitle, origRecurring, origUpForGrabs := choreTitle, choreRecurring, choreUpForGrabs
	choreTitle, choreRecurring, choreUpForGrabs = "Daily walk", true, false
	t.Cleanup(func() {
		choreTitle, choreRecurring, choreUpForGrabs = origTitle, origRecurring, origUpForGrabs
	})

	out := captureStdout(func() {
		if err := choreCreateCmd.RunE(choreCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, `"id"`) {
		t.Errorf("expected chore JSON in output, got: %s", out)
	}
}

func TestChoreUpdateCmd_UpForGrabs(t *testing.T) {
	var gotUpForGrabs bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/c1") && r.Method == http.MethodPut {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				gotUpForGrabs, _ = body["up_for_grabs"].(bool)
			}
			fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Dishes","up_for_grabs":true}}}`)
			return
		}
		fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"pending"}}]}`)
	})
	newCmdTestClient(t, handler)
	origID, origUpForGrabs := choreID, choreUpForGrabs
	choreID, choreUpForGrabs = "c1", true
	t.Cleanup(func() { choreID, choreUpForGrabs = origID, origUpForGrabs })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := choreUpdateCmd.Flags().Set("up-for-grabs", "true"); err != nil {
		t.Fatalf("setting up-for-grabs flag: %v", err)
	}

	captureStdout(func() {
		if err := choreUpdateCmd.RunE(choreUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !gotUpForGrabs {
		t.Error("expected up_for_grabs=true in PATCH request body")
	}
}

func TestChoreSearchCmd(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origQuery := choreSearchQuery
	choreSearchQuery = "dishes"
	t.Cleanup(func() { choreSearchQuery = origQuery })

	out := captureStdout(func() {
		if err := choreSearchCmd.RunE(choreSearchCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected search results in output, got: %s", out)
	}
}

func TestChoreCreateCmd_TableOutput(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origTitle, origUpForGrabs, origFmt := choreTitle, choreUpForGrabs, outputFormat
	choreTitle, choreUpForGrabs, outputFormat = "Dishes", false, outputTable
	t.Cleanup(func() { choreTitle, choreUpForGrabs, outputFormat = origTitle, origUpForGrabs, origFmt })

	out := captureStdout(func() {
		if err := choreCreateCmd.RunE(choreCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "TITLE") {
		t.Errorf("expected table header in output, got: %s", out)
	}
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected chore title in table, got: %s", out)
	}
}

func TestChoreSearchCmd_DateWindow(t *testing.T) {
	newCmdTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("q") != "trash" {
			t.Errorf("q param: want %q got %q", "trash", q.Get("q"))
		}
		if q.Get("after") == "" || q.Get("before") == "" {
			t.Errorf("expected default date window on search, got after=%q before=%q", q.Get("after"), q.Get("before"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"c9","attributes":{"summary":"Take out trash","status":"pending"}}]}`)
	}))

	orig := choreSearchQuery
	choreSearchQuery = "trash"
	t.Cleanup(func() { choreSearchQuery = orig })

	out := captureStdout(func() {
		if err := choreSearchCmd.RunE(choreSearchCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Take out trash") {
		t.Errorf("expected matching chore in output, got: %s", out)
	}
}

// Verifies the request params rather than the rendered output: the shared
// choreListCmd singleton may have flags marked changed by earlier tests
// (e.g. --week), which changes what this command prints.
func TestChoreListCmd_SendsDefaultDateWindow(t *testing.T) {
	newCmdTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("after") == "" || q.Get("before") == "" {
			t.Errorf("expected default date window on bare list, got after=%q before=%q", q.Get("after"), q.Get("before"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[]}`)
	}))

	captureStdout(func() {
		if err := choreListCmd.RunE(choreListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRequireWindowPair(t *testing.T) {
	tests := []struct {
		name    string
		after   bool
		before  bool
		wantErr string
	}{
		{name: "no flags passes"},
		{name: "both flags pass", after: true, before: true},
		{name: "lone after fails", after: true, wantErr: "--after requires --before"},
		{name: "lone before fails", before: true, wantErr: "--before requires --after"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("after", "", "")
			cmd.Flags().String("before", "", "")
			if tc.after {
				if err := cmd.Flags().Set("after", "2026-01-01"); err != nil {
					t.Fatalf("setting after: %v", err)
				}
			}
			if tc.before {
				if err := cmd.Flags().Set("before", "2026-01-31"); err != nil {
					t.Fatalf("setting before: %v", err)
				}
			}
			err := requireWindowPair(cmd)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
