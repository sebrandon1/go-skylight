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

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
		check   func(t *testing.T, d *Dashboard)
	}{
		{
			name: "aggregates today's data",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				switch r.URL.Path {
				case "/api/frames/frame1/calendar_events":
					if r.URL.Query().Get("start_date") != today {
						t.Errorf("expected start_date=%s, got %s", today, r.URL.Query().Get("start_date"))
					}
					if err := json.NewEncoder(w).Encode(calendarAPIResponse{
						Data: []calendarAPIEntry{{ID: "ev1", Attributes: struct {
							Summary     string `json:"summary"`
							Description string `json:"description"`
							StartsAt    string `json:"starts_at"`
							EndsAt      string `json:"ends_at"`
							AllDay      bool   `json:"all_day"`
							Color       string `json:"color"`
						}{Summary: "Meeting"}}},
					}); err != nil {
						t.Errorf("encode events: %v", err)
					}
				case "/api/frames/frame1/chores":
					if r.URL.Query().Get("status") != "pending" {
						t.Errorf("expected status=pending, got %s", r.URL.Query().Get("status"))
					}
					if err := json.NewEncoder(w).Encode(choreAPIResponse{
						Data: []choreAPIEntry{
							{ID: "ch1", Attributes: struct {
								Summary      string `json:"summary"`
								Status       string `json:"status"`
								Start        string `json:"start"`
								RewardPoints int    `json:"reward_points"`
								Recurring    bool   `json:"recurring"`
							}{Summary: "Clean room", Status: "pending"}},
						},
					}); err != nil {
						t.Errorf("encode chores: %v", err)
					}
				case "/api/frames/frame1/reward_points":
					if err := json.NewEncoder(w).Encode([]RewardPointEntry{{CategoryID: 1, CurrentPointBalance: 50}}); err != nil {
						t.Errorf("encode points: %v", err)
					}
				case "/api/frames/frame1/meals/sittings":
					if err := json.NewEncoder(w).Encode(mealSittingAPIResponse{
						Data: []mealSittingAPIEntry{
							{ID: "ms1", Attributes: struct {
								Summary string `json:"summary"`
							}{Summary: "dinner"}},
						},
					}); err != nil {
						t.Errorf("encode sittings: %v", err)
					}
				case "/api/frames/frame1/lists":
					if err := json.NewEncoder(w).Encode(listAPIResponse{
						Data: []listAPIEntry{{ID: "l1", Attributes: struct {
							Label string `json:"label"`
							Color string `json:"color"`
							Kind  string `json:"kind"`
						}{Label: "Groceries"}}},
					}); err != nil {
						t.Errorf("encode lists: %v", err)
					}
				default:
					t.Errorf("unexpected path: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			},
			check: func(t *testing.T, d *Dashboard) {
				if d.Date != today {
					t.Errorf("want date=%s got %s", today, d.Date)
				}
				if len(d.Events) != 1 {
					t.Errorf("want 1 event, got %d", len(d.Events))
				}
				if len(d.Chores) != 1 {
					t.Errorf("want 1 chore, got %d", len(d.Chores))
				}
				if len(d.Points) != 1 {
					t.Errorf("want 1 point entry, got %d", len(d.Points))
				}
				if len(d.MealSittings) != 1 {
					t.Errorf("want 1 meal sitting, got %d", len(d.MealSittings))
				}
				if len(d.Lists) != 1 {
					t.Errorf("want 1 list, got %d", len(d.Lists))
				}
			},
		},
		{
			name: "first API error propagates",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
					t.Errorf("write: %v", err)
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			dash, err := client.GetDashboard("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && tc.check != nil {
				tc.check(t, dash)
			}
		})
	}
}
