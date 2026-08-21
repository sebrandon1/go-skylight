package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func calendarMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/source_calendars") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[{"id":"sc1","attributes":{"label":"Family","kind":"google"}}]}`)
		case strings.Contains(r.URL.Path, "/source_calendars/") && r.Method == http.MethodPatch:
			w.WriteHeader(http.StatusOK)
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

func TestCalendarCreateCmd_WithColorAndCategory(t *testing.T) {
	var gotColor, gotCatID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/calendar_events") && r.Method == http.MethodPost:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				gotColor, _ = body["color"].(string)
				gotCatID, _ = body["category_id"].(string)
			}
			fmt.Fprint(w, `{"data":{"id":"e1","type":"calendar_event","attributes":{"summary":"Dentist","color":"red","all_day":false},"relationships":{"categories":{"data":[{"id":"cat1","type":"category"}]}}}}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	})
	newCmdTestClient(t, handler)
	orig := struct{ title, start, color, cat string }{calendarTitle, calendarStartAt, calendarColor, calendarCategoryID}
	calendarTitle, calendarStartAt, calendarColor, calendarCategoryID = "Dentist", "2026-08-10T09:00:00Z", "red", "cat1"
	t.Cleanup(func() {
		calendarTitle, calendarStartAt, calendarColor, calendarCategoryID = orig.title, orig.start, orig.color, orig.cat
	})

	captureStdout(func() {
		if err := calendarCreateCmd.RunE(calendarCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if gotColor != "red" {
		t.Errorf("color in request body: want %q got %q", "red", gotColor)
	}
	if gotCatID != "cat1" {
		t.Errorf("category_id in request body: want %q got %q", "cat1", gotCatID)
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

func TestCalendarUpdateCmd_WithColorAndCategory(t *testing.T) {
	var gotColor, gotCatID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/event1") && r.Method == http.MethodPut:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				gotColor, _ = body["color"].(string)
				gotCatID, _ = body["category_id"].(string)
			}
			fmt.Fprint(w, `{"data":{"id":"event1","type":"calendar_event","attributes":{"summary":"Updated","all_day":false},"relationships":{"categories":{"data":[]}}}}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	})
	newCmdTestClient(t, handler)

	origID, origColor, origCat := calendarEventID, calendarColor, calendarCategoryID
	calendarEventID, calendarColor, calendarCategoryID = "event1", "blue", "cat2"
	t.Cleanup(func() { calendarEventID, calendarColor, calendarCategoryID = origID, origColor, origCat })

	// pflag.Set() marks flags as permanently Changed on the shared singleton.
	if err := calendarUpdateCmd.Flags().Set("color", "blue"); err != nil {
		t.Fatalf("setting color flag: %v", err)
	}
	if err := calendarUpdateCmd.Flags().Set("category-id", "cat2"); err != nil {
		t.Fatalf("setting category-id flag: %v", err)
	}

	captureStdout(func() {
		if err := calendarUpdateCmd.RunE(calendarUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if gotColor != "blue" {
		t.Errorf("color in PUT body: want %q got %q", "blue", gotColor)
	}
	if gotCatID != "cat2" {
		t.Errorf("category_id in PUT body: want %q got %q", "cat2", gotCatID)
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

func TestCalendarSourceEnableCmd(t *testing.T) {
	var capturedPath string
	var capturedBody map[string]any

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/source_calendars/") && r.Method == http.MethodPatch {
			capturedPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&capturedBody)
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
	})
	newCmdTestClient(t, handler)

	orig := calendarSourceID
	calendarSourceID = "sc1"
	t.Cleanup(func() { calendarSourceID = orig })

	if err := calendarSourceEnableCmd.RunE(calendarSourceEnableCmd, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(capturedPath, "/source_calendars/sc1") {
		t.Errorf("path: want suffix /source_calendars/sc1, got %s", capturedPath)
	}
	if sc, ok := capturedBody["source_calendar"].(map[string]any); !ok || sc["enabled"] != true {
		t.Errorf("body: want {source_calendar:{enabled:true}}, got %v", capturedBody)
	}
}

func TestCalendarSourceDisableCmd(t *testing.T) {
	var capturedBody map[string]any

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/source_calendars/") && r.Method == http.MethodPatch {
			_ = json.NewDecoder(r.Body).Decode(&capturedBody)
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
	})
	newCmdTestClient(t, handler)

	orig := calendarSourceID
	calendarSourceID = "sc1"
	t.Cleanup(func() { calendarSourceID = orig })

	if err := calendarSourceDisableCmd.RunE(calendarSourceDisableCmd, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sc, ok := capturedBody["source_calendar"].(map[string]any); !ok || sc["enabled"] != false {
		t.Errorf("body: want {source_calendar:{enabled:false}}, got %v", capturedBody)
	}
}

func TestCalendarDayCmd(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())

	if err := calendarDayCmd.Flags().Set(subDate, "2026-01-01"); err != nil {
		t.Fatalf("setting date flag: %v", err)
	}
	orig := calendarDayDate
	calendarDayDate = "2026-01-01"
	t.Cleanup(func() { calendarDayDate = orig })

	out := captureStdout(func() {
		if err := calendarDayCmd.RunE(calendarDayCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Meeting") {
		t.Errorf("expected event in output, got: %s", out)
	}
}

func TestCalendarDayCmd_DefaultDate(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())

	out := captureStdout(func() {
		if err := calendarDayCmd.RunE(calendarDayCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	_ = out // today may return empty events; just verify no error
}

func TestCalendarDayCmd_InvalidDate(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())

	if err := calendarDayCmd.Flags().Set(subDate, "not-a-date"); err != nil {
		t.Fatalf("setting date flag: %v", err)
	}
	orig := calendarDayDate
	calendarDayDate = "not-a-date"
	t.Cleanup(func() { calendarDayDate = orig })

	if err := calendarDayCmd.RunE(calendarDayCmd, nil); err == nil {
		t.Error("expected error for invalid date, got nil")
	}
}

func TestCalendarCreateCmd_TableOutput(t *testing.T) {
	newCmdTestClient(t, calendarMockHandler())
	origTitle, origStart, origEnd, origFmt := calendarTitle, calendarStartAt, calendarEndAt, outputFormat
	calendarTitle, calendarStartAt, calendarEndAt, outputFormat = "New Event", "2026-01-01T09:00:00Z", "2026-01-01T10:00:00Z", outputTable
	t.Cleanup(func() {
		calendarTitle, calendarStartAt, calendarEndAt, outputFormat = origTitle, origStart, origEnd, origFmt
	})

	out := captureStdout(func() {
		if err := calendarCreateCmd.RunE(calendarCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "TITLE") {
		t.Errorf("expected table header in output, got: %s", out)
	}
	if !strings.Contains(out, "New Event") {
		t.Errorf("expected event title in table, got: %s", out)
	}
}
