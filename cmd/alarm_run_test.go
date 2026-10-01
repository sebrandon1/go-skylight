package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
)

const alarmRunEntryJSON = `{"id":"a1","attributes":{"label":"School","enabled":true,"time":"06:30","hour":6,"minute":30,"rrule":"FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO,TU","sound":"chimes","volume":60,"snoozable":true}}`

type alarmMock struct {
	mu       sync.Mutex
	requests []string
	bodies   []map[string]any
	devices  string
	// emptyWrites makes POST/PATCH succeed with an empty body.
	emptyWrites bool
}

func (m *alarmMock) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		m.requests = append(m.requests, r.Method+" "+r.URL.Path)
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			var body map[string]any
			_ = json.Unmarshal(raw, &body)
			m.bodies = append(m.bodies, body)
		}
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if m.emptyWrites && (r.Method == http.MethodPost || r.Method == http.MethodPatch) {
			w.WriteHeader(http.StatusOK)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/devices"):
			fmt.Fprint(w, m.devices)
		case strings.HasSuffix(r.URL.Path, "/alarms") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[`+alarmRunEntryJSON+`]}`)
		case strings.HasSuffix(r.URL.Path, "/alarms") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":`+alarmRunEntryJSON+`}`)
		case strings.HasSuffix(r.URL.Path, "/alarms/a1") && r.Method == http.MethodPatch:
			fmt.Fprint(w, `{"data":`+alarmRunEntryJSON+`}`)
		case strings.HasSuffix(r.URL.Path, "/alarms/a1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func newAlarmMock(t *testing.T) *alarmMock {
	t.Helper()
	m := &alarmMock{devices: `{"data":[{"id":"frame1","attributes":{"role":"frame"}},{"id":"buddy1","attributes":{"role":"buddy"}}]}`}
	newCmdTestClient(t, m.handler())
	return m
}

func setAlarmGlobals(t *testing.T, deviceID, id string) {
	t.Helper()
	origDevice, origID, origYes, origDryRun, origOutput := alarmDeviceID, alarmID, yes, dryRun, outputFormat
	alarmDeviceID, alarmID, yes, dryRun = deviceID, id, true, false
	t.Cleanup(func() {
		alarmDeviceID, alarmID, yes, dryRun, outputFormat = origDevice, origID, origYes, origDryRun, origOutput
	})
}

// alarmFieldCmd returns a fresh command carrying the alarm field flags so tests
// can mark flags as changed without mutating the shared command singletons.
func alarmFieldCmd(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	c := &cobra.Command{Use: "test"}
	addAlarmFieldFlags(c)
	c.SetContext(context.Background())
	if err := c.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	return c
}

func TestAlarmListCmd_AutoDetectsBuddy(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "", "")

	out := captureStdout(func() {
		if err := alarmListCmd.RunE(alarmListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "School") {
		t.Errorf("expected alarm in output, got: %s", out)
	}
	want := []string{"GET /frames/test-frame/devices", "GET /frames/test-frame/devices/buddy1/alarms"}
	if strings.Join(m.requests, "|") != strings.Join(want, "|") {
		t.Errorf("requests = %v, want %v", m.requests, want)
	}
}

func TestAlarmListCmd_ExplicitDeviceSkipsLookup(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "dev9", "")

	captureStdout(func() {
		if err := alarmListCmd.RunE(alarmListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if len(m.requests) != 1 || m.requests[0] != "GET /frames/test-frame/devices/dev9/alarms" {
		t.Errorf("requests = %v, want only the dev9 alarms GET", m.requests)
	}
}

func TestAlarmListCmd_NoBuddy(t *testing.T) {
	m := newAlarmMock(t)
	m.devices = `{"data":[{"id":"frame1","attributes":{"role":"frame"}}]}`
	setAlarmGlobals(t, "", "")

	err := alarmListCmd.RunE(alarmListCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "no Buddy device") {
		t.Errorf("want no-Buddy error, got %v", err)
	}
}

func TestAlarmListCmd_TableOutput(t *testing.T) {
	newAlarmMock(t)
	setAlarmGlobals(t, "buddy1", "")
	outputFormat = outputTable

	out := captureStdout(func() {
		if err := alarmListCmd.RunE(alarmListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	for _, want := range []string{"LABEL", "School", "06:30", "MO,TU", "chimes", "60"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}

func TestAlarmCreateCmd_DefaultsEnabled(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "buddy1", "")

	c := alarmFieldCmd(t, "--time", "06:30", "--label", "School")
	captureStdout(func() {
		if err := alarmCreateCmd.RunE(c, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if len(m.bodies) != 1 {
		t.Fatalf("want 1 request body, got %d", len(m.bodies))
	}
	body := m.bodies[0]
	if body["enabled"] != true || body["time"] != "06:30" || body["label"] != "School" {
		t.Errorf("unexpected create body: %v", body)
	}
}

func TestAlarmCreateCmd_UnrecognizedResponse(t *testing.T) {
	m := newAlarmMock(t)
	m.emptyWrites = true
	setAlarmGlobals(t, "buddy1", "")

	out := captureStdout(func() {
		if err := alarmCreateCmd.RunE(alarmFieldCmd(t, "--time", "06:30"), nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Alarm created") || !strings.Contains(out, "alarm list") {
		t.Errorf("expected created message pointing at alarm list, got: %s", out)
	}
}

func TestAlarmUpdateCmd_UnrecognizedResponseReReads(t *testing.T) {
	m := newAlarmMock(t)
	m.emptyWrites = true
	setAlarmGlobals(t, "buddy1", "a1")

	out := captureStdout(func() {
		if err := alarmUpdateCmd.RunE(alarmFieldCmd(t, "--label", "x"), nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "School") {
		t.Errorf("expected alarm re-read from list, got: %s", out)
	}
}

func TestAlarmCreateCmd_RespectsEnabledFalse(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "buddy1", "")

	c := alarmFieldCmd(t, "--time", "06:30", "--enabled=false")
	captureStdout(func() {
		if err := alarmCreateCmd.RunE(c, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if len(m.bodies) != 1 || m.bodies[0]["enabled"] != false {
		t.Errorf("want enabled=false in body, got %v", m.bodies)
	}
}

func TestAlarmCreateCmd_InvalidFlagsMakeNoRequests(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "buddy1", "")

	c := alarmFieldCmd(t, "--time", "99:99")
	if err := alarmCreateCmd.RunE(c, nil); err == nil {
		t.Fatal("expected invalid time error")
	}
	if len(m.requests) != 0 {
		t.Errorf("want no API requests, got %v", m.requests)
	}
}

func TestAlarmUpdateCmd(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "buddy1", "a1")

	c := alarmFieldCmd(t, "--volume", "25")
	out := captureStdout(func() {
		if err := alarmUpdateCmd.RunE(c, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "School") {
		t.Errorf("expected updated alarm in output, got: %s", out)
	}
	if len(m.bodies) != 1 || len(m.bodies[0]) != 1 || m.bodies[0]["volume"] != float64(25) {
		t.Errorf("want body {volume:25} only, got %v", m.bodies)
	}
	if m.requests[0] != "PATCH /frames/test-frame/devices/buddy1/alarms/a1" {
		t.Errorf("unexpected request %v", m.requests)
	}
}

func TestAlarmUpdateCmd_NothingToUpdate(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "buddy1", "a1")

	err := alarmUpdateCmd.RunE(alarmFieldCmd(t), nil)
	if err == nil || !strings.Contains(err.Error(), "nothing to update") {
		t.Errorf("want nothing-to-update error, got %v", err)
	}
	if len(m.requests) != 0 {
		t.Errorf("want no API requests, got %v", m.requests)
	}
}

func TestAlarmDeleteCmd(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "buddy1", "a1")

	out := captureStdout(func() {
		if err := alarmDeleteCmd.RunE(alarmDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
	if len(m.requests) != 1 || m.requests[0] != "DELETE /frames/test-frame/devices/buddy1/alarms/a1" {
		t.Errorf("unexpected requests %v", m.requests)
	}
}

func TestAlarmDeleteCmd_DryRun(t *testing.T) {
	m := newAlarmMock(t)
	setAlarmGlobals(t, "", "a1")
	dryRun = true

	out := captureStdout(func() {
		if err := alarmDeleteCmd.RunE(alarmDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
	if len(m.requests) != 0 {
		t.Errorf("dry run must make no API requests, got %v", m.requests)
	}
}

func TestAlarmCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "alarm")
}

func TestAlarmCmd_APIErrors(t *testing.T) {
	explicitDevice := func(t *testing.T) { setAlarmGlobals(t, "buddy1", "a1") }
	runAPIErrorCases(t, []apiErrorCase{
		{
			name:  "list device lookup",
			setup: func(t *testing.T) { setAlarmGlobals(t, "", "") },
			cmd:   func() error { return alarmListCmd.RunE(alarmListCmd, nil) },
		},
		{
			name:  "list",
			setup: explicitDevice,
			cmd:   func() error { return alarmListCmd.RunE(alarmListCmd, nil) },
		},
		{
			name:  "create",
			setup: explicitDevice,
			cmd: func() error {
				return alarmCreateCmd.RunE(alarmFieldCmd(t, "--time", "06:30"), nil)
			},
		},
		{
			name:  "update",
			setup: explicitDevice,
			cmd: func() error {
				return alarmUpdateCmd.RunE(alarmFieldCmd(t, "--label", "x"), nil)
			},
		},
		{
			name:  "delete",
			setup: explicitDevice,
			cmd:   func() error { return alarmDeleteCmd.RunE(alarmDeleteCmd, nil) },
		},
	})
}
