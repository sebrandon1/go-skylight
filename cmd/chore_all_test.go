package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

const allTasksBody = `{"chores":{"late":[{"id":"1","attributes":{"summary":"Late dishes","reward_points":2}}],
"future":[{"id":"2","attributes":{"summary":"Grab me","up_for_grabs":true,"start":"2026-12-01"}}]},
"routines":{"today_timed":[{"id":"3","attributes":{"summary":"Morning routine","routine":true,"start_time":"06:00"}}]}}`

func allTasksHandler(gotQuery *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/chores/all") {
			*gotQuery = r.URL.Path + "?" + r.URL.RawQuery
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, allTasksBody)
	}
}

func TestChoreAllCmd(t *testing.T) {
	var got string
	newCmdTestClient(t, allTasksHandler(&got))
	origLimit, origGrabs := choreAllLimit, choreAllUpForGrabs
	choreAllLimit, choreAllUpForGrabs = 10, true
	t.Cleanup(func() { choreAllLimit, choreAllUpForGrabs = origLimit, origGrabs })

	out := captureStdout(func() {
		if err := choreAllCmd.RunE(choreAllCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(got, "/chores/all?include_up_for_grabs=true&limit=10") {
		t.Errorf("request = %q", got)
	}
	for _, want := range []string{"Late dishes", "Grab me", "Morning routine", `"today_timed"`} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got: %s", want, out)
		}
	}
}

func TestTaskAllCmd_Table(t *testing.T) {
	var got string
	newCmdTestClient(t, allTasksHandler(&got))
	outputFormat = outputTable
	origLimit, origGrabs := choreAllLimit, choreAllUpForGrabs
	choreAllLimit, choreAllUpForGrabs = 0, false
	t.Cleanup(func() { outputFormat = ""; choreAllLimit, choreAllUpForGrabs = origLimit, origGrabs })

	out := captureStdout(func() {
		if err := taskAllCmd.RunE(taskAllCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(got, "limit=") || !strings.Contains(got, "include_up_for_grabs=false") {
		t.Errorf("request = %q", got)
	}
	for _, want := range []string{"KIND", "chore", "late", "Late dishes", "up for grabs", "routine", "today_timed", "06:00"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in table, got: %s", want, out)
		}
	}
}

func TestChoreAllCmd_NegativeLimit(t *testing.T) {
	var got string
	newCmdTestClient(t, allTasksHandler(&got))
	orig := choreAllLimit
	choreAllLimit = -1
	t.Cleanup(func() { choreAllLimit = orig })
	if err := choreAllCmd.RunE(choreAllCmd, nil); err == nil {
		t.Error("expected error for negative limit")
	}
}

func TestChoreAllCmd_APIError(t *testing.T) {
	newCmdTestClient(t, http500Handler)
	if err := choreAllCmd.RunE(choreAllCmd, nil); err == nil {
		t.Error("expected error on API failure")
	}
}
