package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

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
