package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetDashboard(t *testing.T) {
	today := time.Now().Format(DateFormat)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/frames/frame1/calendar_events":
			if r.URL.Query().Get("start_date") != today {
				t.Errorf("Expected start_date %s, got %s", today, r.URL.Query().Get("start_date"))
			}
			json.NewEncoder(w).Encode([]CalendarEvent{{ID: "ev1", Title: "Meeting"}})
		case "/api/frames/frame1/chores":
			if r.URL.Query().Get("status") != "pending" {
				t.Errorf("Expected status pending, got %s", r.URL.Query().Get("status"))
			}
			json.NewEncoder(w).Encode(choreAPIResponse{
				Data: []choreAPIEntry{
					{ID: "ch1", Attributes: struct {
						Summary      string `json:"summary"`
						Status       string `json:"status"`
						Start        string `json:"start"`
						RewardPoints int    `json:"reward_points"`
						Recurring    bool   `json:"recurring"`
					}{Summary: "Clean room", Status: "pending"}},
				},
			})
		case "/api/frames/frame1/reward_points":
			json.NewEncoder(w).Encode([]RewardPointEntry{{CategoryID: 1, CurrentPointBalance: 50}})
		case "/api/frames/frame1/meals/sittings":
			json.NewEncoder(w).Encode([]MealSitting{
				{ID: "ms1", Date: today, MealType: "dinner"},
				{ID: "ms2", Date: "2020-01-01", MealType: "lunch"},
			})
		case "/api/frames/frame1/lists":
			json.NewEncoder(w).Encode([]List{{ID: "l1", Title: "Groceries"}})
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	dash, err := client.GetDashboard("frame1")
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}

	if dash.Date != today {
		t.Errorf("Expected date %s, got %s", today, dash.Date)
	}
	if len(dash.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(dash.Events))
	}
	if len(dash.Chores) != 1 {
		t.Errorf("Expected 1 chore, got %d", len(dash.Chores))
	}
	if len(dash.Points) != 1 {
		t.Errorf("Expected 1 point entry, got %d", len(dash.Points))
	}
	if len(dash.MealSittings) != 1 {
		t.Errorf("Expected 1 meal sitting (today only), got %d", len(dash.MealSittings))
	}
	if len(dash.Lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(dash.Lists))
	}
}

func TestGetDashboardFailsOnFirstError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"fail"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetDashboard("frame1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
