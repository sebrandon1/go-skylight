package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func rewardMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/reward1/redeem"):
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/reward1/unredeem"):
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/reward1") && r.Method == http.MethodPatch:
			fmt.Fprint(w, `{"data":{"id":"reward1","attributes":{"name":"Updated","point_value":99}}}`)
		case strings.HasSuffix(r.URL.Path, "/reward1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/reward_points"):
			fmt.Fprint(w, `[{"category_id":1,"current_point_balance":5}]`)
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[{"id":"1","attributes":{"label":"Mom","color":"#FF0000"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":[{"id":"reward1","attributes":{"name":"Ice cream","point_value":10}}]}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"reward1","attributes":{"name":"Ice cream","point_value":10}}]}`)
		}
	}
}

func TestRewardListCmd(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Ice cream") {
		t.Errorf("expected reward in output, got: %s", out)
	}
}

func rewardFilterMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// CategoryID comes from relationships.category.data.id in JSON-API format.
		fmt.Fprint(w, `{"data":[`+
			`{"id":"r1","attributes":{"name":"Cheap","point_value":10},"relationships":{"category":{"data":{"id":"cat1","type":"category"}}}},`+
			`{"id":"r2","attributes":{"name":"Pricey","point_value":50},"relationships":{"category":{"data":{"id":"cat2","type":"category"}}}}`+
			`]}`)
	}
}

func TestRewardListCmd_FilterByAssigneeID(t *testing.T) {
	newCmdTestClient(t, rewardFilterMockHandler())
	origAssigneeID := rewardListAssigneeID
	rewardListAssigneeID = "cat1"
	t.Cleanup(func() { rewardListAssigneeID = origAssigneeID })

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Cheap") {
		t.Errorf("expected matching reward in output, got: %s", out)
	}
	if strings.Contains(out, "Pricey") {
		t.Errorf("expected non-matching reward to be filtered out, got: %s", out)
	}
}

func TestRewardListCmd_FilterByPointsMin(t *testing.T) {
	newCmdTestClient(t, rewardFilterMockHandler())
	origMin := rewardListPointsMin
	rewardListPointsMin = 30
	t.Cleanup(func() { rewardListPointsMin = origMin })

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Pricey") {
		t.Errorf("expected high-points reward in output, got: %s", out)
	}
	if strings.Contains(out, "Cheap") {
		t.Errorf("expected low-points reward to be filtered out, got: %s", out)
	}
}

func TestRewardListCmd_FilterByPointsMax(t *testing.T) {
	newCmdTestClient(t, rewardFilterMockHandler())
	origMax := rewardListPointsMax
	rewardListPointsMax = 20
	t.Cleanup(func() { rewardListPointsMax = origMax })

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Cheap") {
		t.Errorf("expected low-points reward in output, got: %s", out)
	}
	if strings.Contains(out, "Pricey") {
		t.Errorf("expected high-points reward to be filtered out, got: %s", out)
	}
}

func TestRewardListCmd_FilterCombined(t *testing.T) {
	newCmdTestClient(t, rewardFilterMockHandler())
	origAssigneeID, origMax := rewardListAssigneeID, rewardListPointsMax
	rewardListAssigneeID = "cat1"
	rewardListPointsMax = 20
	t.Cleanup(func() {
		rewardListAssigneeID = origAssigneeID
		rewardListPointsMax = origMax
	})

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Cheap") {
		t.Errorf("expected matching reward in output, got: %s", out)
	}
	if strings.Contains(out, "Pricey") {
		t.Errorf("expected non-matching reward to be filtered out, got: %s", out)
	}
}

func TestRewardListCmd_FilterNoMatches(t *testing.T) {
	newCmdTestClient(t, rewardFilterMockHandler())
	origAssigneeID := rewardListAssigneeID
	rewardListAssigneeID = "cat-nonexistent"
	t.Cleanup(func() { rewardListAssigneeID = origAssigneeID })

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "[]") {
		t.Errorf("expected empty JSON array when no rewards match, got: %s", out)
	}
}

func TestRewardCreateCmd(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())
	origTitle, origPoints := rewardTitle, rewardPoints
	rewardTitle, rewardPoints = "Ice cream", 10
	t.Cleanup(func() { rewardTitle, rewardPoints = origTitle, origPoints })

	out := captureStdout(func() {
		if err := rewardCreateCmd.RunE(rewardCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Ice cream") {
		t.Errorf("expected created reward in output, got: %s", out)
	}
}

func TestRewardCreateCmd_WithOptionalFields(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())
	origTitle, origPoints, origEmoji, origNoRespawn, origCatIDs :=
		rewardTitle, rewardPoints, rewardEmojiIcon, rewardNoRespawn, rewardCategoryIDs
	rewardTitle, rewardPoints, rewardEmojiIcon, rewardNoRespawn, rewardCategoryIDs =
		"Ice cream", 10, "🍦", true, []int{1, 2}
	t.Cleanup(func() {
		rewardTitle, rewardPoints, rewardEmojiIcon, rewardNoRespawn, rewardCategoryIDs =
			origTitle, origPoints, origEmoji, origNoRespawn, origCatIDs
	})

	out := captureStdout(func() {
		if err := rewardCreateCmd.RunE(rewardCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Ice cream") {
		t.Errorf("expected created reward in output, got: %s", out)
	}
}

func TestRewardDeleteCmd(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())
	origID := rewardID
	rewardID = "reward1"
	t.Cleanup(func() { rewardID = origID })

	out := captureStdout(func() {
		if err := rewardDeleteCmd.RunE(rewardDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestRewardDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := rewardID, dryRun
	rewardID, dryRun = "reward1", true
	t.Cleanup(func() { rewardID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := rewardDeleteCmd.RunE(rewardDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestRewardRedeemCmd(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())
	origID := rewardID
	rewardID = "reward1"
	t.Cleanup(func() { rewardID = origID })

	out := captureStdout(func() {
		if err := rewardRedeemCmd.RunE(rewardRedeemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "redeemed successfully") {
		t.Errorf("expected redeem confirmation, got: %s", out)
	}
}

func TestRewardRedeemCmd_DryRun(t *testing.T) {
	origID, origDryRun := rewardID, dryRun
	rewardID, dryRun = "reward1", true
	t.Cleanup(func() { rewardID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := rewardRedeemCmd.RunE(rewardRedeemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestRewardUnredeemCmd(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())
	origID := rewardID
	rewardID = "reward1"
	t.Cleanup(func() { rewardID = origID })

	out := captureStdout(func() {
		if err := rewardUnredeemCmd.RunE(rewardUnredeemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "unredeemed successfully") {
		t.Errorf("expected unredeem confirmation, got: %s", out)
	}
}

func TestRewardUnredeemCmd_DryRun(t *testing.T) {
	origID, origDryRun := rewardID, dryRun
	rewardID, dryRun = "reward1", true
	t.Cleanup(func() { rewardID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := rewardUnredeemCmd.RunE(rewardUnredeemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestRewardPointsCmd(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())

	out := captureStdout(func() {
		if err := rewardPointsCmd.RunE(rewardPointsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, `"name": "Mom"`) {
		t.Errorf("expected resolved category name in output, got: %s", out)
	}
	if !strings.Contains(out, `"balance": 5`) {
		t.Errorf("expected point balance in output, got: %s", out)
	}
}

func TestRewardUpdateCmd(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())
	origID, origTitle := rewardID, rewardTitle
	rewardID, rewardTitle = "reward1", "Updated"
	t.Cleanup(func() { rewardID, rewardTitle = origID, origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := rewardUpdateCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() {
		if err := rewardUpdateCmd.RunE(rewardUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated reward in output, got: %s", out)
	}
}

func TestRewardUpdateCmd_WithOptionalFields(t *testing.T) {
	newCmdTestClient(t, rewardMockHandler())
	origID, origNoRespawn, origCatIDs := rewardID, rewardNoRespawn, rewardCategoryIDs
	rewardID, rewardNoRespawn, rewardCategoryIDs = "reward1", true, []int{1, 2}
	t.Cleanup(func() {
		rewardID, rewardNoRespawn, rewardCategoryIDs = origID, origNoRespawn, origCatIDs
	})

	// Mark optional update flags as changed so Run includes them in RewardData.
	if err := rewardUpdateCmd.Flags().Set("no-respawn", "true"); err != nil {
		t.Fatalf("setting no-respawn flag: %v", err)
	}
	if err := rewardUpdateCmd.Flags().Set("category-ids", "1,2"); err != nil {
		t.Fatalf("setting category-ids flag: %v", err)
	}

	out := captureStdout(func() {
		if err := rewardUpdateCmd.RunE(rewardUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") && !strings.Contains(out, "reward1") {
		t.Errorf("expected updated reward in output, got: %s", out)
	}
}

func rewardStatusMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[`+
			`{"id":"r1","attributes":{"name":"Available Reward","point_value":10}},`+
			`{"id":"r2","attributes":{"name":"Redeemed Reward","point_value":20,"redeemed_at":"2026-01-01T00:00:00Z"}}`+
			`]}`)
	}
}

func TestRewardListCmd_FilterByStatusAvailable(t *testing.T) {
	newCmdTestClient(t, rewardStatusMockHandler())
	origStatus := rewardListStatus
	rewardListStatus = "available"
	t.Cleanup(func() { rewardListStatus = origStatus })

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Available Reward") {
		t.Errorf("expected available reward in output, got: %s", out)
	}
	if strings.Contains(out, "Redeemed Reward") {
		t.Errorf("expected redeemed reward to be filtered out, got: %s", out)
	}
}

func TestRewardListCmd_FilterByStatusRedeemed(t *testing.T) {
	newCmdTestClient(t, rewardStatusMockHandler())
	origStatus := rewardListStatus
	rewardListStatus = "redeemed"
	t.Cleanup(func() { rewardListStatus = origStatus })

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Redeemed Reward") {
		t.Errorf("expected redeemed reward in output, got: %s", out)
	}
	if strings.Contains(out, "Available Reward") {
		t.Errorf("expected available reward to be filtered out, got: %s", out)
	}
}

func TestRewardListCmd_FilterByStatusInvalid(t *testing.T) {
	origStatus := rewardListStatus
	rewardListStatus = "unknown"
	t.Cleanup(func() { rewardListStatus = origStatus })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	var runErr error
	captureStdout(func() {
		runErr = rewardListCmd.RunE(rewardListCmd, nil)
	})
	if runErr == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestRewardListCmd_FilterByStatusCombinedWithPoints(t *testing.T) {
	newCmdTestClient(t, rewardStatusMockHandler())
	origStatus, origMin := rewardListStatus, rewardListPointsMin
	rewardListStatus = "redeemed"
	rewardListPointsMin = 15
	t.Cleanup(func() {
		rewardListStatus = origStatus
		rewardListPointsMin = origMin
	})

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Redeemed Reward") {
		t.Errorf("expected redeemed reward with sufficient points in output, got: %s", out)
	}
	if strings.Contains(out, "Available Reward") {
		t.Errorf("expected available reward to be filtered out, got: %s", out)
	}
}

func TestRewardListCmd_FilterByStatusAndPointsNoMatch(t *testing.T) {
	newCmdTestClient(t, rewardStatusMockHandler())
	origStatus, origMin := rewardListStatus, rewardListPointsMin
	rewardListStatus = "available"
	rewardListPointsMin = 15
	t.Cleanup(func() {
		rewardListStatus = origStatus
		rewardListPointsMin = origMin
	})

	out := captureStdout(func() {
		if err := rewardListCmd.RunE(rewardListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "[]") {
		t.Errorf("expected empty result when no available rewards meet points threshold, got: %s", out)
	}
}

func TestRewardCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "reward")
}
