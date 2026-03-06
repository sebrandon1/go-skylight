package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCalendarEvents(t *testing.T) {
	mockEvents := []CalendarEvent{
		{ID: "1", Title: "Meeting", StartAt: "2024-01-15T10:00:00Z", EndAt: "2024-01-15T11:00:00Z"},
		{ID: "2", Title: "Lunch", StartAt: "2024-01-15T12:00:00Z", EndAt: "2024-01-15T13:00:00Z"},
	}

	mockResponseJSON, _ := json.Marshal(mockEvents)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/calendar_events" {
			t.Errorf("Expected path /api/frames/frame1/calendar_events, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	events, err := client.ListCalendarEvents("frame1", "", "")
	if err != nil {
		t.Fatalf("ListCalendarEvents failed: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].Title != "Meeting" {
		t.Errorf("Expected title 'Meeting', got '%s'", events[0].Title)
	}
}

func TestCreateCalendarEvent(t *testing.T) {
	mockEvent := CalendarEvent{ID: "3", Title: "New Event", StartAt: "2024-01-16T09:00:00Z"}

	mockResponseJSON, _ := json.Marshal(mockEvent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/calendar_events" {
			t.Errorf("Expected path /api/frames/frame1/calendar_events, got %s", r.URL.Path)
		}

		var reqBody CalendarEventRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.CalendarEvent.Title != "New Event" {
			t.Errorf("Expected title 'New Event', got '%s'", reqBody.CalendarEvent.Title)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	event, err := client.CreateCalendarEvent("frame1", CalendarEventData{Title: "New Event", StartAt: "2024-01-16T09:00:00Z"})
	if err != nil {
		t.Fatalf("CreateCalendarEvent failed: %v", err)
	}

	if event.Title != "New Event" {
		t.Errorf("Expected title 'New Event', got '%s'", event.Title)
	}
}

func TestDeleteCalendarEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/calendar_events/evt1" {
			t.Errorf("Expected path /api/frames/frame1/calendar_events/evt1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.DeleteCalendarEvent("frame1", "evt1")
	if err != nil {
		t.Fatalf("DeleteCalendarEvent failed: %v", err)
	}
}

func TestListCalendarEventsWithDateRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start_date") != "2024-01-01" {
			t.Errorf("Expected start_date query param '2024-01-01', got '%s'", r.URL.Query().Get("start_date"))
		}
		if r.URL.Query().Get("end_date") != "2024-01-31" {
			t.Errorf("Expected end_date query param '2024-01-31', got '%s'", r.URL.Query().Get("end_date"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListCalendarEvents("frame1", "2024-01-01", "2024-01-31")
	if err != nil {
		t.Fatalf("ListCalendarEvents with date range failed: %v", err)
	}
}

func TestUpdateCalendarEvent(t *testing.T) {
	mockEvent := CalendarEvent{ID: "1", Title: "Updated Event"}

	mockResponseJSON, _ := json.Marshal(mockEvent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/calendar_events/evt1" {
			t.Errorf("Expected path /api/frames/frame1/calendar_events/evt1, got %s", r.URL.Path)
		}

		var reqBody CalendarEventRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.CalendarEvent.Title != "Updated Event" {
			t.Errorf("Expected title 'Updated Event', got '%s'", reqBody.CalendarEvent.Title)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	event, err := client.UpdateCalendarEvent("frame1", "evt1", CalendarEventData{Title: "Updated Event"})
	if err != nil {
		t.Fatalf("UpdateCalendarEvent failed: %v", err)
	}

	if event.Title != "Updated Event" {
		t.Errorf("Expected title 'Updated Event', got '%s'", event.Title)
	}
}

func TestListSourceCalendars(t *testing.T) {
	mockCalendars := []SourceCalendar{
		{ID: "1", Name: "Google Calendar", Enabled: true, Provider: "google"},
	}

	mockResponseJSON, _ := json.Marshal(mockCalendars)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/source_calendars" {
			t.Errorf("Expected path /api/frames/frame1/source_calendars, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	calendars, err := client.ListSourceCalendars("frame1")
	if err != nil {
		t.Fatalf("ListSourceCalendars failed: %v", err)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar, got %d", len(calendars))
	}

	if calendars[0].Provider != "google" {
		t.Errorf("Expected provider 'google', got '%s'", calendars[0].Provider)
	}
}

func TestCalendarEventErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not found"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListCalendarEvents("frame1", "", "")
	if err == nil {
		t.Error("Expected error for not found, got nil")
	}

	_, err = client.CreateCalendarEvent("frame1", CalendarEventData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for not found, got nil")
	}

	err = client.DeleteCalendarEvent("frame1", "nonexistent")
	if err == nil {
		t.Error("Expected error for not found, got nil")
	}
}

func TestUpdateCalendarEventError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Server error"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.UpdateCalendarEvent("frame1", "evt1", CalendarEventData{Title: "Test"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestListSourceCalendarsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Server error"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListSourceCalendars("frame1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCalendarInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListCalendarEvents("frame1", "", "")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestCreateCalendarEventRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody CalendarEventRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.CalendarEvent.Title != "Birthday Party" {
			t.Errorf("Expected title 'Birthday Party', got '%s'", reqBody.CalendarEvent.Title)
		}
		if reqBody.CalendarEvent.Description != "Fun party" {
			t.Errorf("Expected description 'Fun party', got '%s'", reqBody.CalendarEvent.Description)
		}
		if reqBody.CalendarEvent.StartAt != "2024-06-15T14:00:00Z" {
			t.Errorf("Expected start_at '2024-06-15T14:00:00Z', got '%s'", reqBody.CalendarEvent.StartAt)
		}
		if reqBody.CalendarEvent.EndAt != "2024-06-15T18:00:00Z" {
			t.Errorf("Expected end_at '2024-06-15T18:00:00Z', got '%s'", reqBody.CalendarEvent.EndAt)
		}
		if reqBody.CalendarEvent.Color != "#FF0000" {
			t.Errorf("Expected color '#FF0000', got '%s'", reqBody.CalendarEvent.Color)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"evt1","title":"Birthday Party"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.CreateCalendarEvent("frame1", CalendarEventData{
		Title:       "Birthday Party",
		Description: "Fun party",
		StartAt:     "2024-06-15T14:00:00Z",
		EndAt:       "2024-06-15T18:00:00Z",
		Color:       "#FF0000",
	})
	if err != nil {
		t.Fatalf("CreateCalendarEvent failed: %v", err)
	}
}

func TestDeleteCalendarEventError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Server error"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.DeleteCalendarEvent("frame1", "evt1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestListCalendarEventsWithStartDateOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start_date") != "2024-01-01" {
			t.Errorf("Expected start_date '2024-01-01', got '%s'", r.URL.Query().Get("start_date"))
		}
		if r.URL.Query().Get("end_date") != "" {
			t.Errorf("Expected no end_date param, got '%s'", r.URL.Query().Get("end_date"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListCalendarEvents("frame1", "2024-01-01", "")
	if err != nil {
		t.Fatalf("ListCalendarEvents with start date only failed: %v", err)
	}
}

func TestListCalendarEventsWithEndDateOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start_date") != "" {
			t.Errorf("Expected no start_date param, got '%s'", r.URL.Query().Get("start_date"))
		}
		if r.URL.Query().Get("end_date") != "2024-01-31" {
			t.Errorf("Expected end_date '2024-01-31', got '%s'", r.URL.Query().Get("end_date"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListCalendarEvents("frame1", "", "2024-01-31")
	if err != nil {
		t.Fatalf("ListCalendarEvents with end date only failed: %v", err)
	}
}
