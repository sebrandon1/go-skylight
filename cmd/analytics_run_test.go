package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func analyticsMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[{"id":"1","attributes":{"label":"Mom","color":"#FF0000"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"complete"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"Ice cream","point_value":10}}]}`)
		case strings.HasSuffix(r.URL.Path, "/reward_points"):
			fmt.Fprint(w, `[{"category_id":1,"current_point_balance":5}]`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	}
}

func TestAnalyticsCmd_JSON(t *testing.T) {
	newCmdTestClient(t, analyticsMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	out := captureStdout(func() { analyticsCmd.Run(analyticsCmd, nil) })
	if !strings.Contains(out, `"period_days"`) {
		t.Errorf("expected analytics stats in JSON output, got: %s", out)
	}
}

func TestAnalyticsCmd_Text(t *testing.T) {
	newCmdTestClient(t, analyticsMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	out := captureStdout(func() { analyticsCmd.Run(analyticsCmd, nil) })
	if !strings.Contains(out, "Analytics:") {
		t.Errorf("expected analytics text report, got: %s", out)
	}
}

func TestAnalyticsCmd_CalendarErrorIsNonFatal(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			w.WriteHeader(http.StatusInternalServerError)
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/reward_points"):
			fmt.Fprint(w, `[]`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	})
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	out := captureStdout(func() { analyticsCmd.Run(analyticsCmd, nil) })
	if !strings.Contains(out, "Analytics:") {
		t.Errorf("expected analytics report despite calendar error, got: %s", out)
	}
}

func TestAnalyticsCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "analytics" {
			found = true
			break
		}
	}
	if !found {
		t.Error("analytics command not registered on root")
	}
}
