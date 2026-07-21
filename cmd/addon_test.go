package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
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
