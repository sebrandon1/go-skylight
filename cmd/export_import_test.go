package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestToWantMap(t *testing.T) {
	got := toWantMap([]string{exportResourceChores, exportResourceRewards})
	if !got[exportResourceChores] || !got[exportResourceRewards] {
		t.Errorf("expected chores and rewards to be wanted, got: %v", got)
	}
	if got[exportResourceLists] {
		t.Errorf("expected lists not to be wanted, got: %v", got)
	}
}

func TestParseExportResources_All(t *testing.T) {
	for _, input := range []string{"", "all"} {
		got, err := parseExportResources(input)
		if err != nil {
			t.Fatalf("parseExportResources(%q): unexpected error: %v", input, err)
		}
		if len(got) != len(allExportResources) {
			t.Errorf("parseExportResources(%q): expected %d resources, got %d", input, len(allExportResources), len(got))
		}
	}
}

func TestParseExportResources_Specific(t *testing.T) {
	got, err := parseExportResources("chores,rewards")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != exportResourceChores || got[1] != exportResourceRewards {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestParseExportResources_TrimsSpaces(t *testing.T) {
	got, err := parseExportResources(" chores , rewards ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != exportResourceChores || got[1] != exportResourceRewards {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestParseExportResources_SkipsUnknown(t *testing.T) {
	// #266: unknown tokens are warned and dropped; valid ones remain
	var got []string
	stderr := captureStderr(func() {
		var err error
		got, err = parseExportResources("chores,bogus,rewards")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if len(got) != 2 || got[0] != exportResourceChores || got[1] != exportResourceRewards {
		t.Errorf("unexpected result: %v", got)
	}
	if !strings.Contains(stderr, "bogus") {
		t.Errorf("expected warning for unknown resource, got: %s", stderr)
	}
}

func TestExportCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "export" {
			found = true
			break
		}
	}
	if !found {
		t.Error("export command not registered on root")
	}
}

func TestImportCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "import" {
			found = true
			break
		}
	}
	if !found {
		t.Error("import command not registered on root")
	}
}

func TestExportCmdHasFlags(t *testing.T) {
	flags := []string{"output-file", "resources", "days"}
	for _, f := range flags {
		if exportCmd.Flags().Lookup(f) == nil {
			t.Errorf("expected --%s flag on export command", f)
		}
	}
}

func TestImportCmdHasFlags(t *testing.T) {
	flags := []string{"file", "dry-run", "resources"}
	for _, f := range flags {
		if importCmd.Flags().Lookup(f) == nil {
			t.Errorf("expected --%s flag on import command", f)
		}
	}
}

func TestRunImportDryRun_PrintsCounts(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	data := ExportData{
		Chores:  []lib.Chore{{ID: "1"}, {ID: "2"}},
		Rewards: []lib.Reward{{ID: "r1"}},
	}
	want := map[string]bool{
		exportResourceChores:  true,
		exportResourceRewards: true,
	}
	runImportDryRun(data, want, importTestToday)

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "2 items") {
		t.Errorf("expected chores count 2 in output, got: %s", out)
	}
	if !strings.Contains(out, "1 items") {
		t.Errorf("expected rewards count 1 in output, got: %s", out)
	}
}

// TestRunImportDryRun_ExcludesRoutineChoresFromCount ensures the dry-run
// preview matches what a real import actually does: importChores skips
// Routine==true chores (they're imported separately via importRoutines), so
// the chores count in the preview must exclude them too, or the preview
// would overstate how many plain chores will be created.
func TestRunImportDryRun_ExcludesRoutineChoresFromCount(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	data := ExportData{
		Chores: []lib.Chore{{ID: "1"}, {ID: "2", Routine: true}, {ID: "3", Routine: true}},
	}
	want := map[string]bool{exportResourceChores: true}
	runImportDryRun(data, want, importTestToday)

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "1 items") {
		t.Errorf("expected chores count of 1 (routine chores excluded), got: %s", out)
	}
	if strings.Contains(out, "3 items") {
		t.Errorf("expected routine chores excluded from count, got: %s", out)
	}
}

func TestImportCmd_DryRunChoreCount(t *testing.T) {
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			// Written before start_time/emoji_icon were exported.
			name: "old export file restores each series once",
			file: `{"exported_at":"2026-09-01T08:00:00+10:00","frame_id":"test-frame","chores":[
				{"id":"100-2026-08-27","title":"Bins","status":"complete","due_date":"2026-08-27","recurring":true,"assignee_id":"cat1","recurrence_set":["RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TH"]},
				{"id":"100-2026-09-03","title":"Bins","status":"pending","due_date":"2026-09-03","recurring":true,"assignee_id":"cat1","recurrence_set":["RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TH"]},
				{"id":"100-2026-09-10","title":"Bins","status":"pending","due_date":"2026-09-10","recurring":true,"assignee_id":"cat1","recurrence_set":["RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TH"]},
				{"id":"701","title":"Return library books","status":"complete","due_date":"2026-08-20","recurring":false,"assignee_id":"cat2"}
			]}`,
			want: "chores     1 items",
		},
		{
			// exported_at is ahead of this machine's clock, and its date in its
			// own offset is a day after its UTC date.
			name: "series that ended before the export date is skipped",
			file: `{"exported_at":"2099-01-02T08:00:00+10:00","frame_id":"test-frame","chores":[
				{"id":"400-2099-01-01","title":"Old schedule","status":"complete","due_date":"2099-01-01","recurring":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;UNTIL=20990101"]},
				{"id":"100-2099-01-07","title":"Bins","status":"pending","due_date":"2099-01-07","recurring":true,"recurrence_set":["RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TH"]}
			]}`,
			want: "chores     1 items",
		},
		{
			// A restore made long after the export uses the local date.
			name: "series that ended between the export date and today is skipped",
			file: `{"exported_at":"2020-01-01T08:00:00+10:00","frame_id":"test-frame","chores":[
				{"id":"400-2020-01-01","title":"Old schedule","status":"pending","due_date":"2020-01-01","recurring":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;UNTIL=20200601"]},
				{"id":"100-2020-01-02","title":"Bins","status":"pending","due_date":"2020-01-02","recurring":true,"recurrence_set":["RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TH"]}
			]}`,
			want: "chores     1 items",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "export.json")
			if err := os.WriteFile(file, []byte(tc.file), 0o600); err != nil {
				t.Fatalf("writing export: %v", err)
			}

			origFile, origResources, origDryRun, origFrameID := importFile, importResources, importDryRun, frameID
			importFile, importResources, importDryRun, frameID = file, "chores", true, "test-frame"
			t.Cleanup(func() {
				importFile, importResources, importDryRun, frameID = origFile, origResources, origDryRun, origFrameID
			})

			out := captureStdout(func() {
				if err := importCmd.RunE(importCmd, nil); err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q, got: %s", tc.want, out)
			}
		})
	}
}

func TestRunImportDryRun_ExcludesCategories(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	data := ExportData{
		Routines:   []lib.Routine{{Title: "Morning"}},
		Bounties:   []lib.Bounty{{}},
		Categories: []lib.Category{{ID: "cat1"}},
	}
	want := map[string]bool{
		exportResourceRoutines:   true,
		exportResourceBounties:   true,
		exportResourceCategories: true,
	}
	runImportDryRun(data, want, importTestToday)

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "routines") {
		t.Errorf("expected routines in dry-run output, got: %s", out)
	}
	if !strings.Contains(out, "bounties") {
		t.Errorf("expected bounties in dry-run output, got: %s", out)
	}
	if strings.Contains(out, "categories") {
		t.Errorf("expected categories excluded from dry-run output (export-only), got: %s", out)
	}
}

func TestExportDataRoundTrip(t *testing.T) {
	data := ExportData{
		ExportedAt: "2026-05-02T00:00:00Z",
		FrameID:    "frame-abc",
		Chores:     []lib.Chore{{ID: "1", Title: "Walk the dog"}},
		Rewards:    []lib.Reward{{ID: "r1", Title: "Ice cream"}},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "export.json")

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var got ExportData
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.FrameID != data.FrameID {
		t.Errorf("frame_id: got %q, want %q", got.FrameID, data.FrameID)
	}
	if len(got.Chores) != 1 || got.Chores[0].Title != "Walk the dog" {
		t.Errorf("chores mismatch: %+v", got.Chores)
	}
}

func exportMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		// A routine is a chore, so both ListChores (exported as "chores")
		// and ListRoutines (exported as "routines") hit this same endpoint;
		// the mock returns one plain chore and one routine chore, and each
		// caller filters/dedupes it differently.
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[
				{"id":"c1","attributes":{"summary":"Dishes","start_time":"19:00","emoji_icon":"🍽️"}},
				{"id":"rt1","attributes":{"summary":"Morning Routine","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=6"]}}
			]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"Ice cream","point_value":5,"description":"Two scoops","respawn_on_redemption":true}}]}`)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[{"id":"l1","type":"list","attributes":{"label":"Groceries"}}],"included":[{"id":"i1","type":"list_item","attributes":{"label":"Milk","status":"completed","section":"Dairy","position":1},"relationships":{"list":{"data":{"id":"l1","type":"list"}}}}]}`)
		case strings.HasSuffix(r.URL.Path, "/meals/recipes"):
			fmt.Fprint(w, `{"data":[{"id":"rc1","type":"meal_recipe","attributes":{"summary":"Tacos"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/meals/sittings"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}]}`)
		case strings.HasSuffix(r.URL.Path, "/categories"):
			fmt.Fprint(w, `{"data":[{"id":"cat1","type":"category","attributes":{"label":"Alice","color":"blue"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/messages"):
			fmt.Fprint(w, `{"data":[],"meta":{"next_page_token":""}}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	}
}

func TestExportCmd_AllResourcesToStdout(t *testing.T) {
	newCmdTestClient(t, exportMockHandler())

	origFile, origResources, origDays := exportOutputFile, exportResources, exportDays
	exportOutputFile = ""
	exportResources = "all"
	exportDays = 7
	t.Cleanup(func() {
		exportOutputFile, exportResources, exportDays = origFile, origResources, origDays
	})

	out := captureStdout(func() {
		if err := exportCmd.RunE(exportCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var data ExportData
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("expected valid JSON on stdout, got error %v for: %s", err, out)
	}
	if data.FrameID != "test-frame" {
		t.Errorf("expected frame_id test-frame, got %q", data.FrameID)
	}
	// data.Chores includes 2: a routine is a chore, so the plain chore list
	// and the routine list overlap on the routine chore by design.
	if len(data.Chores) != 2 || len(data.Rewards) != 1 || len(data.Lists) != 1 || len(data.Recipes) != 1 || len(data.CalendarEvents) != 1 {
		t.Errorf("expected one of each legacy resource (chores=2), got: %+v", data)
	}
	if len(data.Chores) > 0 && (data.Chores[0].StartTime != "19:00" || data.Chores[0].EmojiIcon != "🍽️") {
		t.Errorf("expected chore start_time and emoji_icon exported, got: %+v", data.Chores[0])
	}
	if len(data.Lists) > 0 && (len(data.Lists[0].Items) != 1 || data.Lists[0].Items[0].Section != "Dairy" || !data.Lists[0].Items[0].Completed) {
		t.Errorf("expected the list's item exported with section and completed, got: %+v", data.Lists[0].Items)
	}
	if len(data.Rewards) > 0 && (data.Rewards[0].Description != "Two scoops" || !data.Rewards[0].RespawnOnRedemption) {
		t.Errorf("expected reward description and respawn_on_redemption exported, got: %+v", data.Rewards[0])
	}
	if len(data.Routines) != 1 {
		t.Errorf("expected 1 routine, got %d", len(data.Routines))
	}
	if len(data.Categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(data.Categories))
	}
}

// The plain chore query leaves out up-for-grabs chores, so export asks for
// them separately over the same window and keeps one row per ID.
func TestExportCmd_IncludesUpForGrabsChores(t *testing.T) {
	var plainQuery, grabsQuery url.Values
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/chores") && r.URL.Query().Get("include_up_for_grabs") == "true":
			grabsQuery = r.URL.Query()
			fmt.Fprint(w, `{"data":[
				{"id":"u1","attributes":{"summary":"Wash car","up_for_grabs":true}},
				{"id":"u2","attributes":{"summary":"Sweep porch","up_for_grabs":true}}
			]}`)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			plainQuery = r.URL.Query()
			fmt.Fprint(w, `{"data":[
				{"id":"c1","attributes":{"summary":"Dishes"}},
				{"id":"u2","attributes":{"summary":"Sweep porch","up_for_grabs":true}}
			]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	})

	origFile, origResources, origDays := exportOutputFile, exportResources, exportDays
	exportOutputFile, exportResources, exportDays = "", "chores", 7
	t.Cleanup(func() {
		exportOutputFile, exportResources, exportDays = origFile, origResources, origDays
	})

	out := captureStdout(func() {
		if err := exportCmd.RunE(exportCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var data ExportData
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("expected valid JSON on stdout, got error %v for: %s", err, out)
	}
	var ids []string
	for _, c := range data.Chores {
		ids = append(ids, c.ID)
	}
	if strings.Join(ids, ",") != "c1,u2,u1" {
		t.Errorf("expected chores c1,u2,u1, got %v", ids)
	}
	for _, k := range []string{"after", "before", "include_late"} {
		if grabsQuery.Get(k) == "" || grabsQuery.Get(k) != plainQuery.Get(k) {
			t.Errorf("expected the up-for-grabs query to match the plain one on %s, got %q vs %q", k, grabsQuery.Get(k), plainQuery.Get(k))
		}
	}
}

func TestExportCmd_ResourceFilter(t *testing.T) {
	newCmdTestClient(t, exportMockHandler())

	origFile, origResources, origDays := exportOutputFile, exportResources, exportDays
	exportOutputFile = ""
	exportResources = "chores"
	exportDays = 1
	t.Cleanup(func() {
		exportOutputFile, exportResources, exportDays = origFile, origResources, origDays
	})

	out := captureStdout(func() {
		if err := exportCmd.RunE(exportCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	var data ExportData
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("expected valid JSON on stdout, got error %v for: %s", err, out)
	}
	if len(data.Chores) != 2 {
		t.Errorf("expected chores included, got: %+v", data.Chores)
	}
	if len(data.Rewards) != 0 || len(data.Lists) != 0 {
		t.Errorf("expected only chores resource exported, got: %+v", data)
	}
}

func TestExportCmd_WritesToFile(t *testing.T) {
	newCmdTestClient(t, exportMockHandler())

	dir := t.TempDir()
	path := filepath.Join(dir, "export.json")

	origFile, origResources, origDays := exportOutputFile, exportResources, exportDays
	exportOutputFile = path
	exportResources = "chores"
	exportDays = 1
	t.Cleanup(func() {
		exportOutputFile, exportResources, exportDays = origFile, origResources, origDays
	})

	out := captureStdout(func() {
		if err := exportCmd.RunE(exportCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Exported to "+path) {
		t.Errorf("expected confirmation message, got: %s", out)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading exported file: %v", err)
	}
	var data ExportData
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("expected valid JSON in file, got error %v", err)
	}
	if len(data.Chores) != 2 {
		t.Errorf("expected chores in exported file, got: %+v", data.Chores)
	}
}
