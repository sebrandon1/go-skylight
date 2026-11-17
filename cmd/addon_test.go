package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAddonListCmd_PrintsBundle(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","plus":true,"feature_bundle":{"albums":{"enabled":true}}}}}`)
	})
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	out := captureStdout(func() {
		if err := addonListCmd.RunE(addonListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "albums") {
		t.Errorf("expected feature bundle in output, got: %s", out)
	}
}

func TestAddonListCmd_EmptyBundle(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","plus":true}}}`)
	})

	out := captureStdout(func() {
		if err := addonListCmd.RunE(addonListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No add-ons found.") {
		t.Errorf("expected no-addons message, got: %s", out)
	}
}

func TestAddonListCmd_WarnsWithoutPlus(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","plus":false}}}`)
	})

	stderr := captureStderr(func() {
		if err := addonListCmd.RunE(addonListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(stderr, "Plus subscription required") {
		t.Errorf("expected Plus subscription warning on stderr, got: %s", stderr)
	}
}

func TestAddonCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "addon")
}

func TestAddonSettingsCmd_Table(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":{"id":"1","type":"household_config","attributes":{"disney_profile_pictures":true,"disney_screensaver":false}}}`)
	})
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputTable

	out := captureStdout(func() {
		if err := addonSettingsCmd.RunE(addonSettingsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "disney-profile-pictures  yes") || !strings.Contains(out, "disney-screensaver       no") {
		t.Errorf("unexpected table:\n%s", out)
	}
}

func TestAddonSetCmd_SendsChangedFlagOnly(t *testing.T) {
	var body string
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		fmt.Fprint(w, `{"data":{"id":"1","type":"household_config","attributes":{"disney_profile_pictures":true,"disney_screensaver":false}}}`)
	})
	if err := addonSetCmd.Flags().Set(flagDisneyScreensaver, "false"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { addonSetCmd.Flags().Lookup(flagDisneyScreensaver).Changed = false })

	captureStdout(func() {
		if err := addonSetCmd.RunE(addonSetCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if body != `{"disney_screensaver":false}` {
		t.Errorf("PATCH body = %s", body)
	}
}

func TestAddonSetCmd_NoFlags(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request expected")
	})
	// A fresh command with the same flags, none of them set.
	cmd := &cobra.Command{}
	cmd.Flags().BoolVar(&addonDisneyProfilePictures, flagDisneyProfilePictures, false, "")
	cmd.Flags().BoolVar(&addonDisneyScreensaver, flagDisneyScreensaver, false, "")
	if err := addonSetCmd.RunE(cmd, nil); err == nil || !strings.Contains(err.Error(), "nothing to change") {
		t.Errorf("want nothing-to-change error, got %v", err)
	}
}

func TestAddonSettingsCmd_RequiresFrameID(t *testing.T) {
	old := frameID
	t.Cleanup(func() { frameID = old })
	frameID = ""
	if err := addonSettingsCmd.RunE(addonSettingsCmd, nil); err == nil {
		t.Error("expected frame-id error")
	}
}
