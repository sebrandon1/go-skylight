package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

// resetFlags clears the sticky Changed state pflag keeps on shared command singletons.
func resetFlags(t *testing.T, c *cobra.Command, names ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, n := range names {
			c.Flags().Lookup(n).Changed = false
		}
		choreAssigneeIDs, choreApplyToProfiles, choreDate, choreAssigneeID = nil, "", "", ""
	})
}

func TestChoreCreateCmd_LinkedAssigneeIDs(t *testing.T) {
	var body map[string]any
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/create_multiple") {
			_ = json.NewDecoder(r.Body).Decode(&body)
			fmt.Fprint(w, `{"data":[{"id":"1","attributes":{"summary":"Bins"},"relationships":{"linked_task_group":{"data":{"id":"g9","type":"linked_task_group"}}}},{"id":"2","attributes":{"summary":"Bins"}}]}`)
			return
		}
		fmt.Fprint(w, `{"data":[]}`)
	})
	origTitle := choreTitle
	choreTitle = "Bins"
	t.Cleanup(func() { choreTitle = origTitle })
	resetFlags(t, choreCreateCmd, "assignee-ids")
	setFlag(t, choreCreateCmd, "assignee-ids", "cat1,cat2")

	out := captureStdout(func() {
		if err := choreCreateCmd.RunE(choreCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !reflect.DeepEqual(body["category_ids"], []any{"cat1", "cat2"}) {
		t.Errorf("category_ids: %v", body["category_ids"])
	}
	if !strings.Contains(out, "g9") || strings.Count(out, "Bins") < 2 {
		t.Errorf("expected both linked chores and group id in output, got: %s", out)
	}
}

func TestChoreCreateCmd_LinkedValidation(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origTitle := choreTitle
	choreTitle = "Bins"
	t.Cleanup(func() { choreTitle = origTitle })
	resetFlags(t, choreCreateCmd, "assignee-ids", "assignee-id")

	setFlag(t, choreCreateCmd, "assignee-ids", "only-one")
	if err := choreCreateCmd.RunE(choreCreateCmd, nil); err == nil || !strings.Contains(err.Error(), "at least two") {
		t.Errorf("expected at-least-two error, got %v", err)
	}

	setFlag(t, choreCreateCmd, "assignee-ids", "a,b")
	setFlag(t, choreCreateCmd, "assignee-id", "a")
	if err := choreCreateCmd.RunE(choreCreateCmd, nil); err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutual exclusion error, got %v", err)
	}
}

func TestChoreUpdateCmd_ApplyToProfiles(t *testing.T) {
	var body map[string]any
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Bins"}}}`)
	})
	origID := choreID
	choreID = "c1"
	t.Cleanup(func() { choreID = origID })
	resetFlags(t, choreUpdateCmd, "assignee-ids", "apply-to-profiles", "date", "assignee-id")

	// --apply-to-profiles without --date is rejected.
	setFlag(t, choreUpdateCmd, "apply-to-profiles", "all")
	if err := choreUpdateCmd.RunE(choreUpdateCmd, nil); err == nil || !strings.Contains(err.Error(), "--date") {
		t.Errorf("expected --date error, got %v", err)
	}

	setFlag(t, choreUpdateCmd, "date", "2026-11-20")
	setFlag(t, choreUpdateCmd, "assignee-ids", "a,b")
	captureStdout(func() {
		if err := choreUpdateCmd.RunE(choreUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if body["apply_to_profiles"] != "all" || body["after"] != "2026-11-20" || body["before"] != "2026-11-20" {
		t.Errorf("unexpected body: %v", body)
	}
	if !reflect.DeepEqual(body["category_ids"], []any{"a", "b"}) {
		t.Errorf("category_ids: %v", body["category_ids"])
	}

	// --assignee-ids needs scope "all".
	setFlag(t, choreUpdateCmd, "apply-to-profiles", "one")
	if err := choreUpdateCmd.RunE(choreUpdateCmd, nil); err == nil || !strings.Contains(err.Error(), "apply-to-profiles all") {
		t.Errorf("expected scope error, got %v", err)
	}
}

func TestChoreDeleteCmd_ApplyToProfiles(t *testing.T) {
	var got string
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			got = r.URL.Query().Get("apply_to_profiles")
			if r.URL.Query().Get("apply_to") != "" {
				t.Errorf("apply_to must not be sent alongside apply_to_profiles")
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
	origID, origYes := choreID, yes
	choreID, yes = "c1", true
	t.Cleanup(func() { choreID, yes = origID, origYes })
	resetFlags(t, choreDeleteCmd, "apply-to-profiles")
	setFlag(t, choreDeleteCmd, "apply-to-profiles", "one")

	captureStdout(func() {
		if err := choreDeleteCmd.RunE(choreDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if got != "one" {
		t.Errorf("apply_to_profiles: want one got %q", got)
	}

	choreApplyToProfiles = "bogus"
	if err := choreDeleteCmd.RunE(choreDeleteCmd, nil); err == nil {
		t.Error("expected invalid scope error")
	}
}

func TestChoreDeleteCmd_ApplyToProfilesExcludesRecurring(t *testing.T) {
	newCmdTestClient(t, choreMockHandler())
	origID, origYes, origRec := choreID, yes, choreDeleteRecurring
	choreID, yes, choreDeleteRecurring = "c1", true, true
	t.Cleanup(func() { choreID, yes, choreDeleteRecurring = origID, origYes, origRec })
	resetFlags(t, choreDeleteCmd, "apply-to-profiles")
	setFlag(t, choreDeleteCmd, "apply-to-profiles", "all")
	if err := choreDeleteCmd.RunE(choreDeleteCmd, nil); err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutual exclusion error, got %v", err)
	}
}

func TestPrintChoresTable_LinkedColumn(t *testing.T) {
	out := captureStdout(func() {
		printChoresTable([]lib.Chore{
			{ID: "1", Title: "Bins", LinkedTaskGroupID: "g9"},
			{ID: "2", Title: "Solo"},
		})
	})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 || !strings.HasSuffix(lines[0], "LINKED") ||
		!strings.HasSuffix(lines[1], "g9") || !strings.HasSuffix(lines[2], "-") {
		t.Errorf("unexpected LINKED column output:\n%s", out)
	}
}
