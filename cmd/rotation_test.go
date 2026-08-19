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

	out := captureStdout(func() {
		if err := rotationCreateCmd.RunE(rotationCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected created rotation chores in output, got: %s", out)
	}
}

func TestRotationCreateCmd_APIError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	orig := rotationChores
	rotationChores = []string{"Dishes"}
	t.Cleanup(func() { rotationChores = orig })

	err := rotationCreateCmd.RunE(rotationCreateCmd, nil)
	if err == nil {
		t.Fatal("expected error when rotation API returns 500, got nil")
	}
	if !strings.Contains(err.Error(), "creating rotation") {
		t.Errorf("expected 'creating rotation' in error, got: %v", err)
	}
}

func TestRotationCreateCmd_TableOutput(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Dishes"}}}`)
	})

	origChores, origAssignees, origStart, origWeeks, origFmt :=
		rotationChores, rotationAssigneeIDs, rotationStartDate, rotationWeeks, outputFormat
	rotationChores = []string{"Dishes"}
	rotationAssigneeIDs = []string{"a1", "a2"}
	rotationStartDate = "2026-01-01"
	rotationWeeks = 1
	outputFormat = outputTable
	t.Cleanup(func() {
		rotationChores, rotationAssigneeIDs, rotationStartDate, rotationWeeks, outputFormat =
			origChores, origAssignees, origStart, origWeeks, origFmt
	})

	out := captureStdout(func() {
		if err := rotationCreateCmd.RunE(rotationCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "TITLE") {
		t.Errorf("expected table header in output, got: %s", out)
	}
	if !strings.Contains(out, "Dishes") {
		t.Errorf("expected chore title in table, got: %s", out)
	}
}

func TestRotationCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "rotation")
}
