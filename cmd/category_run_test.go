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

	out := captureStdout(func() { categoryListCmd.Run(categoryListCmd, nil) })
	if !strings.Contains(out, "Mom") {
		t.Errorf("expected category name in output, got: %s", out)
	}
}

func TestCategoryCreateCmd(t *testing.T) {
	newCmdTestClient(t, categoryMockHandler())
	origName := categoryName
	categoryName = "Mom"
	t.Cleanup(func() { categoryName = origName })

	out := captureStdout(func() { categoryCreateCmd.Run(categoryCreateCmd, nil) })
	if !strings.Contains(out, "Mom") {
		t.Errorf("expected created category in output, got: %s", out)
	}
}

func TestCategoryDeleteCmd(t *testing.T) {
	newCmdTestClient(t, categoryMockHandler())
	origID := categoryID
	categoryID = "cat1"
	t.Cleanup(func() { categoryID = origID })

	out := captureStdout(func() { categoryDeleteCmd.Run(categoryDeleteCmd, nil) })
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
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

	out := captureStdout(func() { categoryUpdateCmd.Run(categoryUpdateCmd, nil) })
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated category in output, got: %s", out)
	}
}

func TestCategoryCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "category" {
			found = true
			break
		}
	}
	if !found {
		t.Error("category command not registered on root")
	}
}
