package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

// newImportTestClient builds a mock client that serves canned success
// responses for every Create*/AddListItem endpoint import.go calls, except
// for the paths listed in failPaths, which return a 500.
func newImportTestClient(t *testing.T, failPaths map[string]bool) *lib.Client {
	t.Helper()

	responses := map[string]string{
		"/rewards":         `{"data":[{"id":"r1","attributes":{"name":"Reward","point_value":5}}]}`,
		"/chores":          `{"data":{"id":"c1","attributes":{"summary":"Chore"}}}`,
		"/lists":           `{"data":{"id":"l1","type":"list","attributes":{"label":"List","color":"#2178AF","kind":"to_do"}}}`,
		"/list_items":      `{"data":{"id":"i1","type":"list_item","attributes":{"label":"Item","status":"pending","position":0}}}`,
		"/meals/recipes":   `{"data":{"id":"rc1","type":"meal_recipe","attributes":{"summary":"Recipe","description":""}}}`,
		"/meals/sittings":  `{"data":[{"id":"s1","type":"meal_sitting","attributes":{"summary":"dinner"}}]}`,
		"/calendar_events": `{"data":{"id":"e1","type":"calendar_event","attributes":{"summary":"Event","starts_at":"","ends_at":"","all_day":false},"relationships":{"categories":{"data":[]}}}}`,
	}

	return newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		for suffix, fail := range failPaths {
			if fail && strings.HasSuffix(r.URL.Path, suffix) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		for suffix, body := range responses {
			if strings.HasSuffix(r.URL.Path, suffix) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(body))
				return
			}
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})
}

func TestParallelImport_PanicRecovery(t *testing.T) {
	total, failed := parallelImport([]string{"item"}, func(string) (int, int) {
		panic("simulated goroutine panic")
	})
	if total != 0 || failed != 1 {
		t.Errorf("got total=%d failed=%d, want total=0 failed=1", total, failed)
	}
}

func TestParallelImport_MixedPanicAndSuccess(t *testing.T) {
	items := []string{"ok1", "panic", "ok2"}
	total, failed := parallelImport(items, func(s string) (int, int) {
		if s == "panic" {
			panic("simulated goroutine panic")
		}
		return 1, 0
	})
	// Panicked items contribute (0,1) — not (1,1) — since they never counted
	// the attempt. So 2 successful items give total=2; 1 panic gives failed=1.
	if total != 2 || failed != 1 {
		t.Errorf("got total=%d failed=%d, want total=2 failed=1", total, failed)
	}
}

func TestImportRewards(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importRewards(context.Background(), client, []lib.Reward{{Title: "A"}, {Title: "B"}})
		if total != 2 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=2 failed=0", total, failed)
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/rewards": true})
		total, failed := importRewards(context.Background(), client, []lib.Reward{{Title: "A"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
		}
	})
}

func TestImportChores(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importChores(context.Background(), client, []lib.Chore{{Title: "Walk dog"}, {Title: "Dishes"}})
		if total != 2 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=2 failed=0", total, failed)
		}
	})

	t.Run("failure counted", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/chores": true})
		total, failed := importChores(context.Background(), client, []lib.Chore{{Title: "Walk dog"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
		}
	})

	// A routine is a chore, so ListChores and ListRoutines both return it at
	// export time -- importChores must skip routine chores so importRoutines
	// (which handles them separately) doesn't end up creating each routine
	// twice: once as a plain chore, once as a proper routine.
	t.Run("skips routine chores", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importChores(context.Background(), client, []lib.Chore{
			{Title: "Walk dog"},
			{Title: "Make bed", Routine: true},
		})
		if total != 1 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=0 (routine chore skipped)", total, failed)
		}
	})

	// A skipped routine chore is invisible in total/failed by design (it's
	// not this function's job to import it), but that means it can vanish
	// silently if the caller didn't also request the "routines" resource --
	// warn on stderr so the skip is at least visible.
	t.Run("warns on stderr when skipping a routine chore", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		stderr := captureStderr(func() {
			importChores(context.Background(), client, []lib.Chore{{Title: "Make bed", Routine: true}})
		})
		if !strings.Contains(stderr, "Make bed") {
			t.Errorf("expected warning naming the skipped routine chore, got: %s", stderr)
		}
	})
}

func TestImportLists(t *testing.T) {
	t.Run("list and items succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		lists := []lib.List{{
			Title: "Groceries",
			Items: []lib.ListItem{{Title: "Eggs"}, {Title: "Milk"}},
		}}
		total, failed := importLists(context.Background(), client, lists)
		if total != 3 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=3 (1 list + 2 items) failed=0", total, failed)
		}
	})

	t.Run("list creation failure skips its items", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/lists": true})
		lists := []lib.List{{
			Title: "Groceries",
			Items: []lib.ListItem{{Title: "Eggs"}},
		}}
		total, failed := importLists(context.Background(), client, lists)
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1 (item creation skipped)", total, failed)
		}
	})

	t.Run("item failure counted separately from list", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/list_items": true})
		lists := []lib.List{{
			Title: "Groceries",
			Items: []lib.ListItem{{Title: "Eggs"}},
		}}
		total, failed := importLists(context.Background(), client, lists)
		if total != 2 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=2 failed=1", total, failed)
		}
	})
}

func TestImportRecipes(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importRecipes(context.Background(), client, []lib.Recipe{{Title: "Tacos"}})
		if total != 1 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=0", total, failed)
		}
	})

	t.Run("failure counted", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/meals/recipes": true})
		total, failed := importRecipes(context.Background(), client, []lib.Recipe{{Title: "Tacos"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
		}
	})
}

func TestImportSittings(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importSittings(context.Background(), client, []lib.MealSitting{{Summary: "dinner"}})
		if total != 1 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=0", total, failed)
		}
	})

	t.Run("failure counted", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/meals/sittings": true})
		total, failed := importSittings(context.Background(), client, []lib.MealSitting{{Summary: "dinner"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
		}
	})
}

func TestImportCalendarEvents(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importCalendarEvents(context.Background(), client, []lib.CalendarEvent{{Title: "Birthday"}})
		if total != 1 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=0", total, failed)
		}
	})

	t.Run("failure counted", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/calendar_events": true})
		total, failed := importCalendarEvents(context.Background(), client, []lib.CalendarEvent{{Title: "Birthday"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
		}
	})
}

func TestRunImport_AllSuccess(t *testing.T) {
	client := newImportTestClient(t, nil)
	data := ExportData{
		Rewards:        []lib.Reward{{Title: "Reward"}},
		Chores:         []lib.Chore{{Title: "Chore"}},
		Lists:          []lib.List{{Title: "List"}},
		Recipes:        []lib.Recipe{{Title: "Recipe"}},
		MealSittings:   []lib.MealSitting{{Summary: "dinner"}},
		CalendarEvents: []lib.CalendarEvent{{Title: "Event"}},
	}
	want := map[string]bool{
		exportResourceRewards:  true,
		exportResourceChores:   true,
		exportResourceLists:    true,
		exportResourceRecipes:  true,
		exportResourceSittings: true,
		exportResourceCalendar: true,
	}

	var err error
	out := captureStdout(func() { err = runImport(context.Background(), client, data, want) })

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(out, "Imported 6/6 items successfully.") {
		t.Errorf("expected success summary in output, got: %s", out)
	}
}

func TestRunImport_OnlyRequestedResourcesAreImported(t *testing.T) {
	client := newImportTestClient(t, nil)
	data := ExportData{
		Rewards: []lib.Reward{{Title: "Reward"}},
		Chores:  []lib.Chore{{Title: "Chore"}},
	}
	want := map[string]bool{exportResourceRewards: true}

	var err error
	out := captureStdout(func() { err = runImport(context.Background(), client, data, want) })

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(out, "Imported 1/1 items successfully.") {
		t.Errorf("expected only the reward to be imported, got: %s", out)
	}
}

func TestRunImport_ReturnsErrorOnPartialFailure(t *testing.T) {
	client := newImportTestClient(t, map[string]bool{"/rewards": true})
	data := ExportData{
		Rewards: []lib.Reward{{Title: "Reward"}},
		Chores:  []lib.Chore{{Title: "Chore"}},
	}
	want := map[string]bool{
		exportResourceRewards: true,
		exportResourceChores:  true,
	}

	var err error
	_ = captureStdout(func() { err = runImport(context.Background(), client, data, want) })

	if err == nil {
		t.Fatal("expected error when some items fail, got nil")
	}
	if !strings.Contains(err.Error(), "failed to import") {
		t.Errorf("expected 'failed to import' in error, got: %v", err)
	}
}

func TestRunImport_EmptyWant(t *testing.T) {
	client := newImportTestClient(t, nil)
	data := ExportData{Rewards: []lib.Reward{{Title: "Reward"}}}

	var err error
	out := captureStdout(func() { err = runImport(context.Background(), client, data, map[string]bool{}) })

	if err != nil {
		t.Fatalf("expected no error for empty want, got: %v", err)
	}
	if !strings.Contains(out, "Imported 0/0 items") {
		t.Errorf("expected '0/0 items' in output, got: %s", out)
	}
}

func writeImportFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "export.json")
	data := []byte(`{"frame_id":"test-frame","rewards":[{"title":"Reward"}],"chores":[{"title":"Chore"}]}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	return path
}

func TestImportCmd_Run(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[{"id":"x1","attributes":{"name":"X","point_value":1}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Chore"}}}`)
		}
	})

	origFile, origResources, origDryRun := importFile, importResources, importDryRun
	importFile = writeImportFixture(t)
	importResources = "all"
	importDryRun = false
	t.Cleanup(func() { importFile, importResources, importDryRun = origFile, origResources, origDryRun })

	out := captureStdout(func() {
		if err := importCmd.RunE(importCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Imported") {
		t.Errorf("expected import summary in output, got: %s", out)
	}
}

func TestImportCmd_DryRun(t *testing.T) {
	origFile, origResources, origDryRun := importFile, importResources, importDryRun
	importFile = writeImportFixture(t)
	importResources = "all"
	importDryRun = true
	t.Cleanup(func() { importFile, importResources, importDryRun = origFile, origResources, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := importCmd.RunE(importCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestImportCmd_FileNotFound(t *testing.T) {
	origFrameID, origFile := frameID, importFile
	frameID = "test-frame"
	importFile = "/nonexistent/path/export.json"
	t.Cleanup(func() { frameID, importFile = origFrameID, origFile })

	err := importCmd.RunE(importCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing import file, got nil")
	}
	if !strings.Contains(err.Error(), "reading") {
		t.Errorf("expected 'reading' in error, got: %v", err)
	}
}
