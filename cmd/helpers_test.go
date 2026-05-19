package cmd

import (
	"bytes"
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

func TestPrintOutputDevicesTable(t *testing.T) {
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = outputJSON })
	devices := []lib.Device{
		{ID: "d1", Name: "Kitchen Frame", Online: true},
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
