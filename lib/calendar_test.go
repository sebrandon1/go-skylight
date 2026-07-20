package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCalendarEvents(t *testing.T) {
	tests := []struct {
		name           string
		dateMin        string
		dateMax        string
		timezone       string
		status         int
		response       string
		wantLen        int
		wantFirstTitle string
		wantCategoryID string
		wantErr        bool
	}{
		{
			name:           "returns events",
			status:         http.StatusOK,
			response:       `{"data":[{"id":"1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2024-01-15T10:00:00-06:00","ends_at":"2024-01-15T11:00:00-06:00","all_day":false},"relationships":{"categories":{"data":[]}}},{"id":"2","type":"calendar_event","attributes":{"summary":"Lunch","starts_at":"2024-01-15T12:00:00-06:00","ends_at":"2024-01-15T13:00:00-06:00","all_day":false},"relationships":{"categories":{"data":[]}}}]}`,
			wantLen:        2,
			wantFirstTitle: "Meeting",
		},
		{
			name:           "parses category from categories relationship",
			status:         http.StatusOK,
			response:       `{"data":[{"id":"1","type":"calendar_event","attributes":{"summary":"Party","starts_at":"2024-06-01T10:00:00-05:00","all_day":false},"relationships":{"categories":{"data":[{"id":"cat42","type":"category"}]}}}]}`,
			wantLen:        1,
			wantFirstTitle: "Party",
			wantCategoryID: "cat42",
		},
		{
			name:     "passes date range and timezone params",
			dateMin:  "2024-01-01",
			dateMax:  "2024-01-31",
			timezone: "America/Chicago",
			status:   http.StatusOK,
			response: `{"data":[]}`,
		},
		{
			name:    "not found returns error",
			status:  http.StatusNotFound,
			wantErr: true,
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			status:   http.StatusOK,
			response: `not valid json`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/calendar_events" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				q := r.URL.Query()
				if tc.dateMin != "" && q.Get("date_min") != tc.dateMin {
					t.Errorf("date_min: want %q got %q", tc.dateMin, q.Get("date_min"))
				}
				if tc.dateMax != "" && q.Get("date_max") != tc.dateMax {
					t.Errorf("date_max: want %q got %q", tc.dateMax, q.Get("date_max"))
				}
				if tc.timezone != "" && q.Get("timezone") != tc.timezone {
					t.Errorf("timezone: want %q got %q", tc.timezone, q.Get("timezone"))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			events, err := client.ListCalendarEvents("frame1", tc.dateMin, tc.dateMax, tc.timezone)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(events) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(events))
			}
			if tc.wantFirstTitle != "" && len(events) > 0 && events[0].Title != tc.wantFirstTitle {
				t.Errorf("events[0].Title: want %q got %q", tc.wantFirstTitle, events[0].Title)
			}
			if tc.wantCategoryID != "" && len(events) > 0 && events[0].CategoryID != tc.wantCategoryID {
				t.Errorf("events[0].CategoryID: want %q got %q", tc.wantCategoryID, events[0].CategoryID)
			}
		})
	}
}

func TestCreateCalendarEvent(t *testing.T) {
	trueVal := true
	tests := []struct {
		name           string
		input          CalendarEventData
		status         int
		response       string
		wantTitle      string
		wantBodyAllDay *bool
		wantErr        bool
	}{
		{
			name:      "creates event",
			input:     CalendarEventData{Title: "New Event", StartAt: "2024-01-16T09:00:00-06:00"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"3","type":"calendar_event","attributes":{"summary":"New Event","starts_at":"2024-01-16T09:00:00-06:00","ends_at":"","all_day":false},"relationships":{"categories":{"data":[]}}}}`,
			wantTitle: "New Event",
		},
		{
			name:      "sends all fields in request body",
			input:     CalendarEventData{Title: "Birthday Party", Description: "Fun party", StartAt: "2024-06-15T14:00:00-05:00", EndAt: "2024-06-15T18:00:00-05:00"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"evt1","type":"calendar_event","attributes":{"summary":"Birthday Party","starts_at":"2024-06-15T14:00:00-05:00","ends_at":"2024-06-15T18:00:00-05:00","all_day":false},"relationships":{"categories":{"data":[]}}}}`,
			wantTitle: "Birthday Party",
		},
		{
			name:           "AllDay true is sent in request body",
			input:          CalendarEventData{Title: "Holiday", AllDay: &trueVal},
			status:         http.StatusCreated,
			response:       `{"data":{"id":"4","type":"calendar_event","attributes":{"summary":"Holiday","all_day":true},"relationships":{"categories":{"data":[]}}}}`,
			wantTitle:      "Holiday",
			wantBodyAllDay: &trueVal,
		},
		{
			name:    "not found returns error",
			input:   CalendarEventData{Title: "Test"},
			status:  http.StatusNotFound,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/calendar_events" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
					t.Errorf("decoding request body: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			event, err := client.CreateCalendarEvent("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if event.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, event.Title)
			}
			if tc.wantBodyAllDay != nil {
				got, ok := capturedBody["all_day"].(bool)
				if !ok || got != *tc.wantBodyAllDay {
					t.Errorf("request body all_day: want %v got %v (present=%v)", *tc.wantBodyAllDay, got, ok)
				}
			}
		})
	}
}

func TestUpdateCalendarEvent(t *testing.T) {
	falseVal := false
	tests := []struct {
		name           string
		eventID        string
		input          CalendarEventData
		status         int
		response       string
		wantTitle      string
		wantBodyAllDay *bool
		wantErr        bool
	}{
		{
			name:      "updates event",
			eventID:   "evt1",
			input:     CalendarEventData{Title: "Updated Event"},
			status:    http.StatusOK,
			response:  `{"data":{"id":"1","type":"calendar_event","attributes":{"summary":"Updated Event","starts_at":"","ends_at":"","all_day":false},"relationships":{"categories":{"data":[]}}}}`,
			wantTitle: "Updated Event",
		},
		{
			name:           "AllDay false is sent in request body",
			eventID:        "evt1",
			input:          CalendarEventData{AllDay: &falseVal},
			status:         http.StatusOK,
			response:       `{"data":{"id":"1","type":"calendar_event","attributes":{"summary":"","starts_at":"","ends_at":"","all_day":false},"relationships":{"categories":{"data":[]}}}}`,
			wantBodyAllDay: &falseVal,
		},
		{
			name:    "server error returns error",
			eventID: "evt1",
			input:   CalendarEventData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			eventID:  "evt1",
			input:    CalendarEventData{Title: "Test"},
			status:   http.StatusOK,
			response: `{invalid json`,
			wantErr:  true,
		},
		{
			name:    "bad request returns error",
			eventID: "evt1",
			input:   CalendarEventData{Title: "Test"},
			status:  http.StatusBadRequest,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				_ = json.NewDecoder(r.Body).Decode(&capturedBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			event, err := client.UpdateCalendarEvent("frame1", tc.eventID, tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if tc.wantTitle != "" && event.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, event.Title)
			}
			if tc.wantBodyAllDay != nil {
				got, ok := capturedBody["all_day"].(bool)
				if !ok || got != *tc.wantBodyAllDay {
					t.Errorf("request body all_day: want %v got %v (present=%v)", *tc.wantBodyAllDay, got, ok)
				}
			}
		})
	}
}

func TestDeleteCalendarEvent(t *testing.T) {
	tests := []struct {
		name    string
		eventID string
		status  int
		wantErr bool
	}{
		{
			name:    "deletes with 204",
			eventID: "evt1",
			status:  http.StatusNoContent,
		},
		{
			name:    "deletes with 200 OK",
			eventID: "evt1",
			status:  http.StatusOK,
		},
		{
			name:    "server error returns error",
			eventID: "evt1",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.DeleteCalendarEvent("frame1", tc.eventID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestListSourceCalendars(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		response     string
		wantLen      int
		wantProvider string
		checkEnabled bool
		wantEnabled  bool
		wantErr      bool
	}{
		{
			name:         "returns calendars",
			status:       http.StatusOK,
			response:     `{"data":[{"id":"1","type":"source_calendar_detail","attributes":{"label":"Google Calendar","kind":"google","editable":true}}]}`,
			wantLen:      1,
			wantProvider: "google",
			checkEnabled: true,
			wantEnabled:  true,
		},
		{
			name:         "editable false maps to Enabled false",
			status:       http.StatusOK,
			response:     `{"data":[{"id":"2","type":"source_calendar_detail","attributes":{"label":"Read-only","kind":"gmail","editable":false}}]}`,
			wantLen:      1,
			checkEnabled: true,
			wantEnabled:  false,
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/source_calendars" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			calendars, err := client.ListSourceCalendars("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(calendars) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(calendars))
			}
			if tc.wantProvider != "" && len(calendars) > 0 && calendars[0].Provider != tc.wantProvider {
				t.Errorf("Provider: want %q got %q", tc.wantProvider, calendars[0].Provider)
			}
			if tc.checkEnabled && len(calendars) > 0 && calendars[0].Enabled != tc.wantEnabled {
				t.Errorf("Enabled: want %v got %v", tc.wantEnabled, calendars[0].Enabled)
			}
		})
	}
}
