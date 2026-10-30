package lib

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testNudgeJSON = `{"id":"76304","type":"nudge","attributes":{"id":76304,"body":"Dinner!","deliver_at":"2027-10-30T14:00:00.000Z","rrule":"FREQ=WEEKLY;BYDAY=SA","recurring_until":null,"voice_kind":"kirk_voice","recurring":true,"audio_url":"https://s3.example/nudge.mp3"},"relationships":{"categories":{"data":[{"id":"c1","type":"category"},{"id":"c2","type":"category"}]}}}`

// newMessageTestClient starts a mock server and returns a client pointed at it.
func newMessageTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"))
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	return client
}

// decodeBody reads a JSON request body into a generic map.
func decodeBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("decoding body %q: %v", b, err)
	}
	return m
}

func TestListMessages(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantErr  bool
	}{
		{name: "flattens nudges", status: http.StatusOK, response: `{"data":[` + testNudgeJSON + `]}`, wantLen: 1},
		{name: "empty list", status: http.StatusOK, response: `{"data":[]}`, wantLen: 0},
		{name: "missing window returns error", status: http.StatusUnprocessableEntity, response: `{"errors":["After is required."]}`, wantErr: true},
		{name: "invalid JSON returns error", status: http.StatusOK, response: `nope`, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newMessageTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/nudges" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if got := r.URL.Query().Get("after"); got != "2027-10-01" {
					t.Errorf("after = %q", got)
				}
				if got := r.URL.Query().Get("before"); got != "2027-11-30" {
					t.Errorf("before = %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.response))
			})

			msgs, err := client.ListMessages(context.Background(), "frame1", "2027-10-01", "2027-11-30")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if len(msgs) != tc.wantLen {
				t.Fatalf("wantLen=%d got %d", tc.wantLen, len(msgs))
			}
			if tc.wantLen == 0 {
				return
			}
			m := msgs[0]
			if m.ID != "76304" || m.Body != "Dinner!" || m.DeliverAt != "2027-10-30T14:00:00.000Z" {
				t.Errorf("unexpected message: %+v", m)
			}
			if m.RRule != "FREQ=WEEKLY;BYDAY=SA" || !m.Recurring || m.VoiceKind != "kirk_voice" || m.AudioURL == "" {
				t.Errorf("unexpected recurrence/voice fields: %+v", m)
			}
			if len(m.CategoryIDs) != 2 || m.CategoryIDs[0] != "c1" || m.CategoryIDs[1] != "c2" {
				t.Errorf("CategoryIDs = %v", m.CategoryIDs)
			}
		})
	}
}

func TestMessageToMessage_NullFields(t *testing.T) {
	var e messageAPIEntry
	if err := json.Unmarshal([]byte(`{"id":"1","attributes":{"body":"hi","rrule":null,"audio_url":null}}`), &e); err != nil {
		t.Fatal(err)
	}
	m := e.toMessage()
	if m.RRule != "" || m.AudioURL != "" || m.Recurring {
		t.Errorf("null fields should stay empty: %+v", m)
	}
	if m.CategoryIDs == nil || len(m.CategoryIDs) != 0 {
		t.Errorf("CategoryIDs should be empty non-nil, got %#v", m.CategoryIDs)
	}
}

func TestSendMessage(t *testing.T) {
	tests := []struct {
		name      string
		data      MessageData
		wantAt    string // exact deliver_at expected; empty means "about now"
		wantRRule bool
	}{
		{
			name:      "scheduled keeps deliver_at and rrule",
			data:      MessageData{Body: "Practice", CategoryIDs: []string{"c1"}, DeliverAt: "2027-10-30T09:00:00-05:00", RRule: "FREQ=WEEKLY"},
			wantAt:    "2027-10-30T09:00:00-05:00",
			wantRRule: true,
		},
		{
			name: "instant fills deliver_at with now",
			data: MessageData{Body: "Dinner!", CategoryIDs: []string{"c1", "c2"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var body map[string]any
			client := newMessageTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/api/frames/frame1/nudges" {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				}
				body = decodeBody(t, r)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":` + testNudgeJSON + `}`))
			})

			before := time.Now().Add(-time.Second)
			msg, err := client.SendMessage(context.Background(), "frame1", tc.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg.ID != "76304" {
				t.Errorf("ID = %q", msg.ID)
			}
			if body["body"] != tc.data.Body {
				t.Errorf("body = %v", body["body"])
			}
			ids, _ := body["category_ids"].([]any)
			if len(ids) != len(tc.data.CategoryIDs) {
				t.Errorf("category_ids = %v", body["category_ids"])
			}
			_, hasRRule := body["rrule"]
			if hasRRule != tc.wantRRule {
				t.Errorf("rrule present = %v, want %v", hasRRule, tc.wantRRule)
			}
			at, _ := body["deliver_at"].(string)
			if tc.wantAt != "" {
				if at != tc.wantAt {
					t.Errorf("deliver_at = %q, want %q", at, tc.wantAt)
				}
				return
			}
			parsed, err := time.Parse(time.RFC3339, at)
			if err != nil {
				t.Fatalf("deliver_at %q is not RFC 3339: %v", at, err)
			}
			if parsed.Before(before) || parsed.After(time.Now().Add(time.Second)) {
				t.Errorf("deliver_at %s is not about now", at)
			}
		})
	}
}

func TestSendMessage_ValidationError(t *testing.T) {
	client := newMessageTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"errors":["Category ids is required."]}`))
	})
	_, err := client.SendMessage(context.Background(), "frame1", MessageData{Body: "x"})
	if !IsValidation(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestUpdateMessage_SendsOnlySetFields(t *testing.T) {
	var body map[string]any
	client := newMessageTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/frames/frame1/nudges/76304" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":` + testNudgeJSON + `}`))
	})

	msg, err := client.UpdateMessage(context.Background(), "frame1", "76304", MessageData{Body: "Dinner!"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Body != "Dinner!" {
		t.Errorf("Body = %q", msg.Body)
	}
	if len(body) != 1 || body["body"] != "Dinner!" {
		t.Errorf("expected only body in PATCH, got %v", body)
	}
}

func TestUpdateMessage_NotFound(t *testing.T) {
	client := newMessageTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	if _, err := client.UpdateMessage(context.Background(), "frame1", "missing", MessageData{Body: "x"}); err == nil {
		t.Error("expected error for 404")
	}
}

func TestDeleteMessage(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{name: "200 with empty body", status: http.StatusOK},
		{name: "404 returns error", status: http.StatusNotFound, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newMessageTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete || r.URL.Path != "/api/frames/frame1/nudges/76304" {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte("\n"))
			})
			err := client.DeleteMessage(context.Background(), "frame1", "76304")
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}
