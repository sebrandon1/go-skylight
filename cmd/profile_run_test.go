package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func profileMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/p1") && r.Method == http.MethodPatch:
			fmt.Fprint(w, `{"data":{"id":"p1","attributes":{"label":"Updated","color":"#00FF00","linked_to_profile":true}}}`)
		case strings.HasSuffix(r.URL.Path, "/p1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/categories") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"p1","attributes":{"label":"Alice","color":"#FF0000","linked_to_profile":true}}}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{"label":"Alice","color":"#FF0000","linked_to_profile":true}}]}`)
		}
	}
}

func TestProfileListCmd(t *testing.T) {
	newCmdTestClient(t, profileMockHandler())

	out := captureStdout(func() {
		if err := profileListCmd.RunE(profileListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected profile name in output, got: %s", out)
	}
}

func TestProfileCreateCmd(t *testing.T) {
	newCmdTestClient(t, profileMockHandler())
	origName := profileName
	profileName = "Alice"
	t.Cleanup(func() { profileName = origName })

	out := captureStdout(func() {
		if err := profileCreateCmd.RunE(profileCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected created profile in output, got: %s", out)
	}
}

func TestProfileUpdateCmd(t *testing.T) {
	newCmdTestClient(t, profileMockHandler())
	origID, origName := profileID, profileName
	profileID, profileName = "p1", "Updated"
	t.Cleanup(func() { profileID, profileName = origID, origName })

	if err := profileUpdateCmd.Flags().Set("name", "Updated"); err != nil {
		t.Fatalf("setting name flag: %v", err)
	}

	out := captureStdout(func() {
		if err := profileUpdateCmd.RunE(profileUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated profile in output, got: %s", out)
	}
}

func TestProfileDeleteCmd(t *testing.T) {
	newCmdTestClient(t, profileMockHandler())
	origID, origYes := profileID, yes
	profileID, yes = "p1", true
	t.Cleanup(func() { profileID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := profileDeleteCmd.RunE(profileDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestProfileDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := profileID, dryRun
	profileID, dryRun = "p1", true
	t.Cleanup(func() { profileID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := profileDeleteCmd.RunE(profileDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestProfileCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "profile")
}
