package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func bountyMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/chores") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Dishes","reward_points":10}}}`)
		case strings.HasSuffix(r.URL.Path, "/chores") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"pending","reward_points":10}}]}`)
		case strings.HasSuffix(r.URL.Path, "/c1") && r.Method == http.MethodPut:
			fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Updated","reward_points":10}}}`)
		case strings.HasSuffix(r.URL.Path, "/c1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/rewards") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"Ice cream","point_value":10}}]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"Ice cream","point_value":10}}]}`)
		case strings.HasSuffix(r.URL.Path, "/r1") && r.Method == http.MethodPatch:
			fmt.Fprint(w, `{"data":{"id":"r1","attributes":{"name":"Updated Reward","point_value":10}}}`)
		case strings.HasSuffix(r.URL.Path, "/r1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestBountyCreateCmd(t *testing.T) {
	newCmdTestClient(t, bountyMockHandler())
	origTitle, origPoints, origRewardTitle := bountyTitle, bountyPoints, bountyRewardTitle
	bountyTitle, bountyPoints, bountyRewardTitle = "Dishes", 10, "Ice cream"
	t.Cleanup(func() { bountyTitle, bountyPoints, bountyRewardTitle = origTitle, origPoints, origRewardTitle })

	out := captureStdout(func() {
		if err := bountyCreateCmd.RunE(bountyCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dishes") || !strings.Contains(out, "Ice cream") {
		t.Errorf("expected chore and reward in output, got: %s", out)
	}
}

func TestBountyListCmd(t *testing.T) {
	newCmdTestClient(t, bountyMockHandler())

	out := captureStdout(func() {
		if err := bountyListCmd.RunE(bountyListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected matched bounty in output, got: %s", out)
	}
}

func TestBountyDeleteCmd(t *testing.T) {
	newCmdTestClient(t, bountyMockHandler())
	origChoreID, origRewardID, origYes := bountyChoreID, bountyRewardID, yes
	bountyChoreID, bountyRewardID, yes = "c1", "r1", true
	t.Cleanup(func() { bountyChoreID, bountyRewardID, yes = origChoreID, origRewardID, origYes })

	out := captureStdout(func() {
		if err := bountyDeleteCmd.RunE(bountyDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Bounty deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestBountyDeleteCmd_DryRun(t *testing.T) {
	origChoreID, origRewardID, origDryRun := bountyChoreID, bountyRewardID, dryRun
	bountyChoreID, bountyRewardID, dryRun = "c1", "r1", true
	t.Cleanup(func() { bountyChoreID, bountyRewardID, dryRun = origChoreID, origRewardID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := bountyDeleteCmd.RunE(bountyDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestBountyUpdateCmd(t *testing.T) {
	newCmdTestClient(t, bountyMockHandler())
	origChoreID, origRewardID, origTitle := bountyChoreID, bountyRewardID, bountyTitle
	bountyChoreID, bountyRewardID, bountyTitle = "c1", "r1", "Updated"
	t.Cleanup(func() { bountyChoreID, bountyRewardID, bountyTitle = origChoreID, origRewardID, origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := bountyUpdateCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() {
		if err := bountyUpdateCmd.RunE(bountyUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated bounty in output, got: %s", out)
	}
}

func TestBountyCreateCmd_InvalidDate(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot) // won't be reached
	})
	orig := bountyDueDate
	bountyDueDate = "not-a-date"
	t.Cleanup(func() { bountyDueDate = orig })

	err := bountyCreateCmd.RunE(bountyCreateCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid date, got nil")
	}
	if !strings.Contains(err.Error(), "due-date") && !strings.Contains(err.Error(), "YYYY-MM-DD") {
		t.Errorf("expected date validation error, got: %v", err)
	}
}

func TestBountyCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "bounty")
}
