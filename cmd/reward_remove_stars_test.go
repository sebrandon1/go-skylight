package cmd

import (
	"fmt"
	"net/http"
	"os"
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

	out := captureStdout(func() { rewardRemoveStarsCmd.Run(rewardRemoveStarsCmd, nil) })
	if !strings.Contains(out, "current_point_balance") {
		t.Errorf("expected updated points in output, got: %s", out)
	}
}

func TestRewardRemoveStarsCmdExists(t *testing.T) {
	assertCommandRegistered(t, rewardCmd, "remove-stars")
}

// TestRewardRemoveStarsCmd_InvalidAssigneeID_Crasher is invoked as a subprocess
// by TestRewardRemoveStarsCmd_InvalidAssigneeID to exercise the fatal() path
// for a non-positive assignee ID without terminating the real test binary.
func TestRewardRemoveStarsCmd_InvalidAssigneeID_Crasher(t *testing.T) {
	if os.Getenv("WANT_REMOVE_STARS_ASSIGNEE_CRASH") != "1" {
		t.Skip("only runs as a subprocess of TestRewardRemoveStarsCmd_InvalidAssigneeID")
	}
	frameID = "test-frame"
	removeStarsAssigneeID = -1
	removeStarsPoints = 10
	rewardRemoveStarsCmd.Run(rewardRemoveStarsCmd, nil)
}

func TestRewardRemoveStarsCmd_InvalidAssigneeID(t *testing.T) {
	runCrasherTest(t, "TestRewardRemoveStarsCmd_InvalidAssigneeID_Crasher", "WANT_REMOVE_STARS_ASSIGNEE_CRASH", "assignee-id")
}

// TestRewardRemoveStarsCmd_InvalidPoints_Crasher is invoked as a subprocess
// by TestRewardRemoveStarsCmd_InvalidPoints to exercise the fatal() path
// for a non-positive points value without terminating the real test binary.
func TestRewardRemoveStarsCmd_InvalidPoints_Crasher(t *testing.T) {
	if os.Getenv("WANT_REMOVE_STARS_POINTS_CRASH") != "1" {
		t.Skip("only runs as a subprocess of TestRewardRemoveStarsCmd_InvalidPoints")
	}
	frameID = "test-frame"
	removeStarsAssigneeID = 1
	removeStarsPoints = -5
	rewardRemoveStarsCmd.Run(rewardRemoveStarsCmd, nil)
}

func TestRewardRemoveStarsCmd_InvalidPoints(t *testing.T) {
	runCrasherTest(t, "TestRewardRemoveStarsCmd_InvalidPoints_Crasher", "WANT_REMOVE_STARS_POINTS_CRASH", "points")
}

// TestRewardRemoveStarsCmd_APIError_Crasher is invoked as a subprocess by
// TestRewardRemoveStarsCmd_APIError to exercise the fatal() path when the
// RemoveStars API call returns a server error.
func TestRewardRemoveStarsCmd_APIError_Crasher(t *testing.T) {
	if os.Getenv("WANT_REMOVE_STARS_API_CRASH") != "1" {
		t.Skip("only runs as a subprocess of TestRewardRemoveStarsCmd_APIError")
	}
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	removeStarsAssigneeID = 1
	removeStarsPoints = 10
	rewardRemoveStarsCmd.Run(rewardRemoveStarsCmd, nil)
}

func TestRewardRemoveStarsCmd_APIError(t *testing.T) {
	runCrasherTest(t, "TestRewardRemoveStarsCmd_APIError_Crasher", "WANT_REMOVE_STARS_API_CRASH", "removing stars")
}
