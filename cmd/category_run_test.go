package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func categoryMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/cat1") && r.Method == http.MethodPatch:
			fmt.Fprint(w, `{"data":{"id":"cat1","attributes":{"label":"Updated","color":"#00FF00"}}}`)
		case strings.HasSuffix(r.URL.Path, "/cat1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/categories") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"cat1","attributes":{"label":"Mom","color":"#FF0000"}}}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"cat1","attributes":{"label":"Mom","color":"#FF0000"}}]}`)
		}
	}
}

func TestCategoryListCmd(t *testing.T) {
	newCmdTestClient(t, categoryMockHandler())

	out := captureStdout(func() {
		if err := categoryListCmd.RunE(categoryListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Mom") {
		t.Errorf("expected category name in output, got: %s", out)
	}
}

func TestCategoryCreateCmd(t *testing.T) {
	newCmdTestClient(t, categoryMockHandler())
	origName := categoryName
	categoryName = "Mom"
	t.Cleanup(func() { categoryName = origName })

	out := captureStdout(func() {
		if err := categoryCreateCmd.RunE(categoryCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Mom") {
		t.Errorf("expected created category in output, got: %s", out)
	}
}

func TestCategoryDeleteCmd(t *testing.T) {
	newCmdTestClient(t, categoryMockHandler())
	origID, origYes := categoryID, yes
	categoryID, yes = "cat1", true
	t.Cleanup(func() { categoryID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := categoryDeleteCmd.RunE(categoryDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestCategoryDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := categoryID, dryRun
	categoryID, dryRun = "cat1", true
	t.Cleanup(func() { categoryID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := categoryDeleteCmd.RunE(categoryDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestCategoryUpdateCmd(t *testing.T) {
	newCmdTestClient(t, categoryMockHandler())
	origID, origName := categoryID, categoryName
	categoryID, categoryName = "cat1", "Updated"
	t.Cleanup(func() { categoryID, categoryName = origID, origName })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := categoryUpdateCmd.Flags().Set("name", "Updated"); err != nil {
		t.Fatalf("setting name flag: %v", err)
	}

	out := captureStdout(func() {
		if err := categoryUpdateCmd.RunE(categoryUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated category in output, got: %s", out)
	}
}

func TestCategoryCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "category")
}
