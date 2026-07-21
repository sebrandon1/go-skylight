package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestChoreStreakCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[{"id":"1","attributes":{"label":"Mom","color":"#FF0000"}}]}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"complete","due_date":"2026-01-01"},"relationships":{"category":{"data":{"id":"1"}}}}]}`)
		}
	})

	out := captureStdout(func() {
		if err := choreStreakCmd.RunE(choreStreakCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if out == "" {
		t.Error("expected non-empty streak output")
	}
}

func TestChoreStreakCmd_ChoreFetchError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[]}`)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	err := choreStreakCmd.RunE(choreStreakCmd, nil)
	if err == nil {
		t.Fatal("expected error when chore API returns 500, got nil")
	}
	if !strings.Contains(err.Error(), "listing chores") {
		t.Errorf("expected 'listing chores' in error, got: %v", err)
	}
}

func TestChoreStreakCmdExists(t *testing.T) {
	assertCommandRegistered(t, choreCmd, "streak")
}
