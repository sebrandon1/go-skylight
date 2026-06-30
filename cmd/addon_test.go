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

	out := captureStdout(func() { addonListCmd.Run(addonListCmd, nil) })
	if !strings.Contains(out, "albums") {
		t.Errorf("expected feature bundle in output, got: %s", out)
	}
}

func TestAddonListCmd_EmptyBundle(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","plus":true}}}`)
	})

	out := captureStdout(func() { addonListCmd.Run(addonListCmd, nil) })
	if !strings.Contains(out, "No add-ons found.") {
		t.Errorf("expected no-addons message, got: %s", out)
	}
}

func TestAddonListCmd_WarnsWithoutPlus(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","plus":false}}}`)
	})

	stderr := captureStderr(func() { addonListCmd.Run(addonListCmd, nil) })
	if !strings.Contains(stderr, "Plus subscription required") {
		t.Errorf("expected Plus subscription warning on stderr, got: %s", stderr)
	}
}

func TestAddonCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "addon" {
			found = true
			break
		}
	}
	if !found {
		t.Error("addon command not registered on root")
	}
}
