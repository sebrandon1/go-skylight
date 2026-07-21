package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestRewardRemoveStarsCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Fprint(w, `[{"category_id":1,"current_point_balance":5}]`)
	})

	origAssignee, origPoints := removeStarsAssigneeID, removeStarsPoints
	removeStarsAssigneeID, removeStarsPoints = 1, 10
	t.Cleanup(func() { removeStarsAssigneeID, removeStarsPoints = origAssignee, origPoints })

	out := captureStdout(func() {
		if err := rewardRemoveStarsCmd.RunE(rewardRemoveStarsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "current_point_balance") {
		t.Errorf("expected updated points in output, got: %s", out)
	}
}

func TestRewardRemoveStarsCmdExists(t *testing.T) {
	assertCommandRegistered(t, rewardCmd, "remove-stars")
}

func TestRewardRemoveStarsCmd_InvalidAssigneeID(t *testing.T) {
	origFrameID, origAssignee, origPoints := frameID, removeStarsAssigneeID, removeStarsPoints
	frameID = "test-frame"
	removeStarsAssigneeID = -1
	removeStarsPoints = 10
	t.Cleanup(func() { frameID, removeStarsAssigneeID, removeStarsPoints = origFrameID, origAssignee, origPoints })

	err := rewardRemoveStarsCmd.RunE(rewardRemoveStarsCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid assignee-id, got nil")
	}
	if !strings.Contains(err.Error(), "assignee-id") {
		t.Errorf("expected 'assignee-id' in error, got: %v", err)
	}
}

func TestRewardRemoveStarsCmd_InvalidPoints(t *testing.T) {
	origFrameID, origAssignee, origPoints := frameID, removeStarsAssigneeID, removeStarsPoints
	frameID = "test-frame"
	removeStarsAssigneeID = 1
	removeStarsPoints = -5
	t.Cleanup(func() { frameID, removeStarsAssigneeID, removeStarsPoints = origFrameID, origAssignee, origPoints })

	err := rewardRemoveStarsCmd.RunE(rewardRemoveStarsCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid points, got nil")
	}
	if !strings.Contains(err.Error(), "points") {
		t.Errorf("expected 'points' in error, got: %v", err)
	}
}

func TestRewardRemoveStarsCmd_APIError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	origAssignee, origPoints := removeStarsAssigneeID, removeStarsPoints
	removeStarsAssigneeID = 1
	removeStarsPoints = 10
	t.Cleanup(func() { removeStarsAssigneeID, removeStarsPoints = origAssignee, origPoints })

	err := rewardRemoveStarsCmd.RunE(rewardRemoveStarsCmd, nil)
	if err == nil {
		t.Fatal("expected error for API failure, got nil")
	}
	if !strings.Contains(err.Error(), "removing stars") {
		t.Errorf("expected 'removing stars' in error, got: %v", err)
	}
}

func TestRewardRemoveStarsCmd_MissingFrameID(t *testing.T) {
	origFrameID, origAssignee, origPoints := frameID, removeStarsAssigneeID, removeStarsPoints
	frameID = ""
	removeStarsAssigneeID = 1
	removeStarsPoints = 10
	t.Cleanup(func() { frameID, removeStarsAssigneeID, removeStarsPoints = origFrameID, origAssignee, origPoints })

	err := rewardRemoveStarsCmd.RunE(rewardRemoveStarsCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing frame ID, got nil")
	}
	if !strings.Contains(err.Error(), "--frame-id is required") {
		t.Errorf("expected '--frame-id is required' in error, got: %v", err)
	}
}

func TestRewardRemoveStarsCmd_PointsAPIError(t *testing.T) {
	first := true
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && first {
			first = false
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})
	origAssignee, origPoints := removeStarsAssigneeID, removeStarsPoints
	removeStarsAssigneeID = 1
	removeStarsPoints = 10
	t.Cleanup(func() { removeStarsAssigneeID, removeStarsPoints = origAssignee, origPoints })

	err := rewardRemoveStarsCmd.RunE(rewardRemoveStarsCmd, nil)
	if err == nil {
		t.Fatal("expected error fetching updated points, got nil")
	}
	if !strings.Contains(err.Error(), "fetching updated reward points") {
		t.Errorf("expected 'fetching updated reward points' in error, got: %v", err)
	}
}
