package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

const testMessageEntry = `{"id":"m1","type":"nudge","attributes":{"body":"Dinner is ready!","deliver_at":"2027-10-30T14:00:00.000Z","rrule":null,"recurring":false,"voice_kind":"kirk_voice","audio_url":null},"relationships":{"categories":{"data":[{"id":"c1","type":"category"}]}}}`

// messageMockHandler serves the nudges endpoints and records the last
// request body sent to POST/PATCH in *gotBody.
func messageMockHandler(t *testing.T, gotBody *map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			b, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(b, gotBody); err != nil {
				t.Errorf("decoding request body: %v", err)
			}
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"label":"Dad"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/nudges/m1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/nudges") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[`+testMessageEntry+`]}`)
		default:
			fmt.Fprint(w, `{"data":`+testMessageEntry+`}`)
		}
	}
}

// noAPICallHandler fails the test if any request reaches the mock server.
func noAPICallHandler(t *testing.T) http.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("expected no API call, got %s %s", r.Method, r.URL.Path)
	}
}

// setMessageFlags sets flags on cmd and the backing vars, resetting both
// (including pflag's Changed state) when the test ends.
func setMessageFlags(t *testing.T, cmd *cobra.Command, flags map[string]string) {
	t.Helper()
	origID, origText, origTo, origAt, origRepeat := messageID, messageText, messageTo, messageSendAt, messageRepeat
	for name, val := range flags {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("setting --%s: %v", name, err)
		}
	}
	t.Cleanup(func() {
		messageID, messageText, messageTo, messageSendAt, messageRepeat = origID, origText, origTo, origAt, origRepeat
		for name := range flags {
			cmd.Flags().Lookup(name).Changed = false
		}
	})
}

func TestMessageCommandsRegistered(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "message")
	for _, use := range []string{subList, "send", subUpdate, subDelete} {
		assertCommandRegistered(t, messageCmd, use)
	}
}

func TestMessageListCmd(t *testing.T) {
	var body map[string]any
	newCmdTestClient(t, messageMockHandler(t, &body))

	out := captureStdout(func() {
		if err := messageListCmd.RunE(messageListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dinner is ready!") {
		t.Errorf("expected message body in output, got: %s", out)
	}
}

func TestMessageSendCmd(t *testing.T) {
	tests := []struct {
		name       string
		flags      map[string]string
		wantAt     string
		wantRRule  string
		wantErrMsg string
	}{
		{
			name:   "instant",
			flags:  map[string]string{"text": "Dinner is ready!", "to": "c1,c2"},
			wantAt: "now",
		},
		{
			name:      "scheduled recurring",
			flags:     map[string]string{"text": "Practice", "to": "c1", "send-at": "2027-10-30T09:00:00Z", "repeat": "FREQ=WEEKLY;BYDAY=SA"},
			wantAt:    "2027-10-30T09:00:00Z",
			wantRRule: "FREQ=WEEKLY;BYDAY=SA",
		},
		{
			name:       "repeat without send-at",
			flags:      map[string]string{"text": "x", "to": "c1", "repeat": "FREQ=DAILY"},
			wantErrMsg: "--repeat requires --send-at",
		},
		{
			name:       "empty recipients",
			flags:      map[string]string{"text": "x", "to": ","},
			wantErrMsg: "--to requires at least one category ID",
		},
		{
			name:       "bad send-at",
			flags:      map[string]string{"text": "x", "to": "c1", "send-at": "tomorrow"},
			wantErrMsg: "invalid --send-at",
		},
		{
			name:       "text too long",
			flags:      map[string]string{"text": strings.Repeat("a", maxMessageLen+1), "to": "c1"},
			wantErrMsg: "the limit is 220",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{}
			newCmdTestClient(t, messageMockHandler(t, &body))
			messageTo = nil // StringSlice flags append across Set calls
			setMessageFlags(t, messageSendCmd, tc.flags)

			var err error
			out := captureStdout(func() { err = messageSendCmd.RunE(messageSendCmd, nil) })
			if tc.wantErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrMsg) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErrMsg, err)
				}
				if len(body) != 0 {
					t.Errorf("expected no API call, got body %v", body)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(out, "m1") {
				t.Errorf("expected sent message in output, got: %s", out)
			}
			ids, _ := body["category_ids"].([]any)
			if want := len(strings.Split(tc.flags["to"], ",")); len(ids) != want {
				t.Errorf("category_ids = %v, want %d ids", body["category_ids"], want)
			}
			at, _ := body["deliver_at"].(string)
			if tc.wantAt == "now" {
				if _, err := time.Parse(time.RFC3339, at); err != nil {
					t.Errorf("deliver_at %q should default to now: %v", at, err)
				}
			} else if at != tc.wantAt {
				t.Errorf("deliver_at = %q, want %q", at, tc.wantAt)
			}
			if got, _ := body["rrule"].(string); got != tc.wantRRule {
				t.Errorf("rrule = %q, want %q", got, tc.wantRRule)
			}
		})
	}
}

func TestMessageUpdateCmd_SendsOnlyChangedFields(t *testing.T) {
	body := map[string]any{}
	newCmdTestClient(t, messageMockHandler(t, &body))
	messageTo = nil
	setMessageFlags(t, messageUpdateCmd, map[string]string{"message-id": "m1", "text": "New text"})

	out := captureStdout(func() {
		if err := messageUpdateCmd.RunE(messageUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "m1") {
		t.Errorf("expected updated message in output, got: %s", out)
	}
	if len(body) != 1 || body["body"] != "New text" {
		t.Errorf("expected only body in PATCH, got %v", body)
	}
}

func TestMessageUpdateCmd_ReplacesRecipientsAndSchedule(t *testing.T) {
	body := map[string]any{}
	newCmdTestClient(t, messageMockHandler(t, &body))
	messageTo = nil
	setMessageFlags(t, messageUpdateCmd, map[string]string{
		"message-id": "m1", "to": "c1,c2", "send-at": "2027-11-01T08:00:00Z", "repeat": "FREQ=DAILY",
	})

	captureStdout(func() {
		if err := messageUpdateCmd.RunE(messageUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if _, ok := body["body"]; ok {
		t.Errorf("body should not be sent when --text is unchanged: %v", body)
	}
	if ids, _ := body["category_ids"].([]any); len(ids) != 2 {
		t.Errorf("category_ids = %v", body["category_ids"])
	}
	if body["deliver_at"] != "2027-11-01T08:00:00Z" || body["rrule"] != "FREQ=DAILY" {
		t.Errorf("unexpected schedule fields: %v", body)
	}
}

func TestMessageUpdateCmd_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]string
		wantErr string
	}{
		{name: "empty text", flags: map[string]string{"message-id": "m1", "text": ""}, wantErr: "--text must not be empty"},
		{name: "empty recipients", flags: map[string]string{"message-id": "m1", "to": ","}, wantErr: "--to requires"},
		{name: "bad send-at", flags: map[string]string{"message-id": "m1", "send-at": "soon"}, wantErr: "invalid --send-at"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{}
			newCmdTestClient(t, messageMockHandler(t, &body))
			messageTo = nil
			setMessageFlags(t, messageUpdateCmd, tc.flags)
			err := messageUpdateCmd.RunE(messageUpdateCmd, nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got %v", tc.wantErr, err)
			}
			if len(body) != 0 {
				t.Errorf("expected no API call, got body %v", body)
			}
		})
	}
}

func TestMessageDeleteCmd(t *testing.T) {
	body := map[string]any{}
	newCmdTestClient(t, messageMockHandler(t, &body))
	origID, origYes := messageID, yes
	messageID, yes = "m1", true
	t.Cleanup(func() { messageID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := messageDeleteCmd.RunE(messageDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestMessageDeleteCmd_DryRun(t *testing.T) {
	newCmdTestClient(t, noAPICallHandler(t))
	origID, origDryRun := messageID, dryRun
	messageID, dryRun = "m1", true
	t.Cleanup(func() { messageID, dryRun = origID, origDryRun })

	out := captureStdout(func() {
		if err := messageDeleteCmd.RunE(messageDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run: would delete message m1") {
		t.Errorf("expected dry-run preview, got: %s", out)
	}
}

func TestMessageCmds_APIErrors(t *testing.T) {
	runAPIErrorCases(t, []apiErrorCase{
		{name: "list", cmd: func() error { return messageListCmd.RunE(messageListCmd, nil) }},
		{
			name: "send",
			setup: func(t *testing.T) {
				origText, origTo := messageText, messageTo
				messageText, messageTo = "x", []string{"c1"}
				t.Cleanup(func() { messageText, messageTo = origText, origTo })
			},
			cmd: func() error { return messageSendCmd.RunE(messageSendCmd, nil) },
		},
		{name: "update", cmd: func() error { return messageUpdateCmd.RunE(messageUpdateCmd, nil) }},
		{
			name: "delete",
			setup: func(t *testing.T) {
				origYes := yes
				yes = true
				t.Cleanup(func() { yes = origYes })
			},
			cmd: func() error { return messageDeleteCmd.RunE(messageDeleteCmd, nil) },
		},
	})
}

func TestMessageListWindow(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name, start, end      string
		wantAfter, wantBefore string
		wantErr               bool
	}{
		{name: "defaults to today plus 30 days", wantAfter: "2026-10-30", wantBefore: "2026-11-29"},
		{name: "end defaults relative to start", start: "2027-01-01", wantAfter: "2027-01-01", wantBefore: "2027-01-31"},
		{name: "explicit window", start: "2026-11-01", end: "2026-11-01", wantAfter: "2026-11-01", wantBefore: "2026-11-01"},
		{name: "end before start", start: "2026-12-01", end: "2026-11-01", wantErr: true},
		{name: "bad start", start: "12/01/2026", wantErr: true},
		{name: "bad end", end: "next week", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origStart, origEnd := messageStartDate, messageEndDate
			messageStartDate, messageEndDate = tc.start, tc.end
			t.Cleanup(func() { messageStartDate, messageEndDate = origStart, origEnd })

			after, before, err := messageListWindow(now)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if after != tc.wantAfter || before != tc.wantBefore {
				t.Errorf("got %s..%s, want %s..%s", after, before, tc.wantAfter, tc.wantBefore)
			}
		})
	}
}

func TestParseSendAt(t *testing.T) {
	origLocal := time.Local
	time.Local = time.FixedZone("CDT", -5*60*60)
	t.Cleanup(func() { time.Local = origLocal })

	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{in: "2026-11-07T08:00:00Z", want: "2026-11-07T08:00:00Z"},
		{in: "2026-11-07T08:00:00-06:00", want: "2026-11-07T08:00:00-06:00"},
		{in: "2026-11-07T08:00:00", want: "2026-11-07T08:00:00-05:00"},
		{in: "2026-11-07T08:00", want: "2026-11-07T08:00:00-05:00"},
		{in: "2026-11-07 08:00", want: "2026-11-07T08:00:00-05:00"},
		{in: "2026-11-07", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseSendAt(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateMessageText(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "ok", in: "Dinner!"},
		{name: "exactly at limit", in: strings.Repeat("a", maxMessageLen)},
		{name: "multibyte counted as runes", in: strings.Repeat("🍕", maxMessageLen)},
		{name: "over limit", in: strings.Repeat("a", maxMessageLen+1), wantErr: true},
		{name: "empty", in: "", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateMessageText(tc.in); (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestPrintMessagesTable(t *testing.T) {
	origLocal, origNames := time.Local, activeCatNames
	time.Local = time.UTC
	activeCatNames = map[string]string{"c1": "Dad", "c2": "Mom"}
	t.Cleanup(func() { time.Local, activeCatNames = origLocal, origNames })

	out := captureStdout(func() {
		printMessagesTable([]lib.Message{
			{ID: "m1", Body: "Weekly", DeliverAt: "2027-10-30T14:00:00.000Z", RRule: "FREQ=WEEKLY", CategoryIDs: []string{"c1", "c2"}},
			{ID: "m2", Body: "Once", DeliverAt: "not-a-time", CategoryIDs: []string{"c9"}},
		})
	})
	for _, want := range []string{"DELIVER AT", "2027-10-30 14:00", "Dad,Mom", "FREQ=WEEKLY", "not-a-time", "c9", "-"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}

func TestPrintOutput_MessagesTable(t *testing.T) {
	origFormat := outputFormat
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = origFormat })

	out := captureStdout(func() { printOutput([]lib.Message{{ID: "m1", Body: "Hi"}}) })
	if !strings.Contains(out, "DELIVER AT") {
		t.Errorf("expected messages table, got: %s", out)
	}
}

func TestMessageDeleteCmd_ConfirmationDeclined(t *testing.T) {
	newCmdTestClient(t, noAPICallHandler(t))
	origID, origYes := messageID, yes
	messageID, yes = "m1", false
	t.Cleanup(func() { messageID, yes = origID, origYes })
	mockStdin(t, "n\n")

	out := captureStdout(func() {
		if err := messageDeleteCmd.RunE(messageDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "deleted successfully") {
		t.Errorf("expected no deletion when confirmation declined, got: %s", out)
	}
}

func TestMessageCmds_MissingFrameID(t *testing.T) {
	origFrameID := frameID
	frameID = ""
	t.Cleanup(func() { frameID = origFrameID })

	for _, c := range []*cobra.Command{messageListCmd, messageSendCmd, messageUpdateCmd, messageDeleteCmd} {
		t.Run(c.Use, func(t *testing.T) {
			if err := c.RunE(c, nil); err == nil {
				t.Error("expected error when --frame-id is missing")
			}
		})
	}
}

func TestMessageListCmd_InvalidWindow(t *testing.T) {
	newCmdTestClient(t, noAPICallHandler(t))
	origStart := messageStartDate
	messageStartDate = "not-a-date"
	t.Cleanup(func() { messageStartDate = origStart })

	if err := messageListCmd.RunE(messageListCmd, nil); err == nil || !strings.Contains(err.Error(), "invalid date") {
		t.Errorf("expected invalid date error, got %v", err)
	}
}
