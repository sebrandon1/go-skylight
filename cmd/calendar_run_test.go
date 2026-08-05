package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func calendarMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/source_calendars"):
			fmt.Fprint(w, `{"data":[{"id":"sc1","attributes":{"label":"Family","kind":"google"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"e1","type":"calendar_event","attributes":{"summary":"New Event","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}}`)
		case strings.HasSuffix(r.URL.Path, "/event1") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":{"id":"event1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}}`)
		case strings.HasSuffix(r.URL.Path, "/event1") && r.Method == http.MethodPut:
			fmt.Fprint(w, `{"data":{"id":"event1","type":"calendar_event","attributes":{"summary":"Updated","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}}`)
		case strings.HasSuffix(r.URL.Path, "/event1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	}
}

func TestCalendarGetCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())
	origID := calendarEventID
	calendarEventID = "event1"
	t.Cleanup(func() { calendarEventID = origID })

	out := captureStdout(func() {
		if err := calendarGetCmd.RunE(calendarGetCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Meeting") {
		t.Errorf("expected event in output, got: %s", out)
	}
}

func TestCalendarGetCmd_APIError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	origID := calendarEventID
	calendarEventID = "event1"
	t.Cleanup(func() { calendarEventID = origID })

	if err := calendarGetCmd.RunE(calendarGetCmd, nil); err == nil {
		t.Error("expected error when API returns 404, got nil")
	}
}

func TestCalendarListCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())
	origStart, origEnd := calendarStartDate, calendarEndDate
	calendarStartDate, calendarEndDate = "2026-01-01", "2026-01-07"
	t.Cleanup(func() { calendarStartDate, calendarEndDate = origStart, origEnd })

	out := captureStdout(func() {
		if err := calendarListCmd.RunE(calendarListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Meeting") {
		t.Errorf("expected event in output, got: %s", out)
	}
}

func TestCalendarCreateCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())
	origTitle, origStart := calendarTitle, calendarStartAt
	calendarTitle, calendarStartAt = "New Event", "2026-01-01T10:00:00Z"
	t.Cleanup(func() { calendarTitle, calendarStartAt = origTitle, origStart })

	out := captureStdout(func() {
		if err := calendarCreateCmd.RunE(calendarCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "New Event") {
		t.Errorf("expected created event in output, got: %s", out)
	}
}

func TestCalendarDeleteCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())
	origID, origYes := calendarEventID, yes
	calendarEventID, yes = "event1", true
	t.Cleanup(func() { calendarEventID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := calendarDeleteCmd.RunE(calendarDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestCalendarDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := calendarEventID, dryRun
	calendarEventID, dryRun = "event1", true
	t.Cleanup(func() { calendarEventID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := calendarDeleteCmd.RunE(calendarDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestCalendarDeleteCmd_ConfirmationDeclined(t *testing.T) {
	origID, origYes := calendarEventID, yes
	calendarEventID, yes = "event1", false
	t.Cleanup(func() { calendarEventID, yes = origID, origYes })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	mockStdin(t, "n\n")

	out := captureStdout(func() {
		if err := calendarDeleteCmd.RunE(calendarDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "deleted successfully") {
		t.Errorf("expected no deletion when confirmation declined, got: %s", out)
	}
}

func TestSourceCalendarsCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())

	out := captureStdout(func() {
		if err := sourceCalendarsCmd.RunE(sourceCalendarsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Family") {
		t.Errorf("expected source calendar in output, got: %s", out)
	}
}

func TestCalendarUpdateCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())
	origID, origTitle := calendarEventID, calendarTitle
	calendarEventID, calendarTitle = "event1", "Updated"
	t.Cleanup(func() { calendarEventID, calendarTitle = origID, origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := calendarUpdateCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() {
		if err := calendarUpdateCmd.RunE(calendarUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated event in output, got: %s", out)
	}
}

func TestCalendarCreateCountdownCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())
	origTitle, origDate := calendarTitle, calendarCountdownDate
	calendarTitle, calendarCountdownDate = "Birthday", "2026-06-01"
	t.Cleanup(func() { calendarTitle, calendarCountdownDate = origTitle, origDate })

	out := captureStdout(func() {
		if err := calendarCreateCountdownCmd.RunE(calendarCreateCountdownCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "New Event") {
		t.Errorf("expected created countdown event in output, got: %s", out)
	}
}

func TestCalendarWeekCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())

	out := captureStdout(func() {
		if err := calendarWeekCmd.RunE(calendarWeekCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if out == "" {
		t.Error("expected non-empty weekly view output")
	}
}

func TestCalendarCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "calendar")
}
