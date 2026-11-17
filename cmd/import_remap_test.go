package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

// remapRecorder is a mock frame API that records every request as
// "METHOD path" with its body. The target frame has one category, Alice (ID
// 21).
type remapRecorder struct {
	mu   sync.Mutex
	reqs []string
}

func (r *remapRecorder) client(t *testing.T) *lib.Client {
	t.Helper()
	return newMockClient(t, func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		r.mu.Lock()
		r.reqs = append(r.reqs, req.Method+" "+req.URL.Path+" "+string(body))
		r.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		p := req.URL.Path
		switch {
		case strings.HasSuffix(p, "/categories") && req.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[{"id":"921","attributes":{"label":"Alice"}}]}`)
		case strings.HasSuffix(p, "/categories"):
			fmt.Fprint(w, `{"data":{"id":"922","attributes":{"label":"Bob"}}}`)
		case strings.HasSuffix(p, "/rewards"):
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"R","point_value":1}}]}`)
		case strings.HasSuffix(p, "/chores/create_multiple"):
			fmt.Fprint(w, `{"data":[{"id":"rt1","attributes":{"summary":"x"}}]}`)
		case strings.HasSuffix(p, "/chores"):
			fmt.Fprint(w, `{"data":{"id":"c1","attributes":{"summary":"Chore"}}}`)
		case strings.HasSuffix(p, "/meals/recipes"):
			fmt.Fprint(w, `{"data":{"id":"tgt-recipe","type":"meal_recipe","attributes":{"summary":"Tacos"}}}`)
		case strings.HasSuffix(p, "/meals/sittings"):
			fmt.Fprint(w, `{"data":[{"id":"s1","type":"meal_sitting","attributes":{"summary":"dinner"}}]}`)
		case strings.HasSuffix(p, "/calendar_events"):
			fmt.Fprint(w, `{"data":{"id":"e1","type":"calendar_event","attributes":{"summary":"E"},"relationships":{"categories":{"data":[]}}}}`)
		default:
			t.Errorf("unexpected request: %s %s", req.Method, p)
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

// posts returns the recorded non-GET requests whose path ends in suffix.
func (r *remapRecorder) posts(suffix string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for _, q := range r.reqs {
		if strings.HasPrefix(q, "GET ") {
			continue
		}
		if strings.HasSuffix(strings.SplitN(q, " ", 3)[1], suffix) {
			out = append(out, q)
		}
	}
	return out
}

func (r *remapRecorder) writeCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, q := range r.reqs {
		if !strings.HasPrefix(q, "GET ") {
			n++
		}
	}
	return n
}

func remapExport() ExportData {
	return ExportData{
		FrameID: "frame-a",
		Categories: []lib.Category{
			{ID: "811", Name: "Alice"},
			{ID: "812", Name: "Bob", Color: "#112233"},
		},
		Chores:         []lib.Chore{{ID: "1", Title: "Chore", AssigneeID: "811"}},
		Rewards:        []lib.Reward{{Title: "Reward", CategoryID: "811"}},
		Routines:       []lib.Routine{{Title: "Routine", TimeOfDay: lib.RoutineTODMorning, AssigneeID: "811", NextOccurrenceDate: importTestToday}},
		CalendarEvents: []lib.CalendarEvent{{Title: "Event", CategoryID: "811"}},
		Recipes:        []lib.Recipe{{ID: "src-recipe", Title: "Tacos"}},
		MealSittings:   []lib.MealSitting{{Summary: "dinner", RecipeID: "src-recipe", Date: importTestToday}},
	}
}

func setFrameID(t *testing.T, id string) {
	t.Helper()
	orig := frameID
	frameID = id
	t.Cleanup(func() { frameID = orig })
}

func wantSet(resources ...string) map[string]bool {
	m := map[string]bool{}
	for _, r := range resources {
		m[r] = true
	}
	return m
}

func requireContains(t *testing.T, what string, reqs []string, sub string) {
	t.Helper()
	if len(reqs) != 1 {
		t.Fatalf("%s: got %d requests, want 1", what, len(reqs))
	}
	if !strings.Contains(reqs[0], sub) {
		t.Errorf("%s: want %q in %s", what, sub, reqs[0])
	}
}

func TestRunImport_SameFrameKeepsIDs(t *testing.T) {
	setFrameID(t, "frame-a")
	rec := &remapRecorder{}
	data := remapExport()
	want := wantSet(exportResourceChores, exportResourceRewards, exportResourceCalendar, exportResourceRecipes, exportResourceSittings)

	var err error
	_ = captureStdout(func() { err = runImport(context.Background(), rec.client(t), data, want, importTestToday, false) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireContains(t, "chore", rec.posts("/chores"), "811")
	requireContains(t, "reward", rec.posts("/rewards"), "811")
	requireContains(t, "event", rec.posts("/calendar_events"), "811")
	requireContains(t, "sitting", rec.posts("/meals/sittings"), "src-recipe")
	for _, q := range rec.reqs {
		if strings.Contains(q, "/categories") {
			t.Errorf("same-frame import touched categories: %s", q)
		}
	}
}

func TestRunImport_CrossFrameRemapsByLabel(t *testing.T) {
	setFrameID(t, "frame-b")
	rec := &remapRecorder{}
	data := remapExport()
	data.Categories = data.Categories[:1] // only Alice, who exists on the target
	want := wantSet(exportResourceChores, exportResourceRewards, exportResourceRoutines, exportResourceCalendar, exportResourceRecipes, exportResourceSittings)

	var err error
	_ = captureStdout(func() { err = runImport(context.Background(), rec.client(t), data, want, importTestToday, false) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireContains(t, "chore", rec.posts("/chores"), "921")
	requireContains(t, "reward", rec.posts("/rewards"), "921")
	requireContains(t, "routine", rec.posts("/chores/create_multiple"), "921")
	requireContains(t, "event", rec.posts("/calendar_events"), "921")
	for _, q := range rec.reqs {
		if strings.Contains(q, "811") {
			t.Errorf("source ID leaked into request: %s", q)
		}
	}
	// The export's own slices stay untouched.
	if data.Chores[0].AssigneeID != "811" {
		t.Errorf("remap modified the caller's data: %q", data.Chores[0].AssigneeID)
	}
}

func TestRunImport_CrossFrameSittingGetsNewRecipeID(t *testing.T) {
	setFrameID(t, "frame-b")
	rec := &remapRecorder{}
	data := remapExport()
	want := wantSet(exportResourceRecipes, exportResourceSittings)

	var err error
	_ = captureStdout(func() { err = runImport(context.Background(), rec.client(t), data, want, importTestToday, false) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireContains(t, "sitting", rec.posts("/meals/sittings"), "tgt-recipe")
}

func TestRunImport_CrossFrameSittingWithoutRecipesDropsLink(t *testing.T) {
	setFrameID(t, "frame-b")
	rec := &remapRecorder{}
	data := remapExport()

	var err error
	_ = captureStdout(func() {
		err = runImport(context.Background(), rec.client(t), data, wantSet(exportResourceSittings), importTestToday, false)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sittings := rec.posts("/meals/sittings")
	if len(sittings) != 1 || strings.Contains(sittings[0], "src-recipe") {
		t.Errorf("sitting should not carry the source recipe ID: %v", sittings)
	}
}

func TestRunImport_CrossFrameUnmatchedLabelFailsBeforeWrites(t *testing.T) {
	setFrameID(t, "frame-b")
	rec := &remapRecorder{}
	data := remapExport()
	data.Rewards = []lib.Reward{{Title: "Reward", CategoryID: "812"}}
	want := wantSet(exportResourceChores, exportResourceRewards, exportResourceRecipes)

	var err error
	_ = captureStdout(func() { err = runImport(context.Background(), rec.client(t), data, want, importTestToday, false) })
	if err == nil || !strings.Contains(err.Error(), `"Bob"`) {
		t.Fatalf("want error naming the unmatched label, got: %v", err)
	}
	if n := rec.writeCount(); n != 0 {
		t.Errorf("got %d writes, want 0: %v", n, rec.reqs)
	}
}

func TestRunImport_CrossFrameUnknownCategoryIDFailsBeforeWrites(t *testing.T) {
	setFrameID(t, "frame-b")
	rec := &remapRecorder{}
	data := remapExport()
	data.Categories = nil
	want := wantSet(exportResourceChores, exportResourceCategories)

	var err error
	_ = captureStdout(func() { err = runImport(context.Background(), rec.client(t), data, want, importTestToday, false) })
	if err == nil || !strings.Contains(err.Error(), "811") {
		t.Fatalf("want error naming the unknown ID, got: %v", err)
	}
	if n := rec.writeCount(); n != 0 {
		t.Errorf("got %d writes, want 0", n)
	}
}

func TestRunImport_CrossFrameCategoriesResourceCreatesMissing(t *testing.T) {
	setFrameID(t, "frame-b")
	rec := &remapRecorder{}
	data := remapExport()
	data.Rewards = []lib.Reward{{Title: "Reward", CategoryID: "812"}}
	want := wantSet(exportResourceChores, exportResourceRewards, exportResourceCategories)

	var err error
	_ = captureStdout(func() { err = runImport(context.Background(), rec.client(t), data, want, importTestToday, false) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cats := rec.posts("/categories")
	requireContains(t, "category", cats, `"label":"Bob"`)
	if strings.Contains(cats[0], "Alice") {
		t.Errorf("matched category must not be created again: %s", cats[0])
	}
	requireContains(t, "reward", rec.posts("/rewards"), "922")
	requireContains(t, "chore", rec.posts("/chores"), "921")
}

func TestRunImportDryRunOnFrame_CrossFramePrintsMappingWithoutWrites(t *testing.T) {
	setFrameID(t, "frame-b")
	rec := &remapRecorder{}
	data := remapExport()
	data.Rewards = []lib.Reward{{Title: "Reward", CategoryID: "812"}}
	want := wantSet(exportResourceChores, exportResourceRewards)

	var err error
	out := captureStdout(func() {
		err = runImportDryRunOnFrame(context.Background(), rec.client(t), data, want, importTestToday, false)
	})
	if err == nil || !strings.Contains(err.Error(), `"Bob"`) {
		t.Errorf("want error naming the unmatched label, got: %v", err)
	}
	for _, sub := range []string{`"Alice": 811 -> 921`, `"Bob": 812 -> no match`} {
		if !strings.Contains(out, sub) {
			t.Errorf("want %q in output:\n%s", sub, out)
		}
	}
	if n := rec.writeCount(); n != 0 {
		t.Errorf("dry run wrote %d times", n)
	}
}

func TestRemapCategories_SkipsRowsImportSkips(t *testing.T) {
	data := ExportData{Chores: []lib.Chore{
		{ID: "1", Title: "done", Status: lib.ChoreStatusComplete, AssigneeID: "gone"},
		{ID: "2", Title: "routine", Routine: true, AssigneeID: "gone"},
		{ID: "3", Title: "open", AssigneeID: "a"},
	}}
	var seen []string
	forEachCategoryRef(&data, wantSet(exportResourceChores), func(id *string) { seen = append(seen, *id) })
	if len(seen) != 1 || seen[0] != "a" {
		t.Errorf("got refs %v, want [a]", seen)
	}
}

func TestImportRecipesMapped_SkippedRecipeMapsToExistingID(t *testing.T) {
	ids := &sync.Map{}
	recipes := []lib.Recipe{{ID: "src-recipe", Title: "Tacos"}}
	var total, failed int
	_ = captureStdout(func() {
		total, failed = importRecipesMapped(context.Background(), nil, recipes,
			map[string]bool{"Tacos": true}, map[string]string{"Tacos": "existing-recipe"}, ids)
	})
	if total != 0 || failed != 0 {
		t.Fatalf("total, failed = %d, %d; want 0, 0", total, failed)
	}
	got := remapSittingRecipes([]lib.MealSitting{{Summary: "dinner", RecipeID: "src-recipe"}}, ids)
	if got[0].RecipeID != "existing-recipe" {
		t.Errorf("RecipeID = %q, want existing-recipe", got[0].RecipeID)
	}
}
