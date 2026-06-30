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

	out := captureStdout(func() { rewardRemoveStarsCmd.Run(rewardRemoveStarsCmd, nil) })
	if !strings.Contains(out, "current_point_balance") {
		t.Errorf("expected updated points in output, got: %s", out)
	}
}

func TestRewardRemoveStarsCmdExists(t *testing.T) {
	assertCommandRegistered(t, rewardCmd, "remove-stars")
}
