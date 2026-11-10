//go:build integration

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

// integrationTestPrefix marks items these tests create. The sweep in
// lib/integration_test.go removes anything an interrupted run leaves behind.
const integrationTestPrefix = "integration-test-"

var (
	cliAuthOnce sync.Once
	cliAuthErr  error
)

// integrationClient authenticates the way the CLI does, so a refresh token is
// rotated and saved back to the config. Without one it logs in by email and
// password as a separate session.
func integrationClient(t *testing.T) *lib.Client {
	t.Helper()

	cliAuthOnce.Do(func() {
		loadConfig()
		outputFormat = outputJSON
		if frameID == "" {
			return
		}
		if handled, err := tryRefreshTokenAuth(); handled {
			cliAuthErr = err
			return
		}
		if email == "" || password == "" {
			return
		}
		tok, err := lib.LoginHeadless(email, password, integrationTestPrefix+frameID)
		if err != nil {
			cliAuthErr = err
			return
		}
		autoClient, cliAuthErr = lib.NewClientWithToken("integration", tok.AccessToken, cliClientOptions()...)
	})

	if cliAuthErr != nil {
		t.Fatalf("integration auth: %v", cliAuthErr)
	}
	if autoClient == nil {
		t.Skip("skipping: set SKYLIGHT_FRAME_ID and credentials, or add them to ~/.skylight/config")
	}
	return autoClient
}

func dateFromNow(days int) string {
	return time.Now().AddDate(0, 0, days).Format(lib.DateFormat)
}

// prefixedItems is what the test prefix currently matches on the frame.
type prefixedItems struct {
	chores   []lib.Chore
	rewards  []lib.Reward
	lists    []lib.List
	recipes  []lib.Recipe
	sittings []lib.MealSitting
	events   []lib.CalendarEvent
}

// findPrefixed lists every resource and keeps what starts with prefix. Test
// items are all dated within the next month.
func findPrefixed(ctx context.Context, t *testing.T, client *lib.Client, prefix string) prefixedItems {
	t.Helper()
	var found prefixedItems

	chores, err := listChoresWithUpForGrabs(ctx, client, lib.ChoreListOptions{After: dateFromNow(0), Before: dateFromNow(30)})
	if err != nil {
		t.Fatalf("listing chores: %v", err)
	}
	seenChores := map[string]bool{}
	for _, c := range chores {
		if strings.HasPrefix(c.Title, prefix) && !seenChores[c.ID] {
			seenChores[c.ID] = true
			found.chores = append(found.chores, c)
		}
	}

	rewards, err := client.ListRewards(ctx, frameID)
	if err != nil {
		t.Fatalf("listing rewards: %v", err)
	}
	for _, r := range rewards {
		if strings.HasPrefix(r.Title, prefix) {
			found.rewards = append(found.rewards, r)
		}
	}

	lists, err := client.ListLists(ctx, frameID)
	if err != nil {
		t.Fatalf("listing lists: %v", err)
	}
	for _, l := range lists {
		if strings.HasPrefix(l.Title, prefix) {
			found.lists = append(found.lists, l)
		}
	}

	recipes, err := client.ListRecipes(ctx, frameID)
	if err != nil {
		t.Fatalf("listing recipes: %v", err)
	}
	for _, r := range recipes {
		if strings.HasPrefix(r.Title, prefix) {
			found.recipes = append(found.recipes, r)
		}
	}

	sittings, err := client.ListMealSittings(ctx, frameID, lib.MealSittingListOptions{DateMin: dateFromNow(0), DateMax: dateFromNow(30)})
	if err != nil {
		t.Fatalf("listing sittings: %v", err)
	}
	for _, s := range sittings {
		if strings.HasPrefix(s.Summary, prefix) {
			found.sittings = append(found.sittings, s)
		}
	}

	events, err := client.ListCalendarEvents(ctx, frameID, dateFromNow(0), dateFromNow(30), "")
	if err != nil {
		t.Fatalf("listing calendar events: %v", err)
	}
	for _, e := range events {
		if strings.HasPrefix(e.Title, prefix) {
			found.events = append(found.events, e)
		}
	}

	return found
}

// deletePrefixed removes everything findPrefixed matches. Sittings go before
// the recipes they link to.
func deletePrefixed(ctx context.Context, t *testing.T, client *lib.Client, prefix string) {
	t.Helper()
	found := findPrefixed(ctx, t, client, prefix)

	report := func(what string, err error) {
		if err != nil && !lib.IsNotFound(err) {
			t.Errorf("cleanup %s: %v", what, err)
		}
	}
	for _, s := range found.sittings {
		report("sitting", client.DeleteMealSitting(ctx, frameID, s.ID, s.Date))
	}
	for _, c := range found.chores {
		if c.Recurring {
			base, _, _ := strings.Cut(c.ID, "-")
			report("recurring chore", client.DeleteRecurringChore(ctx, frameID, base))
			continue
		}
		report("chore", client.DeleteChore(ctx, frameID, c.ID))
	}
	for _, r := range found.rewards {
		report("reward", client.DeleteReward(ctx, frameID, r.ID))
	}
	for _, l := range found.lists {
		report("list", client.DeleteList(ctx, frameID, l.ID))
	}
	for _, r := range found.recipes {
		report("recipe", client.DeleteRecipe(ctx, frameID, r.ID))
	}
	for _, e := range found.events {
		report("event", client.DeleteCalendarEvent(ctx, frameID, e.ID))
	}
}

// firstProfileAndMealCategory returns the IDs needed to create test items.
func firstProfileAndMealCategory(ctx context.Context, t *testing.T, client *lib.Client) (profileID, mealCategoryID string) {
	t.Helper()
	profiles, err := client.ListProfiles(ctx, frameID)
	if err != nil {
		t.Fatalf("listing profiles: %v", err)
	}
	mealCategories, err := client.ListMealCategories(ctx, frameID)
	if err != nil {
		t.Fatalf("listing meal categories: %v", err)
	}
	if len(profiles) == 0 || len(mealCategories) == 0 {
		t.Skip("frame needs a profile and a meal category")
	}
	return profiles[0].ID, mealCategories[0].ID
}

// TestIntegration_ImportThenSkipExisting runs a real import, then repeats it
// with skip-existing, which must create nothing further. It covers #550
// (bounties duplicating their chores and rewards), #551 and #562 (recipe-linked
// sittings), and #552 (skip-existing duplicates).
func TestIntegration_ImportThenSkipExisting(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	profileID, mealCategoryID := firstProfileAndMealCategory(ctx, t, client)
	prefix := fmt.Sprintf("%s%d-import", integrationTestPrefix, time.Now().UnixMilli())
	t.Cleanup(func() { deletePrefixed(ctx, t, client, prefix) })

	linkedRecipe, err := client.CreateRecipe(ctx, frameID, lib.RecipeData{Title: prefix + "-linked", MealCategoryID: mealCategoryID})
	if err != nil {
		t.Fatalf("creating linked recipe: %v", err)
	}

	eventStart := time.Now().AddDate(0, 0, 5).UTC().Truncate(time.Minute)
	data := ExportData{
		Chores:  []lib.Chore{{ID: "1", Title: prefix + "-chore", DueDate: dateFromNow(3), AssigneeID: profileID}},
		Rewards: []lib.Reward{{Title: prefix + "-reward", Points: 1, CategoryID: profileID}},
		Lists:   []lib.List{{Title: prefix + "-list"}},
		Recipes: []lib.Recipe{{Title: prefix + "-recipe", MealCategoryID: mealCategoryID}},
		MealSittings: []lib.MealSitting{
			{Summary: prefix + "-sitting", MealCategoryID: mealCategoryID, Date: dateFromNow(4)},
			{Summary: linkedRecipe.Title, RecipeID: linkedRecipe.ID, MealCategoryID: mealCategoryID, Date: dateFromNow(5)},
		},
		CalendarEvents: []lib.CalendarEvent{{
			Title:   prefix + "-event",
			StartAt: eventStart.Format(time.RFC3339),
			EndAt:   eventStart.Add(time.Hour).Format(time.RFC3339),
		}},
		Bounties: []lib.Bounty{{
			Chore:  lib.Chore{Title: prefix + "-bounty", Points: 1, DueDate: dateFromNow(3), AssigneeID: profileID},
			Reward: lib.Reward{Title: prefix + "-bounty-reward", Points: 1, CategoryID: profileID},
		}},
	}
	want := toWantMap([]string{
		exportResourceChores, exportResourceRewards, exportResourceLists, exportResourceRecipes,
		exportResourceSittings, exportResourceCalendar, exportResourceBounties,
	})

	if err := runImport(ctx, client, data, want, dateFromNow(0), false); err != nil {
		t.Fatalf("first import: %v", err)
	}

	check := func(stage string, found prefixedItems) {
		t.Helper()
		counts := []struct {
			what      string
			got, want int
		}{
			{"chores", len(found.chores), 1},
			{"rewards", len(found.rewards), 1},
			{"lists", len(found.lists), 1},
			{"recipes", len(found.recipes), 2},
			{"sittings", len(found.sittings), 2},
			{"events", len(found.events), 1},
		}
		for _, c := range counts {
			if c.got != c.want {
				t.Errorf("%s: %s = %d, want %d", stage, c.what, c.got, c.want)
			}
		}
		for _, c := range found.chores {
			if strings.Contains(c.Title, "-bounty") {
				t.Errorf("%s: bounty imported as well as its chore: %q", stage, c.Title)
			}
		}
	}

	// Writes can take a moment to appear in listings.
	var found prefixedItems
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(2 * time.Second)
		}
		found = findPrefixed(ctx, t, client, prefix)
		if len(found.chores) >= 1 && len(found.sittings) >= 2 && len(found.events) >= 1 {
			break
		}
	}
	check("after import", found)

	if err := runImport(ctx, client, data, want, dateFromNow(0), true); err != nil {
		t.Fatalf("skip-existing import: %v", err)
	}
	time.Sleep(2 * time.Second)
	check("after skip-existing", findPrefixed(ctx, t, client, prefix))
}

// TestIntegration_TemplateApplyAndSave applies a template holding a recurring
// chore and a reward, then saves one back from the frame, which must keep the
// series as one chore and record the reward's category (#553). The template
// is written by hand, because applying one saved from the frame would copy
// every real chore. HOME is a temp dir so real templates are untouched.
func TestIntegration_TemplateApplyAndSave(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	profileID, _ := firstProfileAndMealCategory(ctx, t, client)
	categoryID, err := strconv.Atoi(profileID)
	if err != nil {
		t.Fatalf("category ID is not an integer: %v", err)
	}

	prefix := fmt.Sprintf("%s%d-template", integrationTestPrefix, time.Now().UnixMilli())
	t.Cleanup(func() { deletePrefixed(ctx, t, client, prefix) })
	t.Setenv("HOME", t.TempDir())

	dir, err := templateDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(frameTemplate{
		Name: "applied",
		Chores: []choreTemplate{{
			Title:         prefix + "-chore",
			AssigneeID:    profileID,
			DueDate:       dateFromNow(2),
			RecurrenceSet: []string{"RRULE:FREQ=WEEKLY;INTERVAL=1"},
		}},
		Rewards: []rewardTemplate{{Title: prefix + "-reward", Points: 1, CategoryIDs: []int{categoryID}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "applied.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	origName, origResources, origStart, origDryRun := templateName, templateResources, templateStartDate, templateDryRun
	t.Cleanup(func() {
		templateName, templateResources, templateStartDate, templateDryRun = origName, origResources, origStart, origDryRun
	})
	templateResources, templateStartDate, templateDryRun = nil, "", false

	templateName = "applied"
	templateApplyCmd.SetContext(ctx)
	if err := templateApplyCmd.RunE(templateApplyCmd, nil); err != nil {
		t.Fatalf("template apply: %v", err)
	}
	time.Sleep(2 * time.Second)

	templateName = "saved"
	templateSaveCmd.SetContext(ctx)
	if err := templateSaveCmd.RunE(templateSaveCmd, nil); err != nil {
		t.Fatalf("template save: %v", err)
	}

	raw, err = os.ReadFile(filepath.Join(dir, "saved.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved frameTemplate
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}

	chores := 0
	for _, c := range saved.Chores {
		if strings.HasPrefix(c.Title, prefix) {
			chores++
			if len(c.RecurrenceSet) == 0 {
				t.Errorf("saved chore lost its recurrence: %+v", c)
			}
		}
	}
	if chores != 1 {
		t.Errorf("saved %d chores for the series, want 1", chores)
	}

	rewards := 0
	for _, r := range saved.Rewards {
		if strings.HasPrefix(r.Title, prefix) {
			rewards++
			if len(r.CategoryIDs) == 0 {
				t.Errorf("saved reward lost its category: %+v", r)
			}
		}
	}
	if rewards != 1 {
		t.Errorf("saved %d rewards, want 1", rewards)
	}
}
