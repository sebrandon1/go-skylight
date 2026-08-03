package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCreateChoreRotation(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		idx := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
			Data: choreAPIEntry{
				ID: fmt.Sprintf("ch%d", idx),
				Attributes: choreAPIAttributes{
					Summary: body["summary"].(string),
					Start:   body["start"].(string),
				},
			},
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	result, err := client.CreateChoreRotation(context.Background(), "frame1", RotationData{
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
		t.Errorf("want 4 chores, got %d", len(result.Chores))
	}
	if callCount.Load() != 4 {
		t.Errorf("want 4 API calls, got %d", callCount.Load())
	}
}

func TestCreateChoreRotationValidation(t *testing.T) {
	client, _ := NewClientWithToken("u", "t")

	tests := []struct {
		name string
		data RotationData
	}{
		{"no chores", RotationData{AssigneeIDs: []string{"a1"}, StartDate: "2024-01-01", Weeks: 1}},
		{"no assignees", RotationData{Chores: []string{"Task"}, StartDate: "2024-01-01", Weeks: 1}},
		{"zero weeks", RotationData{Chores: []string{"Task"}, AssigneeIDs: []string{"a1"}, StartDate: "2024-01-01", Weeks: 0}},
		{"bad date", RotationData{Chores: []string{"Task"}, AssigneeIDs: []string{"a1"}, StartDate: "not-a-date", Weeks: 1}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.CreateChoreRotation(context.Background(), "frame1", tc.data)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestCreateChoreRotationPartialFailure(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if idx >= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
				t.Errorf("write: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
			Data: choreAPIEntry{
				ID:         "ch1",
				Attributes: choreAPIAttributes{Summary: "Task"},
			},
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	result, err := client.CreateChoreRotation(context.Background(), "frame1", RotationData{
		Chores:      []string{"Dishes", "Vacuum"},
		AssigneeIDs: []string{"a1", "a2"},
		StartDate:   "2024-01-01",
		Weeks:       2,
	})
	if err == nil {
		t.Fatal("expected error for partial failure, got nil")
	}
	if result == nil {
		t.Fatal("expected partial result, got nil")
	}
	// Parallel workers may complete more than two creates before failures land;
	// require at least one success and fewer than the full 4-job batch.
	if len(result.Chores) < 1 {
		t.Errorf("want some partial chores, got %d", len(result.Chores))
	}
	if len(result.Chores) >= 4 {
		t.Errorf("want incomplete batch on partial failure, got %d chores", len(result.Chores))
	}
}
