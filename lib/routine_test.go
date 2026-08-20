package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		Title:      "Make bed",
		TimeOfDay:  "morning",
		CategoryID: "9740544",
		StartDate:  "2026-08-10",
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
	if routine.NextOccurrenceDate != "2026-08-10" {
		t.Errorf("NextOccurrenceDate: want %q got %q", "2026-08-10", routine.NextOccurrenceDate)
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

func TestCreateRoutine_EmptyCategoryID(t *testing.T) {
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
		TimeOfDay: "morning",
		StartDate: "2026-08-10",
	})
	if err == nil {
		t.Fatal("expected error for empty category ID, got nil")
	}
	if called {
		t.Error("expected no HTTP call for an empty category ID")
	}
}

func TestCreateRoutine_EmptyResponseData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.CreateRoutine(context.Background(), "frame1", RoutineData{
		Title:      "Make bed",
		TimeOfDay:  "morning",
		CategoryID: "9740544",
		StartDate:  "2026-08-10",
	})
	if err == nil {
		t.Fatal("expected error when create_multiple returns no chores, got nil")
	}
}

func TestCreateRoutine_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.CreateRoutine(context.Background(), "frame1", RoutineData{
		Title:      "Make bed",
		TimeOfDay:  "morning",
		CategoryID: "9740544",
		StartDate:  "2026-08-10",
	})
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}

func TestCreateRoutine_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid`))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.CreateRoutine(context.Background(), "frame1", RoutineData{
		Title:      "Make bed",
		TimeOfDay:  "morning",
		CategoryID: "9740544",
		StartDate:  "2026-08-10",
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON response, got nil")
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

func TestListRoutines_QueriesTodayThroughLookaheadWindow(t *testing.T) {
	wantAfter := time.Now().Format(DateFormat)
	wantBefore := time.Now().AddDate(0, 0, routineLookaheadDays).Format(DateFormat)

	var gotAfter, gotBefore string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAfter = r.URL.Query().Get("after")
		gotBefore = r.URL.Query().Get("before")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	if _, err := client.ListRoutines(context.Background(), "frame1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAfter != wantAfter {
		t.Errorf("after: want %q got %q", wantAfter, gotAfter)
	}
	if gotBefore != wantBefore {
		t.Errorf("before: want %q got %q", wantBefore, gotBefore)
	}
}

func TestListRoutines_DoesNotGroupDifferentBaseIDs(t *testing.T) {
	// create_multiple's sibling records (one per assignee) share no
	// server-provided correlation key -- verified live, group/series are
	// each self-referential to their own record id, not shared across
	// siblings. ListRoutines must not invent a client-side heuristic that
	// merges unrelated base IDs just because they share a title/date/rule.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[
			{"id":"111","attributes":{"summary":"Make bed","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"c1","type":"category"}}}},
			{"id":"222","attributes":{"summary":"Make bed","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]},"relationships":{"category":{"data":{"id":"c2","type":"category"}}}}
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
	if len(routines) != 2 {
		t.Fatalf("expected 2 separate routine entries (different base IDs, no grouping), got %d: %+v", len(routines), routines)
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

func TestListRoutines_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	if _, err := client.ListRoutines(context.Background(), "frame1"); err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}

func TestListRoutines_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	if _, err := client.ListRoutines(context.Background(), "frame1"); err == nil {
		t.Fatal("expected error for invalid JSON response, got nil")
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

func TestDeleteRoutine_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	if err := client.DeleteRoutine(context.Background(), "frame1", "97874955"); err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}

func TestUpdateRoutine(t *testing.T) {
	var gotApplyTo string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		gotApplyTo = r.URL.Query().Get("apply_to")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"97874955","attributes":{"summary":"Evening Walk","start":"2026-08-10","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=20"]},"relationships":{"category":{"data":{"id":"9740544","type":"category"}}}}}`)
	}))
	defer srv.Close()
	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()
	client, _ := NewClientWithToken("u", "t")
	routine, err := client.UpdateRoutine(context.Background(), "frame1", "97874955", RoutineUpdateData{Title: "Evening Walk", TimeOfDay: "evening"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if routine.Title != "Evening Walk" {
		t.Errorf("Title: want %q got %q", "Evening Walk", routine.Title)
	}
	if routine.TimeOfDay != "evening" {
		t.Errorf("TimeOfDay: want %q got %q", "evening", routine.TimeOfDay)
	}
	if gotApplyTo != "all" {
		t.Errorf("apply_to: want %q got %q", "all", gotApplyTo)
	}
}

func TestUpdateRoutine_InvalidTimeOfDay(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()
	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()
	client, _ := NewClientWithToken("u", "t")
	_, err := client.UpdateRoutine(context.Background(), "frame1", "97874955", RoutineUpdateData{TimeOfDay: "midnight"})
	if err == nil {
		t.Fatal("expected error for invalid time-of-day, got nil")
	}
	if called {
		t.Error("expected no HTTP call for an invalid time-of-day")
	}
}

func TestUpdateRoutine_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()
	client, _ := NewClientWithToken("u", "t")
	_, err := client.UpdateRoutine(context.Background(), "frame1", "97874955", RoutineUpdateData{Title: "Test"})
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}
