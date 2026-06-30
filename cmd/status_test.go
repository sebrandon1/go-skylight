package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func statusMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[{"id":"1","attributes":{"label":"Mom","color":"#FF0000"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/reward_points"):
			fmt.Fprint(w, `[{"category_id":1,"current_point_balance":42}]`)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"pending"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}]}`)
		case strings.HasSuffix(r.URL.Path, "/test-frame"):
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestStatusCmd_Text(t *testing.T) {
	newCmdTestClient(t, statusMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	out := captureStdout(func() { statusCmd.Run(statusCmd, nil) })

	if !strings.Contains(out, "Kitchen") {
		t.Errorf("expected frame name in output, got: %s", out)
	}
	if !strings.Contains(out, "Mom: 42") {
		t.Errorf("expected resolved category name with points, got: %s", out)
	}
}

func TestStatusCmd_JSON(t *testing.T) {
	newCmdTestClient(t, statusMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	out := captureStdout(func() { statusCmd.Run(statusCmd, nil) })

	if !strings.Contains(out, `"frame": "Kitchen"`) {
		t.Errorf("expected frame name in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"pending_chores": 1`) {
		t.Errorf("expected pending chore count in JSON output, got: %s", out)
	}
}

func TestStatusCmd_NoPoints(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/reward_points"):
			fmt.Fprint(w, `[]`)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	})
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	out := captureStdout(func() { statusCmd.Run(statusCmd, nil) })

	if !strings.Contains(out, "Points:  none") {
		t.Errorf("expected 'none' for empty points, got: %s", out)
	}
}

func TestStatusCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("status command not registered on root")
	}
}
