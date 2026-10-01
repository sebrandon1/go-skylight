package lib

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const alarmEntryJSON = `{"id":"a1","type":"alarm","attributes":{"label":"School","enabled":true,"time":"06:30","hour":6,"minute":30,"rrule":"FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO,TU","fires_on":null,"sound":"chimes","volume":60,"snoozable":true}}`

type capturedRequest struct {
	method string
	path   string
	rawURI string
	body   map[string]any
}

func newAlarmTestClient(t *testing.T, status int, response string) (*Client, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.rawURI = r.RequestURI
		raw, _ := io.ReadAll(r.Body)
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &captured.body); err != nil {
				t.Errorf("request body is not a JSON object: %v", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if response != "" {
			if _, err := w.Write([]byte(response)); err != nil {
				t.Errorf("write response: %v", err)
			}
		}
	}))
	t.Cleanup(srv.Close)

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	t.Cleanup(func() { SkylightURL = old })

	client, err := NewClientWithToken("u", "t")
	if err != nil {
		t.Fatalf("NewClientWithToken: %v", err)
	}
	return client, captured
}

func TestListAlarms(t *testing.T) {
	client, req := newAlarmTestClient(t, http.StatusOK, `{"data":[`+alarmEntryJSON+`]}`)

	alarms, err := client.ListAlarms(context.Background(), "f1", "d1")
	if err != nil {
		t.Fatalf("ListAlarms: %v", err)
	}
	if req.method != http.MethodGet || req.path != "/api/frames/f1/devices/d1/alarms" {
		t.Errorf("unexpected request %s %s", req.method, req.path)
	}
	if len(alarms) != 1 {
		t.Fatalf("want 1 alarm, got %d", len(alarms))
	}
	a := alarms[0]
	if a.ID != "a1" || a.Label != "School" || !a.Enabled || a.Time != "06:30" || a.Hour != 6 || a.Minute != 30 || a.Volume != 60 || !a.Snoozable {
		t.Errorf("unexpected alarm: %+v", a)
	}
	if string(a.Sound) != `"chimes"` {
		t.Errorf("sound: want raw \"chimes\", got %s", a.Sound)
	}
	if a.FiresOn != nil {
		t.Errorf("fires_on: null should decode to nil, got %s", a.FiresOn)
	}
}

func TestListAlarmsError(t *testing.T) {
	client, _ := newAlarmTestClient(t, http.StatusInternalServerError, "")
	if _, err := client.ListAlarms(context.Background(), "f1", "d1"); err == nil {
		t.Fatal("expected error on 500")
	}
}

func TestCreateAlarmSendsFlatBody(t *testing.T) {
	client, req := newAlarmTestClient(t, http.StatusCreated, `{"data":`+alarmEntryJSON+`}`)

	label, tm := "School", "06:30"
	enabled := true
	alarm, err := client.CreateAlarm(context.Background(), "f1", "d1", AlarmData{Label: &label, Time: &tm, Enabled: &enabled})
	if err != nil {
		t.Fatalf("CreateAlarm: %v", err)
	}
	if req.method != http.MethodPost || req.path != "/api/frames/f1/devices/d1/alarms" {
		t.Errorf("unexpected request %s %s", req.method, req.path)
	}
	if _, wrapped := req.body["alarm"]; wrapped {
		t.Error(`body must be flat, not wrapped in {"alarm": ...}`)
	}
	if req.body["label"] != "School" || req.body["time"] != "06:30" || req.body["enabled"] != true {
		t.Errorf("unexpected body: %v", req.body)
	}
	if alarm.ID != "a1" {
		t.Errorf("want ID a1, got %q", alarm.ID)
	}
}

func TestUpdateAlarmSendsOnlySetFields(t *testing.T) {
	client, req := newAlarmTestClient(t, http.StatusOK, `{"data":`+alarmEntryJSON+`}`)

	disabled := false
	if _, err := client.UpdateAlarm(context.Background(), "f1", "d1", "a1", AlarmData{Enabled: &disabled}); err != nil {
		t.Fatalf("UpdateAlarm: %v", err)
	}
	if req.method != http.MethodPatch || req.path != "/api/frames/f1/devices/d1/alarms/a1" {
		t.Errorf("unexpected request %s %s", req.method, req.path)
	}
	if len(req.body) != 1 || req.body["enabled"] != false {
		t.Errorf("want body {enabled:false} only, got %v", req.body)
	}
}

func TestDeleteAlarm(t *testing.T) {
	client, req := newAlarmTestClient(t, http.StatusNoContent, "")

	if err := client.DeleteAlarm(context.Background(), "f1", "d1", "a1"); err != nil {
		t.Fatalf("DeleteAlarm: %v", err)
	}
	if req.method != http.MethodDelete || req.path != "/api/frames/f1/devices/d1/alarms/a1" {
		t.Errorf("unexpected request %s %s", req.method, req.path)
	}
}

func TestAlarmMutationErrors(t *testing.T) {
	label := "x"
	calls := map[string]func(*Client) error{
		"create": func(c *Client) error {
			_, err := c.CreateAlarm(context.Background(), "f1", "d1", AlarmData{Label: &label})
			return err
		},
		"update": func(c *Client) error {
			_, err := c.UpdateAlarm(context.Background(), "f1", "d1", "a1", AlarmData{Label: &label})
			return err
		},
		"delete": func(c *Client) error {
			return c.DeleteAlarm(context.Background(), "f1", "d1", "a1")
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			client, _ := newAlarmTestClient(t, http.StatusUnprocessableEntity, `{"errors":["bad"]}`)
			if err := call(client); err == nil {
				t.Fatal("expected error on 422")
			}
		})
	}
}

func TestAlarmIDsAreEscaped(t *testing.T) {
	client, req := newAlarmTestClient(t, http.StatusNoContent, "")

	if err := client.DeleteAlarm(context.Background(), "f1", "d1", "a/../1"); err != nil {
		t.Fatalf("DeleteAlarm: %v", err)
	}
	if req.rawURI != "/api/frames/f1/devices/d1/alarms/a%2F..%2F1" {
		t.Errorf("alarm ID not escaped in path: %q", req.rawURI)
	}
}

func TestAlarmRRule(t *testing.T) {
	tests := []struct {
		name    string
		days    []string
		want    string
		wantErr bool
	}{
		{name: "weekdays", days: []string{"mon", "tue", "wed", "thu", "fri"}, want: "FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO,TU,WE,TH,FR"},
		{name: "case and spaces", days: []string{" Sat", "SUN "}, want: "FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=SA,SU"},
		{name: "duplicates removed", days: []string{"mon", "mon"}, want: "FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO"},
		{name: "invalid day", days: []string{"funday"}, wantErr: true},
		{name: "no days", days: nil, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AlarmRRule(tc.days)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseAlarmBody(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantID  string
		wantNil bool
	}{
		{name: "JSON-API envelope", body: `{"data":` + alarmEntryJSON + `}`, wantID: "a1"},
		{name: "flat object with string id", body: `{"id":"a2","label":"Gym","time":"05:45"}`, wantID: "a2"},
		{name: "flat object with numeric id", body: `{"id":42,"label":"Gym"}`, wantID: "42"},
		{name: "empty body", body: ``, wantNil: true},
		{name: "empty object", body: `{}`, wantNil: true},
		{name: "null id", body: `{"id":null}`, wantNil: true},
		{name: "not JSON", body: `ok`, wantNil: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseAlarmBody([]byte(tc.body))
			if tc.wantNil {
				if got != nil {
					t.Errorf("want nil, got %+v", got)
				}
				return
			}
			if got == nil || got.ID != tc.wantID {
				t.Fatalf("want ID %q, got %+v", tc.wantID, got)
			}
		})
	}

	flat := parseAlarmBody([]byte(`{"id":"a2","label":"Gym","time":"05:45","volume":30}`))
	if flat.Label != "Gym" || flat.Time != "05:45" || flat.Volume != 30 {
		t.Errorf("flat attributes not decoded: %+v", flat)
	}
}

func TestCreateAlarmEmptyResponse(t *testing.T) {
	client, _ := newAlarmTestClient(t, http.StatusCreated, "")
	label := "x"

	alarm, err := client.CreateAlarm(context.Background(), "f1", "d1", AlarmData{Label: &label})
	if err != nil {
		t.Fatalf("empty 201 body must not be an error: %v", err)
	}
	if alarm != nil {
		t.Errorf("want nil alarm, got %+v", alarm)
	}
}

func TestUpdateAlarmFallsBackToList(t *testing.T) {
	tests := []struct {
		name    string
		alarmID string
		wantNil bool
	}{
		{name: "found in list", alarmID: "a1"},
		{name: "missing from list", alarmID: "zzz", wantNil: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var methods []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method)
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"data":[` + alarmEntryJSON + `]}`))
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(srv.Close)
			client, _ := NewClientWithToken("u", "t", WithBaseURL(srv.URL))

			label := "x"
			alarm, err := client.UpdateAlarm(context.Background(), "f1", "d1", tc.alarmID, AlarmData{Label: &label})
			if err != nil {
				t.Fatalf("UpdateAlarm: %v", err)
			}
			if tc.wantNil != (alarm == nil) {
				t.Fatalf("wantNil=%v, got %+v", tc.wantNil, alarm)
			}
			if !tc.wantNil && alarm.ID != "a1" {
				t.Errorf("want re-read alarm a1, got %+v", alarm)
			}
			if len(methods) != 2 || methods[0] != http.MethodPatch || methods[1] != http.MethodGet {
				t.Errorf("want PATCH then GET, got %v", methods)
			}
		})
	}
}
