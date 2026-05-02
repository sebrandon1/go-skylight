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
	outputFormat = outputJSON
	output := captureStdout(func() {
		printOutput(map[string]string{"key": "value"})
	})
	if !strings.Contains(output, `"key": "value"`) {
		t.Errorf("Expected JSON output, got: %s", output)
	}
}

func TestPrintOutputChoresTable(t *testing.T) {
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

func TestValidateDate_Empty(t *testing.T) {
	if err := validateDate(""); err != nil {
		t.Errorf("expected nil for empty string, got %v", err)
	}
}

func TestValidateDate_Valid(t *testing.T) {
	if err := validateDate("2026-04-27"); err != nil {
		t.Errorf("expected nil for valid date, got %v", err)
	}
}

func TestValidateDate_Invalid(t *testing.T) {
	cases := []string{"not-a-date", "04/27/2026", "2026-13-01", "2026-4-7"}
	for _, c := range cases {
		if err := validateDate(c); err == nil {
			t.Errorf("expected error for %q, got nil", c)
		}
	}
}

func TestPrintOutputBountiesTable(t *testing.T) {
	outputFormat = outputTable
	bounties := []lib.Bounty{
		{Chore: lib.Chore{ID: "c1", Title: "Walk dog", Points: 10, DueDate: "2026-04-28"}, Reward: lib.Reward{ID: "r1", Title: "Ice cream"}},
	}
	output := captureStdout(func() { printOutput(bounties) })
	if !strings.Contains(output, "Walk dog") {
		t.Errorf("Expected chore title in bounties table, got: %s", output)
	}
	if !strings.Contains(output, "Ice cream") {
		t.Errorf("Expected reward title in bounties table, got: %s", output)
	}
}

func TestPrintOutputSourceCalendarsTable(t *testing.T) {
	outputFormat = outputTable
	cals := []lib.SourceCalendar{
		{ID: "sc1", Name: "Google Calendar", Provider: "google", Color: "blue"},
	}
	output := captureStdout(func() { printOutput(cals) })
	if !strings.Contains(output, "Google Calendar") {
		t.Errorf("Expected source calendar name in table, got: %s", output)
	}
	if !strings.Contains(output, "PROVIDER") {
		t.Errorf("Expected PROVIDER header in table, got: %s", output)
	}
}

func TestPrintOutputDevicesTable(t *testing.T) {
	outputFormat = outputTable
	devices := []lib.Device{
		{ID: "d1", Name: "Kitchen Frame", Online: true},
		{ID: "d2", Name: "Bedroom Frame", Online: false},
	}
	output := captureStdout(func() { printOutput(devices) })
	if !strings.Contains(output, "Kitchen Frame") {
		t.Errorf("Expected device name in table, got: %s", output)
	}
	if !strings.Contains(output, boolYes) {
		t.Errorf("Expected %q for online device, got: %s", boolYes, output)
	}
	if !strings.Contains(output, boolNo) {
		t.Errorf("Expected %q for offline device, got: %s", boolNo, output)
	}
}

func TestPrintOutputAvatarsTable(t *testing.T) {
	outputFormat = outputTable
	avatars := []lib.Avatar{
		{ID: "a1", Name: "Alice Avatar", ImageURL: "https://example.com/avatar.png"},
	}
	output := captureStdout(func() { printOutput(avatars) })
	if !strings.Contains(output, "Alice Avatar") {
		t.Errorf("Expected avatar name in table, got: %s", output)
	}
	if !strings.Contains(output, "IMAGE URL") {
		t.Errorf("Expected IMAGE URL header in table, got: %s", output)
	}
}

func TestPrintOutputColorsTable(t *testing.T) {
	outputFormat = outputTable
	colors := []lib.Color{
		{Name: "Sky Blue", Hex: "#87CEEB"},
	}
	output := captureStdout(func() { printOutput(colors) })
	if !strings.Contains(output, "Sky Blue") {
		t.Errorf("Expected color name in table, got: %s", output)
	}
	if !strings.Contains(output, "#87CEEB") {
		t.Errorf("Expected hex value in table, got: %s", output)
	}
}

func TestPrintOutputListsTable(t *testing.T) {
	outputFormat = outputTable
	lists := []lib.List{
		{ID: "l1", Title: "Shopping", Color: "green", Kind: "checklist", Items: []lib.ListItem{{}, {}}},
	}
	output := captureStdout(func() { printOutput(lists) })
	if !strings.Contains(output, "Shopping") {
		t.Errorf("Expected list title in table, got: %s", output)
	}
	if !strings.Contains(output, "2") {
		t.Errorf("Expected item count in table, got: %s", output)
	}
}

func TestPrintOutputMealCategoriesTable(t *testing.T) {
	outputFormat = outputTable
	cats := []lib.MealCategory{
		{ID: "mc1", Name: "Dinner", Color: "red"},
	}
	output := captureStdout(func() { printOutput(cats) })
	if !strings.Contains(output, "Dinner") {
		t.Errorf("Expected meal category name in table, got: %s", output)
	}
}

func TestPrintOutputRecipesTable(t *testing.T) {
	outputFormat = outputTable
	recipes := []lib.Recipe{
		{ID: "rec1", Title: "Pasta", MealCategoryID: "mc1", URL: "https://example.com/pasta"},
	}
	output := captureStdout(func() { printOutput(recipes) })
	if !strings.Contains(output, "Pasta") {
		t.Errorf("Expected recipe title in table, got: %s", output)
	}
	if !strings.Contains(output, "URL") {
		t.Errorf("Expected URL header in table, got: %s", output)
	}
}

func TestPrintOutputMealSittingsTable(t *testing.T) {
	outputFormat = outputTable
	sittings := []lib.MealSitting{
		{ID: "ms1", Summary: "Taco Tuesday", Date: "2026-04-28", RecipeID: "rec1", MealCategoryID: "mc1"},
	}
	output := captureStdout(func() { printOutput(sittings) })
	if !strings.Contains(output, "Taco Tuesday") {
		t.Errorf("Expected meal sitting summary in table, got: %s", output)
	}
}

func TestPrintOutputPhotosTable(t *testing.T) {
	outputFormat = outputTable
	photos := []lib.Photo{
		{ID: "p1", AssetType: "image/jpeg", Status: "active", CreatedAt: "2026-04-28T10:00:00Z"},
	}
	output := captureStdout(func() { printOutput(photos) })
	if !strings.Contains(output, "p1") {
		t.Errorf("Expected photo ID in table, got: %s", output)
	}
	if !strings.Contains(output, "image/jpeg") {
		t.Errorf("Expected asset type in table, got: %s", output)
	}
}
