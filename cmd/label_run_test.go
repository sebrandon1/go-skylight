package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func labelMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/l1") && r.Method == http.MethodPatch:
			fmt.Fprint(w, `{"data":{"id":"l1","attributes":{"label":"Updated","color":"#00FF00","linked_to_profile":false}}}`)
		case strings.HasSuffix(r.URL.Path, "/l1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/categories") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"l1","attributes":{"label":"Sports","color":"#0000FF","linked_to_profile":false}}}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"l1","attributes":{"label":"Sports","color":"#0000FF","linked_to_profile":false}}]}`)
		}
	}
}

func TestLabelListCmd(t *testing.T) {
	newCmdTestClient(t, labelMockHandler())

	out := captureStdout(func() {
		if err := labelListCmd.RunE(labelListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Sports") {
		t.Errorf("expected label name in output, got: %s", out)
	}
}

func TestLabelCreateCmd(t *testing.T) {
	newCmdTestClient(t, labelMockHandler())
	origName := labelName
	labelName = "Sports"
	t.Cleanup(func() { labelName = origName })

	out := captureStdout(func() {
		if err := labelCreateCmd.RunE(labelCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Sports") {
		t.Errorf("expected created label in output, got: %s", out)
	}
}

func TestLabelUpdateCmd(t *testing.T) {
	newCmdTestClient(t, labelMockHandler())
	origID, origName := labelID, labelName
	labelID, labelName = "l1", "Updated"
	t.Cleanup(func() { labelID, labelName = origID, origName })

	if err := labelUpdateCmd.Flags().Set("name", "Updated"); err != nil {
		t.Fatalf("setting name flag: %v", err)
	}

	out := captureStdout(func() {
		if err := labelUpdateCmd.RunE(labelUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated label in output, got: %s", out)
	}
}

func TestLabelDeleteCmd(t *testing.T) {
	newCmdTestClient(t, labelMockHandler())
	origID, origYes := labelID, yes
	labelID, yes = "l1", true
	t.Cleanup(func() { labelID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := labelDeleteCmd.RunE(labelDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestLabelDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := labelID, dryRun
	labelID, dryRun = "l1", true
	t.Cleanup(func() { labelID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := labelDeleteCmd.RunE(labelDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestLabelCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "label")
}
