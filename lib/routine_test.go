package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateRoutine(t *testing.T) {
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/chores/create_multiple" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":"97874955","attributes":{"summary":"Make bed","start":"2026-08-10","routine":true,"recurring":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"9740544","type":"category"}}}}]}`))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	routine, err := client.CreateRoutine(context.Background(), "frame1", RoutineData{
		Title:       "Make bed",
		TimeOfDay:   "morning",
		CategoryIDs: []string{"9740544"},
		StartDate:   "2026-08-10",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if routine.ID != "97874955" {
		t.Errorf("ID: want %q got %q", "97874955", routine.ID)
	}
	if routine.Title != "Make bed" {
		t.Errorf("Title: want %q got %q", "Make bed", routine.Title)
	}
	if routine.TimeOfDay != "morning" {
		t.Errorf("TimeOfDay: want %q got %q", "morning", routine.TimeOfDay)
	}
	if routine.AssigneeID != "9740544" {
		t.Errorf("AssigneeID: want %q got %q", "9740544", routine.AssigneeID)
	}
	if routine.StartDate != "2026-08-10" {
		t.Errorf("StartDate: want %q got %q", "2026-08-10", routine.StartDate)
	}

	if capturedBody["routine"] != true {
		t.Errorf("expected routine:true in request body, got %v", capturedBody["routine"])
	}
	catIDs, _ := capturedBody["category_ids"].([]any)
	if len(catIDs) != 1 || catIDs[0] != "9740544" {
		t.Errorf("expected category_ids [9740544], got %v", capturedBody["category_ids"])
	}
	recSet, _ := capturedBody["recurrence_set"].([]any)
	if len(recSet) != 1 || recSet[0] != "RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6" {
		t.Errorf("expected recurrence_set with BYHOUR=6, got %v", capturedBody["recurrence_set"])
	}
}

func TestCreateRoutine_InvalidTimeOfDay(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.CreateRoutine(context.Background(), "frame1", RoutineData{
		Title:     "Make bed",
		TimeOfDay: "noon",
	})
	if err == nil {
		t.Fatal("expected error for invalid time-of-day, got nil")
	}
	if called {
		t.Error("expected no HTTP call for an invalid time-of-day")
	}
}

func TestListRoutines_FiltersToRoutinesOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/frames/frame1/chores" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[
			{"id":"1","attributes":{"summary":"Make bed","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"c1","type":"category"}}}},
			{"id":"2","attributes":{"summary":"Take out trash","start":"2026-08-10","routine":false},"relationships":{"category":{"data":{"id":"c1","type":"category"}}}}
		]}`))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	routines, err := client.ListRoutines(context.Background(), "frame1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(routines) != 1 {
		t.Fatalf("expected 1 routine, got %d: %+v", len(routines), routines)
	}
	if routines[0].Title != "Make bed" {
		t.Errorf("Title: want %q got %q", "Make bed", routines[0].Title)
	}
}

func TestListRoutines_DedupesDailyOccurrences(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[
			{"id":"97871488-2026-08-10-0600","attributes":{"summary":"Make bed","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"c1","type":"category"}}}},
			{"id":"97871488-2026-08-11-0600","attributes":{"summary":"Make bed","start":"2026-08-11","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"c1","type":"category"}}}},
			{"id":"97871488-2026-08-12-0600","attributes":{"summary":"Make bed","start":"2026-08-12","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"c1","type":"category"}}}}
		]}`))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	routines, err := client.ListRoutines(context.Background(), "frame1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(routines) != 1 {
		t.Fatalf("expected 3 daily occurrences of the same routine to dedupe to 1, got %d: %+v", len(routines), routines)
	}
	if routines[0].ID != "97871488" {
		t.Errorf("ID: want base id %q got %q", "97871488", routines[0].ID)
	}
}

func TestListRoutines_MapsBYHOURToTimeOfDay(t *testing.T) {
	tests := []struct {
		byHour string
		want   string
	}{
		{"6", "morning"},
		{"14", "afternoon"},
		{"20", "evening"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"data":[{"id":"1","attributes":{"summary":"X","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=%s"]},"relationships":{"category":{"data":{"id":"c1","type":"category"}}}}]}`, tc.byHour)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			routines, err := client.ListRoutines(context.Background(), "frame1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(routines) != 1 {
				t.Fatalf("expected 1 routine, got %d", len(routines))
			}
			if routines[0].TimeOfDay != tc.want {
				t.Errorf("TimeOfDay: want %q got %q", tc.want, routines[0].TimeOfDay)
			}
		})
	}
}

func TestDeleteRoutine_SendsApplyToAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/chores/97874955" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("apply_to"); got != "all" {
			t.Errorf("apply_to: want %q got %q", "all", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	if err := client.DeleteRoutine(context.Background(), "frame1", "97874955"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
