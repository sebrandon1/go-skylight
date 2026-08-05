package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestPrintJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected JSON output, got empty string")
	}

	if !strings.Contains(output, "\"key\": \"value\"") {
		t.Errorf("Expected pretty-printed JSON with key/value, got: %s", output)
	}
}

func TestPrintJSONWithStruct(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := testStruct{Name: "test", Age: 30}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"name\": \"test\"") {
		t.Errorf("Expected name field in output, got: %s", output)
	}
	if !strings.Contains(output, "\"age\": 30") {
		t.Errorf("Expected age field in output, got: %s", output)
	}
}

func TestPrintJSONWithSlice(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := []string{"a", "b", "c"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected JSON output for slice")
	}
	if !strings.Contains(output, "\"a\"") {
		t.Errorf("Expected element 'a' in output, got: %s", output)
	}
	if !strings.Contains(output, "\"b\"") {
		t.Errorf("Expected element 'b' in output, got: %s", output)
	}
	if !strings.Contains(output, "\"c\"") {
		t.Errorf("Expected element 'c' in output, got: %s", output)
	}
}

func TestPrintJSONWithNestedStruct(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	type inner struct {
		Value string `json:"value"`
	}
	type outer struct {
		Inner inner `json:"inner"`
	}

	data := outer{Inner: inner{Value: "nested"}}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"value\": \"nested\"") {
		t.Errorf("Expected nested value in output, got: %s", output)
	}
	if !strings.Contains(output, "\"inner\"") {
		t.Errorf("Expected inner key in output, got: %s", output)
	}
}

func TestPrintJSONWithEmptyMap(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "{}") {
		t.Errorf("Expected empty JSON object, got: %s", output)
	}
}

func TestPrintJSONWithEmptySlice(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := []string{}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "[]") {
		t.Errorf("Expected empty JSON array, got: %s", output)
	}
}

func TestPrintJSONWithNumbers(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]int{"count": 42}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"count\": 42") {
		t.Errorf("Expected count field with integer value, got: %s", output)
	}
}

func TestPrintJSONWithBooleans(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]bool{"active": true, "deleted": false}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"active\": true") {
		t.Errorf("Expected active=true in output, got: %s", output)
	}
	if !strings.Contains(output, "\"deleted\": false") {
		t.Errorf("Expected deleted=false in output, got: %s", output)
	}
}

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintOutputJSONFallback(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputJSON
	output := captureStdout(func() {
		printOutput(map[string]string{"key": "value"})
	})
	if !strings.Contains(output, `"key": "value"`) {
		t.Errorf("Expected JSON output, got: %s", output)
	}
}

func TestPrintOutputChoresTable(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputTable
	chores := []lib.Chore{
		{ID: "1", Title: "Clean room", Status: "pending", DueDate: "2026-04-28", Points: 10, AssigneeID: "cat1"},
	}
	output := captureStdout(func() { printOutput(chores) })
	if !strings.Contains(output, "Clean room") {
		t.Errorf("Expected chore title in table output, got: %s", output)
	}
	if !strings.Contains(output, "TITLE") {
		t.Errorf("Expected header in table output, got: %s", output)
	}
}

func TestPrintOutputRewardsTable(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputTable
	rewards := []lib.Reward{
		{ID: "r1", Title: "Ice cream", Points: 50, EmojiIcon: "🍦", Redeemed: false},
	}
	output := captureStdout(func() { printOutput(rewards) })
	if !strings.Contains(output, "Ice cream") {
		t.Errorf("Expected reward title in table output, got: %s", output)
	}
	if !strings.Contains(output, "POINTS") {
		t.Errorf("Expected POINTS header, got: %s", output)
	}
}

func TestPrintOutputFramesTable(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputTable
	frames := []lib.Frame{
		{ID: "f1", Name: "Kitchen Frame", TimeZone: "America/New_York"},
	}
	output := captureStdout(func() { printOutput(frames) })
	if !strings.Contains(output, "Kitchen Frame") {
		t.Errorf("Expected frame name in table output, got: %s", output)
	}
	if !strings.Contains(output, "TIMEZONE") {
		t.Errorf("Expected TIMEZONE header, got: %s", output)
	}
}

func TestPrintOutputCalendarTable(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputTable
	events := []lib.CalendarEvent{
		{ID: "e1", Title: "Soccer practice", StartAt: "2026-04-28T16:00:00Z", AllDay: false},
	}
	output := captureStdout(func() { printOutput(events) })
	if !strings.Contains(output, "Soccer practice") {
		t.Errorf("Expected event title in table output, got: %s", output)
	}
	if !strings.Contains(output, "ALL DAY") {
		t.Errorf("Expected ALL DAY header, got: %s", output)
	}
}

func TestPrintOutputCategoriesTable(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputTable
	cats := []lib.Category{
		{ID: "c1", Name: "Alice", Color: "blue"},
	}
	output := captureStdout(func() { printOutput(cats) })
	if !strings.Contains(output, "Alice") {
		t.Errorf("Expected category name in table output, got: %s", output)
	}
	if !strings.Contains(output, "COLOR") {
		t.Errorf("Expected COLOR header, got: %s", output)
	}
}

func TestPrintOutputRewardRedeemedFlag(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputTable
	rewards := []lib.Reward{
		{ID: "r1", Title: "Movie night", Points: 100, Redeemed: true},
		{ID: "r2", Title: "Candy", Points: 20, Redeemed: false},
	}
	output := captureStdout(func() { printOutput(rewards) })
	if !strings.Contains(output, "yes") {
		t.Errorf("Expected 'yes' for redeemed reward, got: %s", output)
	}
	if !strings.Contains(output, "no") {
		t.Errorf("Expected 'no' for unredeemed reward, got: %s", output)
	}
}

func TestPrintOutputUnknownTypeDefaultsToJSON(t *testing.T) {
	t.Cleanup(func() { outputFormat = "" })
	outputFormat = outputTable
	output := captureStdout(func() {
		printOutput(map[string]int{"count": 5})
	})
	if !strings.Contains(output, `"count": 5`) {
		t.Errorf("Expected JSON fallback for unknown type, got: %s", output)
	}
}

func TestPrintJSONWithNilValue(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printJSON(nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "null") {
		t.Errorf("Expected null for nil input, got: %s", output)
	}
}

func TestPrintJSONOutputIsIndented(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// MarshalIndent with two-space indent should produce indented output
	if !strings.Contains(output, "  ") {
		t.Errorf("Expected indented output with two spaces, got: %s", output)
	}
}

func TestValidateDate(t *testing.T) {
	cases := []struct {
		input string
		valid bool
	}{
		{"", true},
		{"2026-01-01", true},
		{"2026-12-31", true},
		{"not-a-date", false},
		{"04/27/2026", false},
		{"2026-13-01", false},
		{"2026-4-7", false},
	}
	for _, c := range cases {
		err := validateDate(c.input)
		if c.valid && err != nil {
			t.Errorf("validateDate(%q): expected no error, got %v", c.input, err)
		}
		if !c.valid && err == nil {
			t.Errorf("validateDate(%q): expected error, got nil", c.input)
		}
	}
}

func TestValidateEnum(t *testing.T) {
	allowed := []string{"json", "table"}
	cases := []struct {
		input string
		valid bool
	}{
		{"", true},
		{"json", true},
		{"table", true},
		{"xml", false},
		{"JSON", false},
	}
	for _, c := range cases {
		err := validateEnum(c.input, allowed)
		if c.valid && err != nil {
			t.Errorf("validateEnum(%q): expected no error, got %v", c.input, err)
		}
		if !c.valid && err == nil {
			t.Errorf("validateEnum(%q): expected error, got nil", c.input)
		}
	}
}

func TestPrintOutputDevicesTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	devices := []lib.Device{
		{ID: "d1", Name: "Kitchen Frame", Online: true, Brightness: 200, SleepsAt: "00:00", WakesAt: "06:00", Nightlight: true},
		{ID: "d2", Name: "Bedroom Frame", Online: false},
	}
	output := captureStdout(func() { printOutput(devices) })
	if !strings.Contains(output, "Kitchen Frame") {
		t.Errorf("Expected device name, got: %s", output)
	}
	if !strings.Contains(output, boolYes) {
		t.Errorf("Expected 'yes' for online device, got: %s", output)
	}
	if !strings.Contains(output, boolNo) {
		t.Errorf("Expected 'no' for offline device, got: %s", output)
	}
	if !strings.Contains(output, "00:00-06:00") {
		t.Errorf("Expected sleep schedule in output, got: %s", output)
	}
}

func TestPrintOutputAvatarsTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	avatars := []lib.Avatar{
		{ID: "a1", Name: "Dragon", ImageURL: "https://cdn.example.com/dragon.png"},
	}
	output := captureStdout(func() { printOutput(avatars) })
	if !strings.Contains(output, "Dragon") {
		t.Errorf("Expected avatar name, got: %s", output)
	}
	if !strings.Contains(output, "IMAGE URL") {
		t.Errorf("Expected IMAGE URL header, got: %s", output)
	}
}

func TestPrintOutputColorsTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	colors := []lib.Color{
		{Name: "Red", Hex: "#FF0000"},
		{Name: "Blue", Hex: "#0000FF"},
	}
	output := captureStdout(func() { printOutput(colors) })
	if !strings.Contains(output, "Red") {
		t.Errorf("Expected color name, got: %s", output)
	}
	if !strings.Contains(output, "#FF0000") {
		t.Errorf("Expected hex value, got: %s", output)
	}
	if !strings.Contains(output, "HEX") {
		t.Errorf("Expected HEX header, got: %s", output)
	}
}

func TestPrintOutputListsTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	lists := []lib.List{
		{ID: "l1", Title: "Shopping", Color: "green", Kind: "checklist", Items: []lib.ListItem{{}, {}}},
	}
	output := captureStdout(func() { printOutput(lists) })
	if !strings.Contains(output, "Shopping") {
		t.Errorf("Expected list title, got: %s", output)
	}
	if !strings.Contains(output, "ITEMS") {
		t.Errorf("Expected ITEMS header, got: %s", output)
	}
	if !strings.Contains(output, "2") {
		t.Errorf("Expected item count '2', got: %s", output)
	}
}

func TestPrintOutputMealCategoriesTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	cats := []lib.MealCategory{
		{ID: "mc1", Name: "Breakfast", Color: "yellow"},
	}
	output := captureStdout(func() { printOutput(cats) })
	if !strings.Contains(output, "Breakfast") {
		t.Errorf("Expected meal category name, got: %s", output)
	}
	if !strings.Contains(output, "COLOR") {
		t.Errorf("Expected COLOR header, got: %s", output)
	}
}

func TestPrintOutputRecipesTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	recipes := []lib.Recipe{
		{ID: "r1", Title: "Pancakes", MealCategoryID: "mc1", URL: "https://example.com/pancakes"},
	}
	output := captureStdout(func() { printOutput(recipes) })
	if !strings.Contains(output, "Pancakes") {
		t.Errorf("Expected recipe title, got: %s", output)
	}
	if !strings.Contains(output, "https://example.com/pancakes") {
		t.Errorf("Expected recipe URL, got: %s", output)
	}
}

func TestPrintOutputMealSittingsTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	sittings := []lib.MealSitting{
		{ID: "s1", Summary: "Taco Tuesday", Date: "2026-05-05", RecipeID: "r1", MealCategoryID: "mc2"},
	}
	output := captureStdout(func() { printOutput(sittings) })
	if !strings.Contains(output, "Taco Tuesday") {
		t.Errorf("Expected sitting summary, got: %s", output)
	}
	if !strings.Contains(output, "SUMMARY") {
		t.Errorf("Expected SUMMARY header, got: %s", output)
	}
}

func TestPrintOutputPhotosTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	photos := []lib.Photo{
		{ID: "p1", AssetType: "image/jpeg", Status: "active", CreatedAt: "2026-05-01T12:00:00Z"},
	}
	output := captureStdout(func() { printOutput(photos) })
	if !strings.Contains(output, "image/jpeg") {
		t.Errorf("Expected photo asset type, got: %s", output)
	}
	if !strings.Contains(output, "STATUS") {
		t.Errorf("Expected STATUS header, got: %s", output)
	}
}

func TestPrintOutputBountiesTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	bounties := []lib.Bounty{
		{
			Chore:  lib.Chore{ID: "c1", Title: "Wash dishes", Points: 10, DueDate: "2026-05-01"},
			Reward: lib.Reward{ID: "rw1", Title: "Candy bar"},
		},
	}
	output := captureStdout(func() { printOutput(bounties) })
	if !strings.Contains(output, "Wash dishes") {
		t.Errorf("Expected chore title, got: %s", output)
	}
	if !strings.Contains(output, "Candy bar") {
		t.Errorf("Expected reward title, got: %s", output)
	}
	if !strings.Contains(output, "CHORE TITLE") {
		t.Errorf("Expected CHORE TITLE header, got: %s", output)
	}
}

func TestPrintOutputSourceCalendarsTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	cals := []lib.SourceCalendar{
		{ID: "sc1", Name: "Family", Provider: "google", Color: "blue"},
	}
	output := captureStdout(func() { printOutput(cals) })
	if !strings.Contains(output, "Family") {
		t.Errorf("Expected calendar name, got: %s", output)
	}
	if !strings.Contains(output, "PROVIDER") {
		t.Errorf("Expected PROVIDER header, got: %s", output)
	}
}

func TestBuildCatNames(t *testing.T) {
	cats := []lib.Category{
		{ID: "1", Name: "Alice"},
		{ID: "2", Name: "Bob"},
	}
	got := buildCatNames(cats)
	if got["1"] != "Alice" || got["2"] != "Bob" {
		t.Errorf("unexpected result: %v", got)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
}

func TestResolveRewardPointNames(t *testing.T) {
	categories := []lib.Category{
		{ID: "1", Name: "Alice"},
		{ID: "2", Name: "Bob"},
	}
	points := []lib.RewardPointEntry{
		{CategoryID: 1, CurrentPointBalance: 10},
		{CategoryID: 2, CurrentPointBalance: 20},
		{CategoryID: 99, CurrentPointBalance: 5}, // no matching category
	}

	got := resolveRewardPointNames(points, categories)
	want := []pointEntry{
		{Name: "Alice", Balance: 10},
		{Name: "Bob", Balance: 20},
		{Name: "99", Balance: 5},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestResolveCatName(t *testing.T) {
	orig := activeCatNames
	t.Cleanup(func() { activeCatNames = orig })

	activeCatNames = nil
	if got := resolveCatName(""); got != "" {
		t.Errorf("empty id: got %q, want %q", got, "")
	}
	if got := resolveCatName("42"); got != "42" {
		t.Errorf("nil map: got %q, want %q", got, "42")
	}

	activeCatNames = map[string]string{"42": "Alice"}
	if got := resolveCatName("42"); got != "Alice" {
		t.Errorf("found: got %q, want Alice", got)
	}
	if got := resolveCatName("99"); got != "99" {
		t.Errorf("not found: got %q, want %q", got, "99")
	}
}

func TestLoadCatNames_Success(t *testing.T) {
	orig := activeCatNames
	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() {
		activeCatNames = orig
		frameID = origFrameID
	})

	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"1","attributes":{"label":"Alice","color":"#FF0000"}}]}`)
	})
	loadCatNames(context.Background(), client)
	if activeCatNames["1"] != "Alice" {
		t.Errorf("expected activeCatNames[\"1\"] == \"Alice\", got %q", activeCatNames["1"])
	}
}

func TestMaybeLoadCatNames_TableMode(t *testing.T) {
	origFmt := outputFormat
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = origFmt })

	orig := activeCatNames
	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() {
		activeCatNames = orig
		frameID = origFrameID
	})

	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"z1","attributes":{"label":"Zara","color":"#000"}}]}`)
	})
	maybeLoadCatNames(context.Background(), client)
	if activeCatNames["z1"] != "Zara" {
		t.Errorf("expected activeCatNames to be populated in table mode, got %v", activeCatNames)
	}
}

func TestMaybeLoadCatNames_JSONMode(t *testing.T) {
	origFmt := outputFormat
	outputFormat = outputJSON
	t.Cleanup(func() { outputFormat = origFmt })

	orig := activeCatNames
	activeCatNames = nil
	t.Cleanup(func() { activeCatNames = orig })

	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("maybeLoadCatNames should not make HTTP calls in JSON mode")
	})
	maybeLoadCatNames(context.Background(), client)
	if activeCatNames != nil {
		t.Errorf("expected activeCatNames to stay nil in JSON mode, got %v", activeCatNames)
	}
}

func TestLoadCatNames_ErrorSilenced(t *testing.T) {
	orig := activeCatNames
	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() {
		activeCatNames = orig
		frameID = origFrameID
	})
	activeCatNames = nil

	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	loadCatNames(context.Background(), client)
	if activeCatNames != nil {
		t.Errorf("expected activeCatNames to remain nil on error, got %v", activeCatNames)
	}
}

func TestGetFrameOrFail_Success(t *testing.T) {
	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"f1","attributes":{"name":"Kitchen"}}}`)
	})

	frame, err := getFrameOrFail(context.Background(), client, "f1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if frame.ID != "f1" || frame.Name != "Kitchen" {
		t.Errorf("unexpected frame: %+v", frame)
	}
}

func TestGetFrameOrFail_Error(t *testing.T) {
	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := getFrameOrFail(context.Background(), client, "f1")
	if err == nil {
		t.Fatal("expected error from getFrameOrFail on server error")
	}
	if !strings.Contains(err.Error(), "getting frame info") {
		t.Errorf("expected error to mention 'getting frame info', got: %v", err)
	}
}

func TestPrintSuccess(t *testing.T) {
	t.Cleanup(func() { quiet = false })

	quiet = false
	out := captureStdout(func() { printSuccess("done") })
	if !strings.Contains(out, "done") {
		t.Errorf("expected message when quiet is false, got: %q", out)
	}

	quiet = true
	out = captureStdout(func() { printSuccess("done") })
	if out != "" {
		t.Errorf("expected no output when quiet is true, got: %q", out)
	}
}

func TestPrintSuccessf(t *testing.T) {
	t.Cleanup(func() { quiet = false })

	quiet = false
	out := captureStdout(func() { printSuccessf("done %d\n", 3) })
	if !strings.Contains(out, "done 3") {
		t.Errorf("expected formatted message when quiet is false, got: %q", out)
	}

	quiet = true
	out = captureStdout(func() { printSuccessf("done %d\n", 3) })
	if out != "" {
		t.Errorf("expected no output when quiet is true, got: %q", out)
	}
}

func TestConfirmAction(t *testing.T) {
	cases := []struct {
		name    string
		yesFlag bool
		input   string
		want    bool
	}{
		{"yes_flag_bypasses_stdin", true, "", true},
		{"y_accepted", false, "y\n", true},
		{"Y_accepted_case_insensitive", false, "Y\n", true},
		{"yes_accepted", false, "yes\n", true},
		{"YES_accepted", false, "YES\n", true},
		{"n_declined", false, "n\n", false},
		{"no_declined", false, "no\n", false},
		{"empty_input_eof_declined", false, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origYes := yes
			yes = tc.yesFlag
			t.Cleanup(func() { yes = origYes })

			if !tc.yesFlag {
				mockStdin(t, tc.input)
			}

			got := confirmAction("test prompt")
			if got != tc.want {
				t.Errorf("confirmAction: got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPrintDryRun(t *testing.T) {
	t.Cleanup(func() { quiet = false })

	out := captureStdout(func() { printDryRun("delete chore %s", "c1") })
	if !strings.Contains(out, "Dry run: would delete chore c1") {
		t.Errorf("expected dry run message, got: %q", out)
	}

	// Unlike printSuccess/printSuccessf, --quiet must not suppress this —
	// it's the only output a dry-run invocation produces.
	quiet = true
	out = captureStdout(func() { printDryRun("delete chore %s", "c1") })
	if !strings.Contains(out, "Dry run: would delete chore c1") {
		t.Errorf("expected dry run message even with --quiet, got: %q", out)
	}
}
