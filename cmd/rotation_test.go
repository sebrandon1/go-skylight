package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestRotationCreateCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Dishes"}}}`)
	})

	origChores, origAssignees, origStart, origWeeks := rotationChores, rotationAssigneeIDs, rotationStartDate, rotationWeeks
	rotationChores = []string{"Dishes"}
	rotationAssigneeIDs = []string{"a1", "a2"}
	rotationStartDate = "2026-01-01"
	rotationWeeks = 1
	t.Cleanup(func() {
		rotationChores, rotationAssigneeIDs, rotationStartDate, rotationWeeks =
			origChores, origAssignees, origStart, origWeeks
	})

	out := captureStdout(func() { rotationCreateCmd.Run(rotationCreateCmd, nil) })
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected created rotation chores in output, got: %s", out)
	}
}

func TestRotationCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "rotation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("rotation command not registered on root")
	}
}
