package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// homeMockHandler backs both the plain chore list and ListRoutines, which
// both call GET /chores -- the plain chore list always sets status=pending,
// ListRoutines never does, so that's used to distinguish them below.
func homeMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/chores") && r.URL.Query().Get("status") == "pending":
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"pending"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"summary":"Morning Routine","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]}}]}`)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[{"id":"l1","type":"list","attributes":{"label":"Groceries"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}]}`)
		case strings.HasSuffix(r.URL.Path, "/meals/sittings"):
			fmt.Fprint(w, `{"data":[{"id":"s1","type":"meal_sitting","attributes":{"summary":"Dinner"}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	}
}

func TestHomeCmd_Text(t *testing.T) {
	newCmdTestClient(t, homeMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	origNoTasks, origNoLists, origNoMeals, origNoRoutines := homeNoTasks, homeNoLists, homeNoMeals, homeNoRoutines
	homeNoTasks, homeNoLists, homeNoMeals, homeNoRoutines = false, false, false, false
	t.Cleanup(func() {
		homeNoTasks, homeNoLists, homeNoMeals, homeNoRoutines = origNoTasks, origNoLists, origNoMeals, origNoRoutines
	})

	out := captureStdout(func() {
		if err := homeCmd.RunE(homeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dishes") || !strings.Contains(out, "Groceries") {
		t.Errorf("expected chore and list in output, got: %s", out)
	}
	if !strings.Contains(out, "Dinner") {
		t.Errorf("expected meal sitting in output, got: %s", out)
	}
	if !strings.Contains(out, "Morning Routine") {
		t.Errorf("expected routine in output, got: %s", out)
	}
}

func TestHomeCmd_JSON(t *testing.T) {
	newCmdTestClient(t, homeMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	origNoTasks, origNoLists, origNoMeals, origNoRoutines := homeNoTasks, homeNoLists, homeNoMeals, homeNoRoutines
	homeNoTasks, homeNoLists, homeNoMeals, homeNoRoutines = false, false, false, false
	t.Cleanup(func() {
		homeNoTasks, homeNoLists, homeNoMeals, homeNoRoutines = origNoTasks, origNoLists, origNoMeals, origNoRoutines
	})

	out := captureStdout(func() {
		if err := homeCmd.RunE(homeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, `"week_start"`) {
		t.Errorf("expected week_start in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"meals"`) {
		t.Errorf("expected meals key in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"routines"`) {
		t.Errorf("expected routines key in JSON output, got: %s", out)
	}
}

func TestHomeCmd_NoTasksNoLists(t *testing.T) {
	newCmdTestClient(t, homeMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	origNoTasks, origNoLists, origNoRoutines := homeNoTasks, homeNoLists, homeNoRoutines
	homeNoTasks, homeNoLists, homeNoRoutines = true, true, false
	t.Cleanup(func() { homeNoTasks, homeNoLists, homeNoRoutines = origNoTasks, origNoLists, origNoRoutines })

	out := captureStdout(func() {
		if err := homeCmd.RunE(homeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "PENDING TASKS") || strings.Contains(out, "=== LISTS ===") {
		t.Errorf("expected tasks/lists sections to be skipped, got: %s", out)
	}
}

func TestHomeCmd_NoMeals(t *testing.T) {
	newCmdTestClient(t, homeMockHandler())
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	origNoMeals, origNoRoutines := homeNoMeals, homeNoRoutines
	homeNoMeals, homeNoRoutines = true, false
	t.Cleanup(func() { homeNoMeals, homeNoRoutines = origNoMeals, origNoRoutines })

	out := captureStdout(func() {
		if err := homeCmd.RunE(homeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "=== MEALS THIS WEEK ===") {
		t.Errorf("expected meals section to be skipped, got: %s", out)
	}
}

func TestHomeCmd_NoRoutines(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/chores") && r.URL.Query().Get("status") != "pending" {
			t.Error("routines call (chores without status filter) should not happen with --no-routines")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		homeMockHandler()(w, r)
	})
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = ""

	origNoRoutines := homeNoRoutines
	homeNoRoutines = true
	t.Cleanup(func() { homeNoRoutines = origNoRoutines })

	out := captureStdout(func() {
		if err := homeCmd.RunE(homeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "=== ROUTINES ===") {
		t.Errorf("expected routines section to be skipped with --no-routines, got: %s", out)
	}
}

// TestHomeCmd_RoutinesNotFound documents a deliberate behavior change: a 404
// on the routines call used to be swallowed as "zero routines" because it
// hit its own dedicated (fictional) /routines endpoint. Now ListRoutines
// shares the plain /chores endpoint, so a 404 there is a generic listing
// failure -- same as it already is for the plain chore list in this same
// command -- and should surface as an error instead of being hidden.
func TestHomeCmd_RoutinesNotFound(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/chores") && r.URL.Query().Get("status") != "pending" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		homeMockHandler()(w, r)
	})
	origNoRoutines := homeNoRoutines
	homeNoRoutines = false
	t.Cleanup(func() { homeNoRoutines = origNoRoutines })

	err := homeCmd.RunE(homeCmd, nil)
	if err == nil {
		t.Fatal("expected error when routines API returns 404, got nil")
	}
	if !strings.Contains(err.Error(), "listing routines") {
		t.Errorf("expected 'listing routines' in error, got: %v", err)
	}
}

func TestHomeCmd_RoutinesFetchError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/chores") && r.URL.Query().Get("status") != "pending" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		homeMockHandler()(w, r)
	})
	origNoRoutines := homeNoRoutines
	homeNoRoutines = false
	t.Cleanup(func() { homeNoRoutines = origNoRoutines })

	err := homeCmd.RunE(homeCmd, nil)
	if err == nil {
		t.Fatal("expected error when routines API returns 500, got nil")
	}
	if !strings.Contains(err.Error(), "listing routines") {
		t.Errorf("expected 'listing routines' in error, got: %v", err)
	}
}

func TestHomeCmd_FetchError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/calendar_events") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
	})

	err := homeCmd.RunE(homeCmd, nil)
	if err == nil {
		t.Fatal("expected error when calendar fetch fails")
	}
	if !strings.Contains(err.Error(), "listing calendar events") {
		t.Errorf("expected 'listing calendar events' in error, got: %v", err)
	}
}

func TestHomeCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "home")
}
