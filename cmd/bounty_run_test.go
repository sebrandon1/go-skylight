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

	out := captureStdout(func() { bountyCreateCmd.Run(bountyCreateCmd, nil) })
	if !strings.Contains(out, "Dishes") || !strings.Contains(out, "Ice cream") {
		t.Errorf("expected chore and reward in output, got: %s", out)
	}
}

func TestBountyListCmd(t *testing.T) {
	newCmdTestClient(t, bountyMockHandler())

	out := captureStdout(func() { bountyListCmd.Run(bountyListCmd, nil) })
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected matched bounty in output, got: %s", out)
	}
}

func TestBountyDeleteCmd(t *testing.T) {
	newCmdTestClient(t, bountyMockHandler())
	origChoreID, origRewardID := bountyChoreID, bountyRewardID
	bountyChoreID, bountyRewardID = "c1", "r1"
	t.Cleanup(func() { bountyChoreID, bountyRewardID = origChoreID, origRewardID })

	out := captureStdout(func() { bountyDeleteCmd.Run(bountyDeleteCmd, nil) })
	if !strings.Contains(out, "Bounty deleted.") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestBountyUpdateCmd(t *testing.T) {
	newCmdTestClient(t, bountyMockHandler())
	origChoreID, origRewardID, origTitle := bountyChoreID, bountyRewardID, bountyTitle
	bountyChoreID, bountyRewardID, bountyTitle = "c1", "r1", "Updated"
	t.Cleanup(func() { bountyChoreID, bountyRewardID, bountyTitle = origChoreID, origRewardID, origTitle })

	if err := bountyUpdateCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() { bountyUpdateCmd.Run(bountyUpdateCmd, nil) })
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated bounty in output, got: %s", out)
	}
}

func TestBountyCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "bounty" {
			found = true
			break
		}
	}
	if !found {
		t.Error("bounty command not registered on root")
	}
}
