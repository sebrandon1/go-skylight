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
		case strings.HasSuffix(r.URL.Path, "/meals/sittings"):
			fmt.Fprint(w, `{"data":[{"id":"s1","type":"meal_sitting","attributes":{"summary":"Dinner"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/lists/l1"):
			fmt.Fprint(w, `{"data":{"id":"l1","attributes":{"label":"Groceries"}},"included":[{"id":"i1","type":"list_item","attributes":{"label":"Milk","status":"pending"}},{"id":"i2","type":"list_item","attributes":{"label":"Eggs","status":"completed"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[{"id":"l1","attributes":{"label":"Groceries"}}]}`)
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

	out := captureStdout(func() {
		if err := statusCmd.RunE(statusCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Kitchen") {
		t.Errorf("expected frame name in output, got: %s", out)
	}
	if !strings.Contains(out, "Mom: 42") {
		t.Errorf("expected resolved category name with points, got: %s", out)
	}
	if !strings.Contains(out, "Meals:   1 today") {
		t.Errorf("expected meal sitting count in output, got: %s", out)
	}
	if !strings.Contains(out, "Lists:   1 active, 1 incomplete items") {
		t.Errorf("expected list summary in output, got: %s", out)
	}
}

func TestStatusCmd_JSON(t *testing.T) {
	newCmdTestClient(t, statusMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	out := captureStdout(func() {
		if err := statusCmd.RunE(statusCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, `"frame": "Kitchen"`) {
		t.Errorf("expected frame name in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"pending_chores": 1`) {
		t.Errorf("expected pending chore count in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"meal_sittings_today": 1`) {
		t.Errorf("expected meal sitting count in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"active_lists": 1`) {
		t.Errorf("expected active list count in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"incomplete_list_items": 1`) {
		t.Errorf("expected incomplete list item count in JSON output, got: %s", out)
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
		case strings.HasSuffix(r.URL.Path, "/meals/sittings"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	})
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	out := captureStdout(func() {
		if err := statusCmd.RunE(statusCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Points:  none") {
		t.Errorf("expected 'none' for empty points, got: %s", out)
	}
}

func TestStatusCmd_ListErrorsSurfaced(t *testing.T) {
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
		case strings.HasSuffix(r.URL.Path, "/meals/sittings"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/lists/l1"):
			w.WriteHeader(http.StatusInternalServerError)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[{"id":"l1","attributes":{"label":"Groceries"}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	})
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	out := captureStdout(func() {
		if err := statusCmd.RunE(statusCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Lists:   1 active, 0 incomplete items (1 lists unavailable)") {
		t.Errorf("expected list-fetch failure to be surfaced, got: %s", out)
	}
}

func TestStatusCmd_GetFrameError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	err := statusCmd.RunE(statusCmd, nil)
	if err == nil {
		t.Fatal("expected error when frame API returns 500, got nil")
	}
	if !strings.Contains(err.Error(), "getting frame") {
		t.Errorf("expected 'getting frame' in error, got: %v", err)
	}
}

func TestStatusCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "status")
}
