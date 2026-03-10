package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCreateChoreRotation(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		idx := callCount.Add(1)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(choreAPISingleResponse{
			Data: choreAPIEntry{
				ID: fmt.Sprintf("ch%d", idx),
				Attributes: struct {
					Summary      string `json:"summary"`
					Status       string `json:"status"`
					Start        string `json:"start"`
					RewardPoints int    `json:"reward_points"`
					Recurring    bool   `json:"recurring"`
				}{
					Summary: body["summary"].(string),
					Start:   body["start"].(string),
				},
			},
		})
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	result, err := client.CreateChoreRotation("frame1", RotationData{
		Chores:      []string{"Dishes", "Vacuum"},
		AssigneeIDs: []string{"a1", "a2"},
		StartDate:   "2024-01-01",
		Weeks:       2,
		Points:      5,
	})
	if err != nil {
		t.Fatalf("CreateChoreRotation failed: %v", err)
	}

	// 2 chores * 2 weeks = 4 total
	if len(result.Chores) != 4 {
		t.Errorf("Expected 4 chores, got %d", len(result.Chores))
	}

	if callCount.Load() != 4 {
		t.Errorf("Expected 4 API calls, got %d", callCount.Load())
	}
}

func TestCreateChoreRotationValidation(t *testing.T) {
	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name string
		data RotationData
	}{
		{"no chores", RotationData{AssigneeIDs: []string{"a1"}, StartDate: "2024-01-01", Weeks: 1}},
		{"no assignees", RotationData{Chores: []string{"Task"}, StartDate: "2024-01-01", Weeks: 1}},
		{"zero weeks", RotationData{Chores: []string{"Task"}, AssigneeIDs: []string{"a1"}, StartDate: "2024-01-01", Weeks: 0}},
		{"bad date", RotationData{Chores: []string{"Task"}, AssigneeIDs: []string{"a1"}, StartDate: "not-a-date", Weeks: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateChoreRotation("frame1", tt.data)
			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func TestCreateChoreRotationPartialFailure(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		if idx >= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"fail"}`))
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(choreAPISingleResponse{
			Data: choreAPIEntry{
				ID: "ch1",
				Attributes: struct {
					Summary      string `json:"summary"`
					Status       string `json:"status"`
					Start        string `json:"start"`
					RewardPoints int    `json:"reward_points"`
					Recurring    bool   `json:"recurring"`
				}{Summary: "Task"},
			},
		})
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	result, err := client.CreateChoreRotation("frame1", RotationData{
		Chores:      []string{"Dishes", "Vacuum"},
		AssigneeIDs: []string{"a1", "a2"},
		StartDate:   "2024-01-01",
		Weeks:       2,
	})
	if err == nil {
		t.Fatal("Expected error for partial failure, got nil")
	}

	if result == nil {
		t.Fatal("Expected partial result, got nil")
	}
	if len(result.Chores) != 2 {
		t.Errorf("Expected 2 chores created before failure, got %d", len(result.Chores))
	}
}
