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
