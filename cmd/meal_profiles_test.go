package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

func TestMealUpdateSittingCmd_ProfileIDs(t *testing.T) {
	var body string
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"sitting1","type":"meal_sitting","attributes":{"summary":"dinner"},"relationships":{"profiles":{"data":[{"id":"6751001","type":"category"}]}}}}`)
	})
	origID, origDate := sittingID, sittingDate
	sittingID, sittingDate = "sitting1", "2026-11-17"
	t.Cleanup(func() { sittingID, sittingDate, sittingProfileIDs = origID, origDate, nil })

	if err := mealUpdateSittingCmd.Flags().Set(flagProfileID, "6751001"); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(func() {
		if err := mealUpdateSittingCmd.RunE(mealUpdateSittingCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(body, `"profile_ids":["6751001"]`) {
		t.Errorf("body = %s", body)
	}
	if !strings.Contains(out, "6751001") {
		t.Errorf("output = %s", out)
	}
}

func TestMealUpdateSittingCmd_ClearProfiles(t *testing.T) {
	var body string
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"sitting1","type":"meal_sitting","attributes":{"summary":"dinner"}}}`)
	})
	origID, origDate := sittingID, sittingDate
	sittingID, sittingDate = "sitting1", "2026-11-17"
	t.Cleanup(func() { sittingID, sittingDate, sittingProfileIDs = origID, origDate, nil })

	if err := mealUpdateSittingCmd.Flags().Set(flagProfileID, ""); err != nil {
		t.Fatal(err)
	}
	captureStdout(func() {
		if err := mealUpdateSittingCmd.RunE(mealUpdateSittingCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(body, `"profile_ids":[]`) {
		t.Errorf("body = %s", body)
	}
}

func TestMealSittingCmds_InvalidDate(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	origID, origDate := sittingID, sittingDate
	sittingID, sittingDate = "sitting1", "not-a-date"
	t.Cleanup(func() { sittingID, sittingDate = origID, origDate })

	for _, c := range []*cobra.Command{mealUpdateSittingCmd, mealGetSittingCmd, mealSittingRecipeCmd} {
		if err := c.RunE(c, nil); err == nil {
			t.Errorf("%s: expected invalid date error", c.Name())
		}
	}
}

func TestImportSittings_ProfileIDs(t *testing.T) {
	var body string
	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"s1","type":"meal_sitting","attributes":{"summary":"x"}}]}`)
	})
	total, failed := importSittings(context.Background(), client, []lib.MealSitting{{Summary: "x", MealCategoryID: "3", Date: "2026-10-01", ProfileIDs: []string{"7", "8"}}}, nil)
	if total != 1 || failed != 0 {
		t.Fatalf("total=%d failed=%d", total, failed)
	}
	if !strings.Contains(body, `"profile_ids":["7","8"]`) {
		t.Errorf("body = %s", body)
	}
}
