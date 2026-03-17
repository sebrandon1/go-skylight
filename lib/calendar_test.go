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
		startDate      string
		endDate        string
		status         int
		response       string
		wantLen        int
		wantFirstTitle string
		wantErr        bool
	}{
		{
			name:           "returns events",
			status:         http.StatusOK,
			response:       `[{"id":"1","title":"Meeting","start_at":"2024-01-15T10:00:00Z"},{"id":"2","title":"Lunch","start_at":"2024-01-15T12:00:00Z"}]`,
			wantLen:        2,
			wantFirstTitle: "Meeting",
		},
		{
			name:      "passes date range params",
			startDate: "2024-01-01",
			endDate:   "2024-01-31",
			status:    http.StatusOK,
			response:  `[]`,
		},
		{
			name:      "passes start date only",
			startDate: "2024-01-01",
			status:    http.StatusOK,
			response:  `[]`,
		},
		{
			name:     "passes end date only",
			endDate:  "2024-01-31",
			status:   http.StatusOK,
			response: `[]`,
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
				if tc.startDate != "" && q.Get("start_date") != tc.startDate {
					t.Errorf("start_date: want %q got %q", tc.startDate, q.Get("start_date"))
				}
				if tc.endDate != "" && q.Get("end_date") != tc.endDate {
					t.Errorf("end_date: want %q got %q", tc.endDate, q.Get("end_date"))
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
			events, err := client.ListCalendarEvents("frame1", tc.startDate, tc.endDate)
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
		})
	}
}

func TestCreateCalendarEvent(t *testing.T) {
	tests := []struct {
		name      string
		input     CalendarEventData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "creates event",
			input:     CalendarEventData{Title: "New Event", StartAt: "2024-01-16T09:00:00Z"},
			status:    http.StatusCreated,
			response:  `{"id":"3","title":"New Event","start_at":"2024-01-16T09:00:00Z"}`,
			wantTitle: "New Event",
		},
		{
			name:      "sends all fields in request body",
			input:     CalendarEventData{Title: "Birthday Party", Description: "Fun party", StartAt: "2024-06-15T14:00:00Z", EndAt: "2024-06-15T18:00:00Z", Color: "#FF0000"},
			status:    http.StatusCreated,
			response:  `{"id":"evt1","title":"Birthday Party"}`,
			wantTitle: "Birthday Party",
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
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/calendar_events" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var body CalendarEventRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if body.CalendarEvent.Title != tc.input.Title {
					t.Errorf("title: want %q got %q", tc.input.Title, body.CalendarEvent.Title)
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
			if !tc.wantErr && event.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, event.Title)
			}
		})
	}
}

func TestUpdateCalendarEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventID   string
		input     CalendarEventData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "updates event",
			eventID:   "evt1",
			input:     CalendarEventData{Title: "Updated Event"},
			status:    http.StatusOK,
			response:  `{"id":"1","title":"Updated Event"}`,
			wantTitle: "Updated Event",
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
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
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
			event, err := client.UpdateCalendarEvent("frame1", tc.eventID, tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && event.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, event.Title)
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
		wantErr      bool
	}{
		{
			name:         "returns calendars",
			status:       http.StatusOK,
			response:     `[{"id":"1","name":"Google Calendar","enabled":true,"provider":"google"}]`,
			wantLen:      1,
			wantProvider: "google",
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
		})
	}
}
