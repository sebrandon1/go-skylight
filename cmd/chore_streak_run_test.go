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

	out := captureStdout(func() { choreStreakCmd.Run(choreStreakCmd, nil) })
	if out == "" {
		t.Error("expected non-empty streak output")
	}
}

func TestChoreStreakCmdExists(t *testing.T) {
	assertCommandRegistered(t, choreCmd, "streak")
}
