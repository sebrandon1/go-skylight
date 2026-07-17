package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListMealPlanSittings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("date_min"); got != "2026-04-01" {
			t.Errorf("date_min: want 2026-04-01 got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mealSittingAPIResponse{
			Data: []mealSittingAPIEntry{{ID: "s1"}},
		})
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	sittings, err := client.ListMealPlanSittings("frame1", MealSittingListOptions{
		DateMin: "2026-04-01",
		DateMax: "2026-04-30",
	})
	if err != nil {
		t.Fatalf("ListMealPlanSittings: %v", err)
	}
	if len(sittings) != 1 || sittings[0].ID != "s1" {
		t.Fatalf("unexpected sittings: %+v", sittings)
	}
}

func TestDeleteMealPlanSitting(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	if err := client.DeleteMealPlanSitting("frame1", "sit9", "2026-05-01"); err != nil {
		t.Fatalf("DeleteMealPlanSitting: %v", err)
	}
	if !strings.Contains(path, "/meals/sittings/sit9/instances/2026-05-01") {
		t.Fatalf("unexpected path %s", path)
	}
}

func TestDeleteMealPlanRange(t *testing.T) {
	var deletes int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(mealSittingAPIResponse{
				Data: []mealSittingAPIEntry{{ID: "s1"}, {ID: "s2"}},
			})
		case r.Method == http.MethodDelete:
			deletes++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	n, err := client.DeleteMealPlanRange("frame1", "2026-04-21", "2026-04-22")
	if err != nil {
		t.Fatalf("DeleteMealPlanRange: %v", err)
	}
	// Each sitting tries dates until one DELETE succeeds (first day works).
	if n != 2 {
		t.Fatalf("want deleted=2, got %d (deletes=%d)", n, deletes)
	}
	if deletes < 2 {
		t.Fatalf("expected at least 2 DELETE calls, got %d", deletes)
	}
}

func TestDeleteMealPlanRangeValidation(t *testing.T) {
	client, _ := NewClientWithToken("u", "t")
	if _, err := client.DeleteMealPlanRange("frame1", "", "2026-04-22"); err == nil {
		t.Fatal("expected error for empty date_min")
	}
	if _, err := client.DeleteMealPlanRange("frame1", "bad", "2026-04-22"); err == nil {
		t.Fatal("expected error for bad date_min")
	}
}

func TestDateRangeInclusive(t *testing.T) {
	got := dateRangeInclusive("2026-04-21", "2026-04-23")
	if len(got) != 3 || got[0] != "2026-04-21" || got[2] != "2026-04-23" {
		t.Fatalf("unexpected range: %v", got)
	}
}
