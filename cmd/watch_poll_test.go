package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestContainsResource(t *testing.T) {
	resources := []string{watchResourceRewards, watchResourceChores}
	if !containsResource(resources, watchResourceRewards) {
		t.Error("expected rewards to be found")
	}
	if containsResource(resources, watchResourceCalendar) {
		t.Error("expected calendar not to be found")
	}
}

func TestFilterOutResource(t *testing.T) {
	resources := []string{watchResourceRewards, watchResourceChores, watchResourceCalendar}
	got := filterOutResource(resources, watchResourceChores)
	want := []string{watchResourceRewards, watchResourceCalendar}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestPrintRedemptionEvent(t *testing.T) {
	e := lib.RedemptionEvent{
		RewardID: "r1", RewardName: "Ice cream", Points: 10,
		CategoryID: "cat1", ChildName: "Alice", ObservedAt: time.Date(2026, 1, 1, 15, 4, 5, 0, time.UTC),
	}

	t.Run("text format", func(t *testing.T) {
		t.Cleanup(func() { outputFormat = "" })
		outputFormat = ""
		out := captureStdout(func() { printRedemptionEvent(e) })
		if !strings.Contains(out, "Ice cream") || !strings.Contains(out, "Alice") {
			t.Errorf("expected reward name and child name in output, got: %s", out)
		}
	})

	t.Run("json format", func(t *testing.T) {
		t.Cleanup(func() { outputFormat = "" })
		outputFormat = outputJSON
		out := captureStdout(func() { printRedemptionEvent(e) })
		if !strings.Contains(out, `"id": "r1"`) {
			t.Errorf("expected reward id in JSON output, got: %s", out)
		}
	})

	t.Run("falls back to category when child name missing", func(t *testing.T) {
		t.Cleanup(func() { outputFormat = "" })
		outputFormat = ""
		noChild := e
		noChild.ChildName = ""
		out := captureStdout(func() { printRedemptionEvent(noChild) })
		if !strings.Contains(out, "cat1") {
			t.Errorf("expected category id fallback in output, got: %s", out)
		}
	})
}

// newWatchTestClient builds a mock client and sets the package-level frameID
// that pollRewards/pollChores/pollCalendar read directly.
func newWatchTestClient(t *testing.T, handler http.HandlerFunc) *lib.Client {
	t.Helper()
	client := newMockClient(t, handler)

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	return client
}

func TestPollRewards(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[
			{"id":"1","attributes":{"name":"Ice cream","point_value":10,"redeemed_at":"2026-01-01T00:00:00Z"}},
			{"id":"2","attributes":{"name":"Movie night","point_value":20}}
		]}`)
	})

	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	out := captureStdout(func() { pollRewards(client, state, "12:00:00") })

	if !strings.Contains(out, "Ice cream") {
		t.Errorf("expected redeemed reward in output, got: %s", out)
	}
	if strings.Contains(out, "Movie night") {
		t.Errorf("unredeemed reward should not be printed, got: %s", out)
	}
	if _, seen := state.seenRewardIDs["1"]; !seen {
		t.Error("expected reward 1 to be marked seen")
	}
	if _, seen := state.seenRewardIDs["2"]; seen {
		t.Error("unredeemed reward should not be marked seen")
	}

	// Polling again should not re-print the already-seen redemption.
	out = captureStdout(func() { pollRewards(client, state, "12:00:01") })
	if strings.Contains(out, "Ice cream") {
		t.Errorf("already-seen redemption should not be re-printed, got: %s", out)
	}
}

func TestPollRewards_Seeding(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"1","attributes":{"name":"Ice cream","point_value":10,"redeemed_at":"2026-01-01T00:00:00Z"}}]}`)
	})

	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
		seeding:       true,
	}

	out := captureStdout(func() { pollRewards(client, state, "12:00:00") })
	if out != "" {
		t.Errorf("expected no output while seeding, got: %s", out)
	}
	if _, seen := state.seenRewardIDs["1"]; !seen {
		t.Error("expected reward to be marked seen even while seeding")
	}
}

func TestPollRewards_ListError(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	state := &watchState{seenRewardIDs: make(map[string]struct{})}
	// Errors go to stderr, not stdout; just confirm it doesn't panic.
	pollRewards(client, state, "12:00:00")
}

func TestPollChores(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Clean room","status":"complete"}}]}`)
	})

	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	out := captureStdout(func() { pollChores(client, state, time.Now(), "12:00:00") })

	if !strings.Contains(out, "Clean room") {
		t.Errorf("expected completed chore in output, got: %s", out)
	}
	if _, seen := state.seenChoreIDs["c1"]; !seen {
		t.Error("expected chore c1 to be marked seen")
	}

	// Already-seen chores should not be re-printed.
	out = captureStdout(func() { pollChores(client, state, time.Now(), "12:00:01") })
	if strings.Contains(out, "Clean room") {
		t.Errorf("already-seen chore should not be re-printed, got: %s", out)
	}
}

func TestPollChores_JSONOutput(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Clean room","status":"complete"}}]}`)
	})

	state := &watchState{seenChoreIDs: make(map[string]struct{})}
	out := captureStdout(func() { pollChores(client, state, time.Now(), "12:00:00") })
	if !strings.Contains(out, `"id": "c1"`) {
		t.Errorf("expected chore id in JSON output, got: %s", out)
	}
}

func TestPollChores_ListError(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	state := &watchState{seenChoreIDs: make(map[string]struct{})}
	pollChores(client, state, time.Now(), "12:00:00")
}

func TestPollCalendar(t *testing.T) {
	soon := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	tooFar := time.Now().Add(2 * time.Hour).Format(time.RFC3339)

	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":[
			{"id":"e1","type":"calendar_event","attributes":{"summary":"Soon","starts_at":%q,"all_day":false},"relationships":{"categories":{"data":[]}}},
			{"id":"e2","type":"calendar_event","attributes":{"summary":"Later","starts_at":%q,"all_day":false},"relationships":{"categories":{"data":[]}}},
			{"id":"e3","type":"calendar_event","attributes":{"summary":"AllDay","starts_at":%q,"all_day":true},"relationships":{"categories":{"data":[]}}}
		]}`, soon, tooFar, soon)
	})

	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	out := captureStdout(func() { pollCalendar(client, state, time.Now(), "12:00:00") })

	if !strings.Contains(out, "Soon") {
		t.Errorf("expected event starting within the hour in output, got: %s", out)
	}
	if strings.Contains(out, "Later") {
		t.Errorf("event more than an hour away should not be printed, got: %s", out)
	}
	if strings.Contains(out, "AllDay") {
		t.Errorf("all-day event should not be printed, got: %s", out)
	}
	if _, seen := state.seenEventIDs["e1"]; !seen {
		t.Error("expected e1 to be marked seen")
	}
	if _, seen := state.seenEventIDs["e2"]; seen {
		t.Error("e2 (too far out) should not be marked seen")
	}

	// Already-seen events should not be re-printed.
	out = captureStdout(func() { pollCalendar(client, state, time.Now(), "12:00:01") })
	if strings.Contains(out, "Soon") {
		t.Errorf("already-seen event should not be re-printed, got: %s", out)
	}
}

func TestPollCalendar_JSONOutput(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	soon := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Soon","starts_at":%q,"all_day":false},"relationships":{"categories":{"data":[]}}}]}`, soon)
	})

	state := &watchState{seenEventIDs: make(map[string]struct{})}
	out := captureStdout(func() { pollCalendar(client, state, time.Now(), "12:00:00") })
	if !strings.Contains(out, `"id": "e1"`) {
		t.Errorf("expected event id in JSON output, got: %s", out)
	}
}

func TestPollCalendar_SkipsUnparsableStartTime(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Bad","starts_at":"not-a-date","all_day":false},"relationships":{"categories":{"data":[]}}}]}`)
	})

	state := &watchState{seenEventIDs: make(map[string]struct{})}
	out := captureStdout(func() { pollCalendar(client, state, time.Now(), "12:00:00") })
	if out != "" {
		t.Errorf("expected no output for unparsable start time, got: %s", out)
	}
	if _, seen := state.seenEventIDs["e1"]; seen {
		t.Error("unparsable event should not be marked seen")
	}
}

func TestPollCalendar_ListError(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	state := &watchState{seenEventIDs: make(map[string]struct{})}
	pollCalendar(client, state, time.Now(), "12:00:00")
}

func TestPollLists(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"l1","attributes":{"label":"Groceries"}}]}`)
	})

	state := &watchState{seenListIDs: make(map[string]struct{})}

	out := captureStdout(func() { pollLists(client, state, "12:00:00") })
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected new list in output, got: %s", out)
	}
	if _, seen := state.seenListIDs["l1"]; !seen {
		t.Error("expected list l1 to be marked seen")
	}

	// Already-seen lists should not be re-printed.
	out = captureStdout(func() { pollLists(client, state, "12:00:01") })
	if strings.Contains(out, "Groceries") {
		t.Errorf("already-seen list should not be re-printed, got: %s", out)
	}
}

func TestPollLists_JSONOutput(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"l1","attributes":{"label":"Groceries"}}]}`)
	})

	state := &watchState{seenListIDs: make(map[string]struct{})}
	out := captureStdout(func() { pollLists(client, state, "12:00:00") })
	if !strings.Contains(out, `"id": "l1"`) {
		t.Errorf("expected list id in JSON output, got: %s", out)
	}
}

func TestPollLists_ListError(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	state := &watchState{seenListIDs: make(map[string]struct{})}
	pollLists(client, state, "12:00:00")
}

func TestPollRoutines(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"title":"Morning Routine"}}]}`)
	})

	state := &watchState{seenRoutineIDs: make(map[string]struct{})}

	out := captureStdout(func() { pollRoutines(client, state, "12:00:00") })
	if !strings.Contains(out, "Morning Routine") {
		t.Errorf("expected new routine in output, got: %s", out)
	}
	if _, seen := state.seenRoutineIDs["r1"]; !seen {
		t.Error("expected routine r1 to be marked seen")
	}

	// Already-seen routines should not be re-printed.
	out = captureStdout(func() { pollRoutines(client, state, "12:00:01") })
	if strings.Contains(out, "Morning Routine") {
		t.Errorf("already-seen routine should not be re-printed, got: %s", out)
	}
}

func TestPollRoutines_JSONOutput(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON

	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"title":"Morning Routine"}}]}`)
	})

	state := &watchState{seenRoutineIDs: make(map[string]struct{})}
	out := captureStdout(func() { pollRoutines(client, state, "12:00:00") })
	if !strings.Contains(out, `"id": "r1"`) {
		t.Errorf("expected routine id in JSON output, got: %s", out)
	}
}

func TestPollRoutines_ListError(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	state := &watchState{seenRoutineIDs: make(map[string]struct{})}
	pollRoutines(client, state, "12:00:00")
}

func TestPoll_AllResources(t *testing.T) {
	client := newWatchTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/routines"):
			fmt.Fprint(w, `{"data":[]}`)
		default:
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
	})

	state := &watchState{
		seenRewardIDs:  make(map[string]struct{}),
		seenChoreIDs:   make(map[string]struct{}),
		seenEventIDs:   make(map[string]struct{}),
		seenListIDs:    make(map[string]struct{}),
		seenRoutineIDs: make(map[string]struct{}),
	}

	// poll() fans the resources out over goroutines; a clean run with no
	// panics and no unexpected requests is the behavior under test.
	poll(client, state, allWatchResources)
}
