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

	out := captureStdout(func() {
		if err := analyticsCmd.RunE(analyticsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, `"period_days"`) {
		t.Errorf("expected analytics stats in JSON output, got: %s", out)
	}
}

func TestAnalyticsCmd_Text(t *testing.T) {
	newCmdTestClient(t, analyticsMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	out := captureStdout(func() {
		if err := analyticsCmd.RunE(analyticsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
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

	out := captureStdout(func() {
		if err := analyticsCmd.RunE(analyticsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Analytics:") {
		t.Errorf("expected analytics report despite calendar error, got: %s", out)
	}
}

func TestAnalyticsCmd_ChoresFetchError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/chores"):
			w.WriteHeader(http.StatusInternalServerError)
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/reward_points"):
			fmt.Fprint(w, `[]`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	})

	err := analyticsCmd.RunE(analyticsCmd, nil)
	if err == nil {
		t.Fatal("expected error when chores API returns 500, got nil")
	}
	if !strings.Contains(err.Error(), "listing chores") {
		t.Errorf("expected 'listing chores' in error, got: %v", err)
	}
}

func TestAnalyticsCmd_StartDateEndDate(t *testing.T) {
	newCmdTestClient(t, analyticsMockHandler())
	orig := struct{ start, end, fmt string }{analyticsStartDate, analyticsEndDate, outputFormat}
	analyticsStartDate, analyticsEndDate, outputFormat = "2026-07-01", "2026-07-31", outputJSON
	t.Cleanup(func() {
		analyticsStartDate, analyticsEndDate, outputFormat = orig.start, orig.end, orig.fmt
	})

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := analyticsCmd.Flags().Set("start-date", "2026-07-01"); err != nil {
		t.Fatalf("setting start-date flag: %v", err)
	}

	out := captureStdout(func() {
		if err := analyticsCmd.RunE(analyticsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, `"start_date": "2026-07-01"`) {
		t.Errorf("expected start_date in output, got: %s", out)
	}
	if !strings.Contains(out, `"end_date": "2026-07-31"`) {
		t.Errorf("expected end_date in output, got: %s", out)
	}
}

func TestAnalyticsCmd_EndDateOnly(t *testing.T) {
	newCmdTestClient(t, analyticsMockHandler())
	orig := struct{ start, end, fmt string }{analyticsStartDate, analyticsEndDate, outputFormat}
	analyticsStartDate, analyticsEndDate, outputFormat = "", "2026-07-31", outputJSON
	t.Cleanup(func() {
		analyticsStartDate, analyticsEndDate, outputFormat = orig.start, orig.end, orig.fmt
	})

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := analyticsCmd.Flags().Set("end-date", "2026-07-31"); err != nil {
		t.Fatalf("setting end-date flag: %v", err)
	}

	out := captureStdout(func() {
		if err := analyticsCmd.RunE(analyticsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, `"end_date": "2026-07-31"`) {
		t.Errorf("expected end_date in output, got: %s", out)
	}
}

func TestAnalyticsCmd_InvalidStartDate(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	orig := struct{ start, end string }{analyticsStartDate, analyticsEndDate}
	analyticsStartDate, analyticsEndDate = "not-a-date", ""
	t.Cleanup(func() { analyticsStartDate, analyticsEndDate = orig.start, orig.end })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := analyticsCmd.Flags().Set("start-date", "not-a-date"); err != nil {
		t.Fatalf("setting start-date flag: %v", err)
	}

	err := analyticsCmd.RunE(analyticsCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid --start-date, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --start-date") {
		t.Errorf("expected 'invalid --start-date' in error, got: %v", err)
	}
}

func TestAnalyticsCmd_InvalidEndDate(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	orig := struct{ start, end string }{analyticsStartDate, analyticsEndDate}
	analyticsStartDate, analyticsEndDate = "", "not-a-date"
	t.Cleanup(func() { analyticsStartDate, analyticsEndDate = orig.start, orig.end })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := analyticsCmd.Flags().Set("end-date", "not-a-date"); err != nil {
		t.Fatalf("setting end-date flag: %v", err)
	}

	err := analyticsCmd.RunE(analyticsCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid --end-date, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --end-date") {
		t.Errorf("expected 'invalid --end-date' in error, got: %v", err)
	}
}

func TestAnalyticsCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "analytics")
}
