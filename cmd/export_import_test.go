package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		got := parseExportResources(input)
		if len(got) != len(allExportResources) {
			t.Errorf("parseExportResources(%q): expected %d resources, got %d", input, len(allExportResources), len(got))
		}
	}
}

func TestParseExportResources_Specific(t *testing.T) {
	got := parseExportResources("chores,rewards")
	if len(got) != 2 || got[0] != exportResourceChores || got[1] != exportResourceRewards {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestParseExportResources_TrimsSpaces(t *testing.T) {
	got := parseExportResources(" chores , rewards ")
	if len(got) != 2 || got[0] != exportResourceChores || got[1] != exportResourceRewards {
		t.Errorf("unexpected result: %v", got)
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
	runImportDryRun(data, want)

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
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"Ice cream","point_value":5}}]}`)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[{"id":"l1","type":"list","attributes":{"label":"Groceries"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/meals/recipes"):
			fmt.Fprint(w, `{"data":[{"id":"rc1","type":"meal_recipe","attributes":{"summary":"Tacos"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/meals/sittings"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[{"id":"e1","type":"calendar_event","attributes":{"summary":"Meeting","starts_at":"2026-01-01T10:00:00Z","all_day":false},"relationships":{"categories":{"data":[]}}}]}`)
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

	out := captureStdout(func() { exportCmd.Run(exportCmd, nil) })

	var data ExportData
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("expected valid JSON on stdout, got error %v for: %s", err, out)
	}
	if data.FrameID != "test-frame" {
		t.Errorf("expected frame_id test-frame, got %q", data.FrameID)
	}
	if len(data.Chores) != 1 || len(data.Rewards) != 1 || len(data.Lists) != 1 || len(data.Recipes) != 1 || len(data.CalendarEvents) != 1 {
		t.Errorf("expected one of each resource, got: %+v", data)
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

	out := captureStdout(func() { exportCmd.Run(exportCmd, nil) })

	var data ExportData
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("expected valid JSON on stdout, got error %v for: %s", err, out)
	}
	if len(data.Chores) != 1 {
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

	out := captureStdout(func() { exportCmd.Run(exportCmd, nil) })
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
	if len(data.Chores) != 1 {
		t.Errorf("expected chores in exported file, got: %+v", data.Chores)
	}
}
