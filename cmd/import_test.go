package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

func TestImportRewards(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importRewards(client, []lib.Reward{{Title: "A"}, {Title: "B"}})
		if total != 2 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=2 failed=0", total, failed)
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/rewards": true})
		total, failed := importRewards(client, []lib.Reward{{Title: "A"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
		}
	})
}

func TestImportChores(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importChores(client, []lib.Chore{{Title: "Walk dog"}, {Title: "Dishes"}})
		if total != 2 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=2 failed=0", total, failed)
		}
	})

	t.Run("failure counted", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/chores": true})
		total, failed := importChores(client, []lib.Chore{{Title: "Walk dog"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
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
		total, failed := importLists(client, lists)
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
		total, failed := importLists(client, lists)
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
		total, failed := importLists(client, lists)
		if total != 2 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=2 failed=1", total, failed)
		}
	})
}

func TestImportRecipes(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importRecipes(client, []lib.Recipe{{Title: "Tacos"}})
		if total != 1 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=0", total, failed)
		}
	})

	t.Run("failure counted", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/meals/recipes": true})
		total, failed := importRecipes(client, []lib.Recipe{{Title: "Tacos"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
		}
	})
}

func TestImportSittings(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importSittings(client, []lib.MealSitting{{Summary: "dinner"}})
		if total != 1 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=0", total, failed)
		}
	})

	t.Run("failure counted", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/meals/sittings": true})
		total, failed := importSittings(client, []lib.MealSitting{{Summary: "dinner"}})
		if total != 1 || failed != 1 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=1", total, failed)
		}
	})
}

func TestImportCalendarEvents(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		client := newImportTestClient(t, nil)
		total, failed := importCalendarEvents(client, []lib.CalendarEvent{{Title: "Birthday"}})
		if total != 1 || failed != 0 {
			t.Errorf("got total=%d failed=%d, want total=1 failed=0", total, failed)
		}
	})

	t.Run("failure counted", func(t *testing.T) {
		client := newImportTestClient(t, map[string]bool{"/calendar_events": true})
		total, failed := importCalendarEvents(client, []lib.CalendarEvent{{Title: "Birthday"}})
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

	out := captureStdout(func() { runImport(client, data, want) })

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

	out := captureStdout(func() { runImport(client, data, want) })

	if !strings.Contains(out, "Imported 1/1 items successfully.") {
		t.Errorf("expected only the reward to be imported, got: %s", out)
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

	out := captureStdout(func() { importCmd.Run(importCmd, nil) })
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

	out := captureStdout(func() { importCmd.Run(importCmd, nil) })
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestImportCmd_FileNotFound_Crasher(t *testing.T) {
	if os.Getenv("WANT_IMPORT_FILE_CRASH") != "1" {
		t.Skip("only runs as a subprocess of TestImportCmd_FileNotFound")
	}
	frameID = "test-frame"
	importFile = "/nonexistent/path/export.json"
	importCmd.Run(importCmd, nil)
}

func TestImportCmd_FileNotFound(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestImportCmd_FileNotFound_Crasher") //nolint:gosec // test binary, fixed flag
	cmd.Env = append(os.Environ(), "WANT_IMPORT_FILE_CRASH=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.Success() {
		t.Fatalf("expected importCmd.Run to exit with a non-zero status, got err=%v", err)
	}
	if !strings.Contains(stderr.String(), "Error reading") {
		t.Errorf("expected file-read error on stderr, got: %s", stderr.String())
	}
}
