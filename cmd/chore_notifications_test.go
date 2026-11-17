package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/sebrandon1/go-skylight/lib"
)

const notifTestJSON = `{"data":{"task_due":{"enabled":true,"filters":{"category_ids":["a"],"up_for_grabs":false}},"task_completed":{"enabled":false,"filters":{"category_ids":[],"up_for_grabs":true}}}}`

// setFlag sets a flag on cmd and marks it unchanged again on cleanup, since
// pflag has no unset API and the commands are shared singletons.
func setFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("setting --%s: %v", name, err)
	}
	t.Cleanup(func() {
		f := cmd.Flags().Lookup(name)
		f.Changed = false
		if sv, ok := f.Value.(interface{ Replace([]string) error }); ok {
			_ = sv.Replace(nil)
		} else {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		}
	})
}

func notifHandler(patched *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			b, _ := io.ReadAll(r.Body)
			*patched = string(b)
		}
		fmt.Fprint(w, notifTestJSON)
	}
}

func TestChoreNotificationsShowCmd(t *testing.T) {
	var patched string
	newCmdTestClient(t, notifHandler(&patched))

	out := captureStdout(func() {
		if err := choreNotificationsShowCmd.RunE(choreNotificationsShowCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, `"task_due"`) || !strings.Contains(out, `"up_for_grabs"`) {
		t.Errorf("expected settings JSON, got: %s", out)
	}
	if patched != "" {
		t.Error("show must not PATCH")
	}
}

func TestChoreNotificationsShowCmd_Table(t *testing.T) {
	var patched string
	newCmdTestClient(t, notifHandler(&patched))
	orig := outputFormat
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = orig })

	out := captureStdout(func() {
		if err := choreNotificationsShowCmd.RunE(choreNotificationsShowCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	for _, want := range []string{"NOTIFICATION", "task due", "task completed", "yes", "no"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q: %s", want, out)
		}
	}
}

func TestChoreNotificationsSetCmd(t *testing.T) {
	var patched string
	newCmdTestClient(t, notifHandler(&patched))
	setFlag(t, choreNotificationsSetCmd, flagTaskDue, "off")
	setFlag(t, choreNotificationsSetCmd, flagTaskCompleted, "on")
	setFlag(t, choreNotificationsSetCmd, flagCompletedUpForGrabs, "off")
	setFlag(t, choreNotificationsSetCmd, flagCompletedCategoryID, "x,y")

	captureStdout(func() {
		if err := choreNotificationsSetCmd.RunE(choreNotificationsSetCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	var body map[string]struct {
		Enabled bool `json:"enabled"`
		Filters struct {
			CategoryIDs []string `json:"category_ids"`
			UpForGrabs  bool     `json:"up_for_grabs"`
		} `json:"filters"`
	}
	if err := json.Unmarshal([]byte(patched), &body); err != nil {
		t.Fatalf("PATCH body %q: %v", patched, err)
	}
	due, completed := body["task_due"], body["task_completed"]
	if due.Enabled || due.Filters.CategoryIDs != nil {
		t.Errorf("task_due should be disabled and sent without filters, got %+v", due)
	}
	if !completed.Enabled || completed.Filters.UpForGrabs || len(completed.Filters.CategoryIDs) != 2 {
		t.Errorf("task_completed wrong: %+v", completed)
	}
}

func TestChoreNotificationsSetCmd_OnlyChangedRuleSent(t *testing.T) {
	var patched string
	newCmdTestClient(t, notifHandler(&patched))
	setFlag(t, choreNotificationsSetCmd, flagTaskCompleted, "on")

	captureStdout(func() {
		if err := choreNotificationsSetCmd.RunE(choreNotificationsSetCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(patched, "task_due") || !strings.Contains(patched, "task_completed") {
		t.Errorf("only task_completed should be sent, got %s", patched)
	}
}

func TestChoreNotificationsSetCmd_ClearCategories(t *testing.T) {
	var patched string
	newCmdTestClient(t, notifHandler(&patched))
	setFlag(t, choreNotificationsSetCmd, flagDueCategoryID, "")
	setFlag(t, choreNotificationsSetCmd, flagDueUpForGrabs, "on")

	captureStdout(func() {
		if err := choreNotificationsSetCmd.RunE(choreNotificationsSetCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(patched, `"category_ids":[]`) {
		t.Errorf("expected cleared category_ids, got %s", patched)
	}
}

func TestChoreNotificationsSetCmd_Errors(t *testing.T) {
	var patched string
	newCmdTestClient(t, notifHandler(&patched))

	if err := choreNotificationsSetCmd.RunE(choreNotificationsSetCmd, nil); err == nil {
		t.Error("expected error when no flags are given")
	}

	setFlag(t, choreNotificationsSetCmd, flagTaskDue, "maybe")
	if err := choreNotificationsSetCmd.RunE(choreNotificationsSetCmd, nil); err == nil {
		t.Error("expected error for invalid on/off value")
	}
	if patched != "" {
		t.Error("no PATCH expected on validation errors")
	}
}

func TestChoreNotificationsSetCmd_DryRun(t *testing.T) {
	var patched string
	newCmdTestClient(t, notifHandler(&patched))
	setFlag(t, choreNotificationsSetCmd, flagTaskDue, "on")
	orig := dryRun
	dryRun = true
	t.Cleanup(func() { dryRun = orig })

	captureStdout(func() {
		if err := choreNotificationsSetCmd.RunE(choreNotificationsSetCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if patched != "" {
		t.Error("dry run must not PATCH")
	}
}

func TestChoreCreateCmd_DueTime(t *testing.T) {
	var sent string
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sent = string(b)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Dishes","start_time":"07:30"}}}`)
	})
	origTitle, origUp, origTime := choreTitle, choreUpForGrabs, choreDueTime
	choreTitle, choreUpForGrabs, choreDueTime = "Dishes", false, "07:30"
	t.Cleanup(func() { choreTitle, choreUpForGrabs, choreDueTime = origTitle, origUp, origTime })

	captureStdout(func() {
		if err := choreCreateCmd.RunE(choreCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(sent, `"start_time":"07:30"`) {
		t.Errorf("expected start_time in body, got %s", sent)
	}

	choreDueTime = "25:00"
	sent = ""
	if err := choreCreateCmd.RunE(choreCreateCmd, nil); err == nil {
		t.Error("expected error for invalid --due-time")
	}
	if sent != "" {
		t.Error("invalid due time must not reach the API")
	}
}

func TestChoreUpdateCmd_DueTime(t *testing.T) {
	var sent string
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sent = string(b)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Dishes","start_time":"19:00"}}}`)
	})
	origID, origTime := choreID, choreDueTime
	choreID, choreDueTime = "c1", "bad"
	t.Cleanup(func() { choreID, choreDueTime = origID, origTime })
	setFlag(t, choreUpdateCmd, flagDueTime, "bad")

	if err := choreUpdateCmd.RunE(choreUpdateCmd, nil); err == nil {
		t.Error("expected error for invalid --due-time")
	}
	if sent != "" {
		t.Error("invalid due time must not reach the API")
	}

	choreDueTime = "19:00"
	captureStdout(func() {
		if err := choreUpdateCmd.RunE(choreUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(sent, `"start_time":"19:00"`) {
		t.Errorf("expected start_time in body, got %s", sent)
	}
}

func TestChoreNotificationsSetCmd_FiltersOnDisabledRule(t *testing.T) {
	var patched string
	newCmdTestClient(t, notifHandler(&patched))
	setFlag(t, choreNotificationsSetCmd, flagCompletedCategoryID, "x")

	err := choreNotificationsSetCmd.RunE(choreNotificationsSetCmd, nil)
	if !errors.Is(err, lib.ErrFiltersOnDisabledRule) {
		t.Errorf("want ErrFiltersOnDisabledRule, got %v", err)
	}
	if patched != "" {
		t.Errorf("no PATCH should be sent, got %s", patched)
	}
}
