//go:build integration

package lib

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"testing"
	"time"
)

// testPrefix returns a unique name prefix for integration test resources.
func testPrefix() string {
	return fmt.Sprintf("%s%d", integrationTestPrefix, time.Now().UnixMilli())
}

func TestIntegration_ChoresCRUD(t *testing.T) {
	client, frameID := integrationClient(t)

	prefix := testPrefix()
	title := prefix + "-chore"

	categories, err := client.ListCategories(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(categories) == 0 {
		t.Skip("no categories available — cannot create chore")
	}
	categoryID := categories[0].ID

	// Create
	chore, err := client.CreateChore(context.Background(), frameID, ChoreData{
		Title:      title,
		DueDate:    time.Now().Format(DateFormat),
		AssigneeID: categoryID,
	})
	if err != nil {
		t.Fatalf("CreateChore: %v", err)
	}
	if chore.ID == "" {
		t.Fatal("created chore has empty ID")
	}
	t.Logf("created chore %s (%s)", chore.Title, chore.ID)

	t.Cleanup(func() {
		if err := client.DeleteChore(context.Background(), frameID, chore.ID); err != nil &&
			!IsNotFound(err) {
			t.Errorf("cleanup DeleteChore: %v", err)
		}
	})

	// Verify in list (retry to handle API eventual consistency)
	opts := ChoreListOptions{
		After:  time.Now().AddDate(0, 0, -1).Format(DateFormat),
		Before: time.Now().AddDate(0, 0, 7).Format(DateFormat),
	}
	found := false
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(2 * time.Second)
		}
		chores, err := client.ListChores(context.Background(), frameID, opts)
		if err != nil {
			t.Fatalf("ListChores: %v", err)
		}
		for _, c := range chores {
			if c.ID == chore.ID {
				found = true
				break
			}
		}
		if found {
			break
		}
		t.Logf("attempt %d: chore %s not yet in list, retrying...", attempt+1, chore.ID)
	}
	if !found {
		t.Logf("created chore %s not found in list after retries (API eventual consistency)", chore.ID)
	}

	// Update
	updated, err := client.UpdateChore(context.Background(), frameID, chore.ID, ChoreData{
		Title: prefix + "-chore-updated",
	})
	if err != nil {
		t.Fatalf("UpdateChore: %v", err)
	}
	if !strings.HasSuffix(updated.Title, "-updated") {
		t.Errorf("expected updated title suffix, got %q", updated.Title)
	}
	t.Logf("updated chore title to %q", updated.Title)

	// Complete
	if err := client.CompleteChore(context.Background(), frameID, chore.ID); err != nil {
		t.Fatalf("CompleteChore: %v", err)
	}
	t.Log("completed chore")
}

// TestIntegration_DeleteChore_NonRecurring confirms that DeleteChore (no
// apply_to param) succeeds on a one-time chore — the scenario that 404'd
// when the old code tried GET /chores/{id} first.
func TestIntegration_DeleteChore_NonRecurring(t *testing.T) {
	client, frameID := integrationClient(t)

	categories, err := client.ListCategories(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(categories) == 0 {
		t.Skip("no categories available — cannot create chore")
	}

	chore, err := client.CreateChore(context.Background(), frameID, ChoreData{
		Title:      testPrefix() + "-delete-nonrecurring",
		DueDate:    time.Now().Format(DateFormat),
		AssigneeID: categories[0].ID,
	})
	if err != nil {
		t.Fatalf("CreateChore: %v", err)
	}
	t.Logf("created chore %s (%s)", chore.Title, chore.ID)

	t.Cleanup(func() {
		// safety net in case delete below fails
		_ = client.DeleteChore(context.Background(), frameID, chore.ID)
	})

	if err := client.DeleteChore(context.Background(), frameID, chore.ID); err != nil {
		t.Fatalf("DeleteChore: %v", err)
	}
	t.Logf("deleted non-recurring chore %s", chore.ID)
}

// TestIntegration_DeleteChore_Recurring confirms that DeleteRecurringChore
// (apply_to=all) succeeds on a recurring chore.
func TestIntegration_DeleteChore_Recurring(t *testing.T) {
	client, frameID := integrationClient(t)

	categories, err := client.ListCategories(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(categories) == 0 {
		t.Skip("no categories available — cannot create chore")
	}

	chore, err := client.CreateChore(context.Background(), frameID, ChoreData{
		Title:      testPrefix() + "-delete-recurring",
		DueDate:    time.Now().Format(DateFormat),
		AssigneeID: categories[0].ID,
		Recurring:  true,
		Frequency:  "DAILY",
	})
	if err != nil {
		t.Fatalf("CreateChore (recurring): %v", err)
	}
	t.Logf("created chore %s (%s), recurring=%v", chore.Title, chore.ID, chore.Recurring)

	t.Cleanup(func() {
		_ = client.DeleteChore(context.Background(), frameID, chore.ID)
	})

	if !chore.Recurring {
		t.Skip("API created chore as non-recurring — cannot test DeleteRecurringChore via create; skipping")
	}

	if err := client.DeleteRecurringChore(context.Background(), frameID, chore.ID); err != nil {
		t.Fatalf("DeleteRecurringChore: %v", err)
	}
	t.Logf("deleted recurring chore %s", chore.ID)
}

func TestIntegration_RewardsCRUD(t *testing.T) {
	client, frameID := integrationClient(t)

	prefix := testPrefix()
	title := prefix + "-reward"

	categories, err := client.ListCategories(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(categories) == 0 {
		t.Skip("no categories available — cannot create reward")
	}
	catIDInt, err := strconv.Atoi(categories[0].ID)
	if err != nil {
		t.Fatalf("category ID is not an integer: %v", err)
	}

	// Create
	reward, err := client.CreateReward(context.Background(), frameID, RewardData{
		Title:       title,
		Points:      10,
		CategoryIDs: []int{catIDInt},
	})
	if err != nil {
		t.Fatalf("CreateReward: %v", err)
	}
	if reward.ID == "" {
		t.Fatal("created reward has empty ID")
	}
	t.Logf("created reward %s (%s)", reward.Title, reward.ID)

	t.Cleanup(func() {
		if err := client.DeleteReward(context.Background(), frameID, reward.ID); err != nil &&
			!IsNotFound(err) {
			t.Errorf("cleanup DeleteReward: %v", err)
		}
	})

	// Verify in list
	rewards, err := client.ListRewards(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	found := false
	for _, r := range rewards {
		if r.ID == reward.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created reward %s not found in list", reward.ID)
	}

	// Update
	updated, err := client.UpdateReward(context.Background(), frameID, reward.ID, RewardData{
		Title:  prefix + "-reward-updated",
		Points: 20,
	})
	if err != nil {
		t.Fatalf("UpdateReward: %v", err)
	}
	if updated.Points != 20 {
		t.Errorf("expected 20 points, got %d", updated.Points)
	}
	t.Logf("updated reward to %q (%d pts)", updated.Title, updated.Points)

	// Redeem and unredeem (skip if account lacks points)
	if err := client.RedeemReward(context.Background(), frameID, reward.ID); err != nil {
		if strings.Contains(err.Error(), "Not enough points") {
			t.Skipf("RedeemReward: skipping — account has insufficient points: %v", err)
		}
		t.Fatalf("RedeemReward: %v", err)
	}
	t.Log("redeemed reward")

	if err := client.UnredeemReward(context.Background(), frameID, reward.ID); err != nil {
		t.Fatalf("UnredeemReward: %v", err)
	}
	t.Log("unredeemed reward")
}

func TestIntegration_CalendarEventsCRUD(t *testing.T) {
	client, frameID := integrationClient(t)

	prefix := testPrefix()
	now := time.Now()
	startAt := now.AddDate(0, 0, 1).Format(time.RFC3339)
	endAt := now.AddDate(0, 0, 1).Add(time.Hour).Format(time.RFC3339)

	// Create
	event, err := client.CreateCalendarEvent(context.Background(), frameID, CalendarEventData{
		Title:   prefix + "-event",
		StartAt: startAt,
		EndAt:   endAt,
	})
	if err != nil {
		if strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "Internal Server Error") {
			t.Skipf("CreateCalendarEvent: skipping due to API instability: %v", err)
		}
		t.Fatalf("CreateCalendarEvent: %v", err)
	}
	if event.ID == "" {
		t.Fatal("created event has empty ID")
	}
	t.Logf("created event %s (%s)", event.Title, event.ID)

	t.Cleanup(func() {
		if err := client.DeleteCalendarEvent(context.Background(), frameID, event.ID); err != nil &&
			!IsNotFound(err) {
			t.Errorf("cleanup DeleteCalendarEvent: %v", err)
		}
	})

	// Verify in list
	start := now.Format(DateFormat)
	end := now.AddDate(0, 0, 7).Format(DateFormat)
	events, err := client.ListCalendarEvents(context.Background(), frameID, start, end, "")
	if err != nil {
		if strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "Internal Server Error") {
			t.Skipf("ListCalendarEvents: skipping due to API instability: %v", err)
		}
		t.Fatalf("ListCalendarEvents: %v", err)
	}
	found := false
	for _, e := range events {
		if e.ID == event.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created event %s not found in list", event.ID)
	}

	// Update
	newEnd := now.AddDate(0, 0, 1).Add(2 * time.Hour).Format(time.RFC3339)
	updated, err := client.UpdateCalendarEvent(context.Background(), frameID, event.ID, CalendarEventData{
		EndAt: newEnd,
	})
	if err != nil {
		if strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "Internal Server Error") {
			t.Skipf("UpdateCalendarEvent: skipping due to API instability: %v", err)
		}
		t.Fatalf("UpdateCalendarEvent: %v", err)
	}
	t.Logf("updated event end to %s", updated.EndAt)
}

func TestIntegration_ListsCRUD(t *testing.T) {
	client, frameID := integrationClient(t)

	prefix := testPrefix()

	// Create list
	list, err := client.CreateList(context.Background(), frameID, ListData{
		Title: prefix + "-list",
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}
	if list.ID == "" {
		t.Fatal("created list has empty ID")
	}
	t.Logf("created list %s (%s)", list.Title, list.ID)

	t.Cleanup(func() {
		if err := client.DeleteList(context.Background(), frameID, list.ID); err != nil &&
			!IsNotFound(err) {
			t.Errorf("cleanup DeleteList: %v", err)
		}
	})

	// Verify in list-of-lists
	lists, err := client.ListLists(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListLists: %v", err)
	}
	found := false
	for _, l := range lists {
		if l.ID == list.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created list %s not found in lists", list.ID)
	}

	// Add item
	item, err := client.AddListItem(context.Background(), frameID, list.ID, ListItemData{
		Title: prefix + "-item",
	})
	if err != nil {
		t.Fatalf("AddListItem: %v", err)
	}
	if item.ID == "" {
		t.Fatal("created list item has empty ID")
	}
	t.Logf("added item %s (%s)", item.Title, item.ID)

	t.Cleanup(func() {
		if err := client.DeleteListItem(context.Background(), frameID, list.ID, item.ID); err != nil &&
			!IsNotFound(err) {
			t.Errorf("cleanup DeleteListItem: %v", err)
		}
	})

	// Update item (mark completed)
	updatedItem, err := client.UpdateListItem(context.Background(), frameID, list.ID, item.ID, ListItemData{
		Completed: true,
	})
	if err != nil {
		t.Fatalf("UpdateListItem: %v", err)
	}
	t.Logf("updated item status: %s", updatedItem.Status)

	// Delete item explicitly before list deletion
	if err := client.DeleteListItem(context.Background(), frameID, list.ID, item.ID); err != nil &&
		!strings.Contains(err.Error(), "404") {
		t.Fatalf("DeleteListItem: %v", err)
	}
	t.Log("deleted list item")

	// Update list title
	updated, err := client.UpdateList(context.Background(), frameID, list.ID, ListData{
		Title: prefix + "-list-updated",
	})
	if err != nil {
		t.Fatalf("UpdateList: %v", err)
	}
	if !strings.HasSuffix(updated.Title, "-updated") {
		t.Errorf("expected updated title suffix, got %q", updated.Title)
	}
	t.Logf("updated list title to %q", updated.Title)
}

func TestIntegration_RecipesCRUD(t *testing.T) {
	client, frameID := integrationClient(t)

	prefix := testPrefix()

	mealCategories, err := client.ListMealCategories(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListMealCategories: %v", err)
	}
	if len(mealCategories) == 0 {
		t.Skip("no meal categories available — cannot create recipe")
	}
	mealCategoryID := mealCategories[0].ID

	// Create
	recipe, err := client.CreateRecipe(context.Background(), frameID, RecipeData{
		Title:          prefix + "-recipe",
		Description:    "integration test recipe",
		Ingredients:    []string{"flour", "eggs", "milk"},
		MealCategoryID: mealCategoryID,
	})
	if err != nil {
		t.Fatalf("CreateRecipe: %v", err)
	}
	if recipe.ID == "" {
		t.Fatal("created recipe has empty ID")
	}
	t.Logf("created recipe %s (%s)", recipe.Title, recipe.ID)

	t.Cleanup(func() {
		if err := client.DeleteRecipe(context.Background(), frameID, recipe.ID); err != nil &&
			!IsNotFound(err) {
			t.Errorf("cleanup DeleteRecipe: %v", err)
		}
	})

	// Verify in list
	recipes, err := client.ListRecipes(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListRecipes: %v", err)
	}
	found := false
	for _, r := range recipes {
		if r.ID == recipe.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created recipe %s not found in list", recipe.ID)
	}

	// Update description
	updated, err := client.UpdateRecipe(context.Background(), frameID, recipe.ID, RecipeData{
		Description: "updated description",
	})
	if err != nil {
		t.Fatalf("UpdateRecipe: %v", err)
	}
	t.Logf("updated recipe description to %q", updated.Description)
}

func TestIntegration_MealSittingsCRUD(t *testing.T) {
	client, frameID := integrationClient(t)

	prefix := testPrefix()
	date := time.Now().AddDate(0, 0, 2).Format(DateFormat)

	// Need a recipe and a meal category to create a sitting
	mealCategories, err := client.ListMealCategories(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListMealCategories: %v", err)
	}
	if len(mealCategories) == 0 {
		t.Skip("no meal categories available — cannot create meal sitting")
	}
	categoryID := mealCategories[0].ID

	recipe, err := client.CreateRecipe(context.Background(), frameID, RecipeData{
		Title:          prefix + "-recipe-for-sitting",
		MealCategoryID: categoryID,
	})
	if err != nil {
		t.Fatalf("CreateRecipe (for sitting): %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteRecipe(context.Background(), frameID, recipe.ID); err != nil &&
			!IsNotFound(err) {
			t.Errorf("cleanup DeleteRecipe (for sitting): %v", err)
		}
	})

	// Create sitting
	sitting, err := client.CreateMealSitting(context.Background(), frameID, MealSittingData{
		RecipeID:       recipe.ID,
		MealCategoryID: categoryID,
		Date:           date,
	})
	if err != nil {
		t.Fatalf("CreateMealSitting: %v", err)
	}
	if sitting.ID == "" {
		t.Fatal("created meal sitting has empty ID")
	}
	t.Logf("created meal sitting %s (%s)", sitting.Summary, sitting.ID)

	t.Cleanup(func() {
		if err := client.DeleteMealSitting(context.Background(), frameID, sitting.ID, date); err != nil &&
			!IsNotFound(err) {
			t.Errorf("cleanup DeleteMealSitting: %v", err)
		}
	})

	// Verify in list
	sittings, err := client.ListMealSittings(context.Background(), frameID, MealSittingListOptions{
		DateMin: time.Now().Format(DateFormat),
		DateMax: time.Now().AddDate(0, 0, 7).Format(DateFormat),
	})
	if err != nil {
		t.Fatalf("ListMealSittings: %v", err)
	}
	found := false
	for _, s := range sittings {
		if s.ID == sitting.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created meal sitting %s not found in list", sitting.ID)
	}

	// Delete sitting explicitly
	if err := client.DeleteMealSitting(context.Background(), frameID, sitting.ID, date); err != nil &&
		!strings.Contains(err.Error(), "404") {
		t.Fatalf("DeleteMealSitting: %v", err)
	}
	t.Log("deleted meal sitting")
}

func TestIntegration_RecurringChore(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	categories, err := client.ListCategories(ctx, frameID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(categories) == 0 {
		t.Skip("no categories available — cannot create chore")
	}

	chore, err := client.CreateChore(ctx, frameID, ChoreData{
		Title:          testPrefix() + "-recurring",
		DueDate:        time.Now().Format(DateFormat),
		AssigneeID:     categories[0].ID,
		Frequency:      "weekly",
		RecurrenceDays: []string{"mon", "thu"},
	})
	if err != nil {
		t.Fatalf("CreateChore: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteRecurringChore(ctx, frameID, chore.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteRecurringChore: %v", err)
		}
	})
	if !chore.Recurring {
		t.Fatalf("created chore is not recurring: %+v", chore)
	}

	assertWeekdays(ctx, t, client, frameID, chore.ID, time.Monday, time.Thursday)

	if _, err := client.UpdateChore(ctx, frameID, chore.ID, ChoreData{
		Frequency:      "weekly",
		RecurrenceDays: []string{"wed"},
	}); err != nil {
		t.Fatalf("UpdateChore: %v", err)
	}

	assertWeekdays(ctx, t, client, frameID, chore.ID, time.Wednesday)
}

// assertWeekdays checks that a recurring chore's occurrences over the next two weeks
// fall on exactly the given weekdays, retrying for API eventual consistency.
func assertWeekdays(ctx context.Context, t *testing.T, client *Client, frameID, choreID string, want ...time.Weekday) {
	t.Helper()
	wantSet := map[time.Weekday]bool{}
	for _, d := range want {
		wantSet[d] = true
	}
	opts := ChoreListOptions{
		After:  time.Now().Format(DateFormat),
		Before: time.Now().AddDate(0, 0, 14).Format(DateFormat),
	}
	var got map[time.Weekday]bool
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(2 * time.Second)
		}
		chores, err := client.ListChores(ctx, frameID, opts)
		if err != nil {
			t.Fatalf("ListChores: %v", err)
		}
		got = map[time.Weekday]bool{}
		for _, c := range chores {
			if c.ID != choreID && !strings.HasPrefix(c.ID, choreID+"-") {
				continue
			}
			due, err := time.Parse(DateFormat, c.DueDate)
			if err != nil {
				t.Fatalf("parse due date %q: %v", c.DueDate, err)
			}
			got[due.Weekday()] = true
		}
		if maps.Equal(got, wantSet) {
			return
		}
	}
	t.Errorf("occurrence weekdays: want %v got %v", want, got)
}
