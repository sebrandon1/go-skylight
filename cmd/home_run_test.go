package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func homeMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"pending"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[{"id":"l1","type":"list","attributes":{"label":"Groceries"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	}
}

func TestHomeCmd_Text(t *testing.T) {
	newCmdTestClient(t, homeMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	origNoTasks, origNoLists := homeNoTasks, homeNoLists
	homeNoTasks, homeNoLists = false, false
	t.Cleanup(func() { homeNoTasks, homeNoLists = origNoTasks, origNoLists })

	out := captureStdout(func() { homeCmd.Run(homeCmd, nil) })
	if !strings.Contains(out, "Dishes") || !strings.Contains(out, "Groceries") {
		t.Errorf("expected chore and list in output, got: %s", out)
	}
}

func TestHomeCmd_JSON(t *testing.T) {
	newCmdTestClient(t, homeMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	origNoTasks, origNoLists := homeNoTasks, homeNoLists
	homeNoTasks, homeNoLists = false, false
	t.Cleanup(func() { homeNoTasks, homeNoLists = origNoTasks, origNoLists })

	out := captureStdout(func() { homeCmd.Run(homeCmd, nil) })
	if !strings.Contains(out, `"week_start"`) {
		t.Errorf("expected week_start in JSON output, got: %s", out)
	}
}

func TestHomeCmd_NoTasksNoLists(t *testing.T) {
	newCmdTestClient(t, homeMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	origNoTasks, origNoLists := homeNoTasks, homeNoLists
	homeNoTasks, homeNoLists = true, true
	t.Cleanup(func() { homeNoTasks, homeNoLists = origNoTasks, origNoLists })

	out := captureStdout(func() { homeCmd.Run(homeCmd, nil) })
	if strings.Contains(out, "PENDING TASKS") || strings.Contains(out, "=== LISTS ===") {
		t.Errorf("expected tasks/lists sections to be skipped, got: %s", out)
	}
}

func TestHomeCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "home")
}
