//go:build integration

package lib

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"
	"time"
)

// noteKnownGap reports whether err is the 404 the live API returns for an
// endpoint the library calls but the API doesn't serve. It logs the gap so it
// stays visible, and a working endpoint falls through to the normal checks.
func noteKnownGap(t *testing.T, err error, method string) bool {
	t.Helper()
	if err != nil && IsNotFound(err) {
		t.Logf("known gap: %s got a 404 from the live API: %v", method, err)
		return true
	}
	return false
}

// futureDate returns the date days from now in the API's format.
func futureDate(days int) string {
	return time.Now().AddDate(0, 0, days).Format(DateFormat)
}

// firstProfile returns a household member to assign test items to.
func firstProfile(ctx context.Context, t *testing.T, client *Client, frameID string) Category {
	t.Helper()
	profiles, err := client.ListProfiles(ctx, frameID)
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if len(profiles) == 0 {
		t.Skip("no profiles available")
	}
	return profiles[0]
}

// firstMealCategory returns a meal category for test recipes and sittings.
func firstMealCategory(ctx context.Context, t *testing.T, client *Client, frameID string) string {
	t.Helper()
	categories, err := client.ListMealCategories(ctx, frameID)
	if err != nil {
		t.Fatalf("ListMealCategories: %v", err)
	}
	if len(categories) == 0 {
		t.Skip("no meal categories available")
	}
	return categories[0].ID
}

// createTestRecipe creates a recipe that is deleted when the test ends.
func createTestRecipe(ctx context.Context, t *testing.T, client *Client, frameID, title, mealCategoryID string) *Recipe {
	t.Helper()
	recipe, err := client.CreateRecipe(ctx, frameID, RecipeData{Title: title, MealCategoryID: mealCategoryID})
	if err != nil {
		t.Fatalf("CreateRecipe: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteRecipe(ctx, frameID, recipe.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteRecipe: %v", err)
		}
	})
	return recipe
}

// createTestList creates a list that is deleted when the test ends.
func createTestList(ctx context.Context, t *testing.T, client *Client, frameID, title string) *List {
	t.Helper()
	list, err := client.CreateList(ctx, frameID, ListData{Title: title})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteList(ctx, frameID, list.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteList: %v", err)
		}
	})
	return list
}

// cleanupChore deletes a one-time chore when the test ends.
func cleanupChore(ctx context.Context, t *testing.T, client *Client, frameID, choreID string) {
	t.Helper()
	t.Cleanup(func() {
		if err := client.DeleteChore(ctx, frameID, choreID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteChore: %v", err)
		}
	})
}

func TestIntegration_ListFrames(t *testing.T) {
	client, frameID := integrationClient(t)

	frames, err := client.ListFrames(context.Background())
	if err != nil {
		t.Fatalf("ListFrames: %v", err)
	}
	for _, f := range frames {
		if f.ID == frameID {
			return
		}
	}
	t.Errorf("frame %s not in ListFrames (%d frames)", frameID, len(frames))
}

func TestIntegration_ListLabels(t *testing.T) {
	client, frameID := integrationClient(t)

	labels, err := client.ListLabels(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListLabels: %v", err)
	}
	for _, l := range labels {
		if l.LinkedToProfile {
			t.Errorf("label %s is linked to a profile", l.ID)
		}
	}
}

func TestIntegration_ListRoutines(t *testing.T) {
	client, frameID := integrationClient(t)

	routines, err := client.ListRoutines(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListRoutines: %v", err)
	}
	for _, r := range routines {
		if r.ID == "" || r.Title == "" {
			t.Errorf("routine missing ID or title: %+v", r)
		}
	}
}

// TestIntegration_ListPhotos pins a live API behaviour: next_page_token keeps
// being returned after the last page, so only an empty page ends the list.
func TestIntegration_ListPhotos(t *testing.T) {
	client, frameID := integrationClient(t)

	const maxPages = 200
	total, token := 0, ""
	for page := 0; page < maxPages; page++ {
		photos, next, err := client.ListPhotos(context.Background(), frameID, PhotoListOptions{PageToken: token})
		if err != nil {
			t.Fatalf("ListPhotos: %v", err)
		}
		total += len(photos)
		if next == "" || len(photos) == 0 {
			t.Logf("%d photos over %d pages", total, page+1)
			return
		}
		token = next
	}
	t.Fatalf("paging did not end within %d pages", maxPages)
}

func TestIntegration_ListAlbums(t *testing.T) {
	client, frameID := integrationClient(t)

	albums, err := client.ListAlbums(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListAlbums: %v", err)
	}
	for _, a := range albums {
		if a.ID == "" {
			t.Error("album has empty ID")
		}
	}
}

func TestIntegration_ListSourceCalendars(t *testing.T) {
	client, frameID := integrationClient(t)

	sources, err := client.ListSourceCalendars(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListSourceCalendars: %v", err)
	}
	t.Logf("source calendars: %d", len(sources))
}

func TestIntegration_ListBounties(t *testing.T) {
	client, frameID := integrationClient(t)

	bounties, err := client.ListBounties(context.Background(), frameID, BountyListOptions{})
	if err != nil {
		t.Fatalf("ListBounties: %v", err)
	}
	for _, b := range bounties {
		if b.Chore.ID == "" || b.Reward.ID == "" {
			t.Errorf("bounty missing its chore or reward: %+v", b)
		}
	}
}

func TestIntegration_GetAvatars(t *testing.T) {
	client, _ := integrationClient(t)

	avatars, err := client.GetAvatars(context.Background())
	if err != nil {
		t.Fatalf("GetAvatars: %v", err)
	}
	if len(avatars) == 0 {
		t.Error("expected at least one avatar")
	}
}

func TestIntegration_GetColors(t *testing.T) {
	client, _ := integrationClient(t)

	colors, err := client.GetColors(context.Background())
	if err != nil {
		t.Fatalf("GetColors: %v", err)
	}
	if len(colors) == 0 {
		t.Error("expected at least one color")
	}
}

func TestIntegration_GetDeviceConfig(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	devices, err := client.ListDevices(ctx, frameID)
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(devices) == 0 {
		t.Skip("no devices on this frame")
	}
	if _, err := client.GetDeviceConfig(ctx, frameID, devices[0].ID); err != nil {
		t.Fatalf("GetDeviceConfig: %v", err)
	}
}

func TestIntegration_GetRewardPoints(t *testing.T) {
	client, frameID := integrationClient(t)

	points, err := client.GetRewardPoints(context.Background(), frameID)
	if err != nil {
		t.Fatalf("GetRewardPoints: %v", err)
	}
	t.Logf("point balances: %d", len(points))
}

func TestIntegration_GetPhotoDownload(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	photos, _, err := client.ListPhotos(ctx, frameID, PhotoListOptions{})
	if err != nil {
		t.Fatalf("ListPhotos: %v", err)
	}
	if len(photos) == 0 || photos[0].ThumbnailURL == "" {
		t.Skip("no photo thumbnail available")
	}

	data, err := client.DownloadPhoto(ctx, photos[0].ThumbnailURL)
	if err != nil {
		t.Fatalf("DownloadPhoto: %v", err)
	}
	if len(data) == 0 {
		t.Error("DownloadPhoto returned no bytes")
	}

	var buf bytes.Buffer
	n, err := client.DownloadPhotoTo(ctx, photos[0].ThumbnailURL, &buf)
	if err != nil {
		t.Fatalf("DownloadPhotoTo: %v", err)
	}
	if n != int64(buf.Len()) || n != int64(len(data)) {
		t.Errorf("DownloadPhotoTo wrote %d bytes, buffer %d, DownloadPhoto %d", n, buf.Len(), len(data))
	}
}

func TestIntegration_RoutinesCRUD(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	profile := firstProfile(ctx, t, client, frameID)
	title := testPrefix() + "-routine"

	routine, err := client.CreateRoutine(ctx, frameID, RoutineData{
		Title:      title,
		TimeOfDay:  RoutineTODMorning,
		CategoryID: profile.ID,
		StartDate:  futureDate(1),
	})
	if err != nil {
		t.Fatalf("CreateRoutine: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteRoutine(ctx, frameID, routine.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteRoutine: %v", err)
		}
	})

	got, err := client.GetRoutine(ctx, frameID, routine.ID)
	if err != nil {
		t.Fatalf("GetRoutine: %v", err)
	}
	if got.Title != title || got.TimeOfDay != RoutineTODMorning {
		t.Errorf("GetRoutine = %+v", got)
	}

	updated, err := client.UpdateRoutine(ctx, frameID, routine.ID, RoutineUpdateData{Title: title + "-updated"})
	if err != nil {
		t.Fatalf("UpdateRoutine: %v", err)
	}
	if updated.Title != title+"-updated" {
		t.Errorf("updated title = %q", updated.Title)
	}

	if err := client.DeleteRoutine(ctx, frameID, routine.ID); err != nil {
		t.Fatalf("DeleteRoutine: %v", err)
	}
}

func TestIntegration_BountiesCRUD(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	profile := firstProfile(ctx, t, client, frameID)
	categoryID, err := strconv.Atoi(profile.ID)
	if err != nil {
		t.Fatalf("category ID is not an integer: %v", err)
	}
	prefix := testPrefix()

	bounty, err := client.CreateBounty(ctx, frameID, BountyData{
		Title:       prefix + "-bounty",
		Points:      7,
		DueDate:     futureDate(3),
		AssigneeID:  profile.ID,
		CategoryIDs: []int{categoryID},
		RewardTitle: prefix + "-bounty-reward",
	})
	if err != nil {
		t.Fatalf("CreateBounty: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteBounty(ctx, frameID, bounty.Chore.ID, bounty.Reward.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteBounty: %v", err)
		}
	})

	listed, err := client.ListBounties(ctx, frameID, BountyListOptions{After: futureDate(0), Before: futureDate(7)})
	if err != nil {
		t.Fatalf("ListBounties: %v", err)
	}
	found := false
	for _, b := range listed {
		found = found || b.Chore.ID == bounty.Chore.ID
	}
	if !found {
		t.Errorf("bounty chore %s not listed", bounty.Chore.ID)
	}

	updated, err := client.UpdateBounty(ctx, frameID, bounty.Chore.ID, bounty.Reward.ID, BountyData{
		Title:       prefix + "-bounty-updated",
		Points:      8,
		RewardTitle: prefix + "-bounty-reward-updated",
	})
	if err != nil {
		t.Fatalf("UpdateBounty: %v", err)
	}
	if updated.Chore.Title != prefix+"-bounty-updated" {
		t.Errorf("updated chore title = %q", updated.Chore.Title)
	}

	if err := client.DeleteBounty(ctx, frameID, bounty.Chore.ID, bounty.Reward.ID); err != nil {
		t.Fatalf("DeleteBounty: %v", err)
	}
}

func TestIntegration_CategoriesCRUD(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	profile := firstProfile(ctx, t, client, frameID)
	name := testPrefix() + "-category"
	notProfile := false

	category, err := client.CreateCategory(ctx, frameID, CategoryData{Name: name, Color: profile.Color, LinkedToProfile: &notProfile})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteCategory(ctx, frameID, category.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteCategory: %v", err)
		}
	})

	updated, err := client.UpdateCategory(ctx, frameID, category.ID, CategoryData{Name: name + "-updated"})
	if err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}
	if updated.Name != name+"-updated" {
		t.Errorf("updated name = %q", updated.Name)
	}

	if err := client.DeleteCategory(ctx, frameID, category.ID); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
}

func TestIntegration_MealCategoriesCRUD(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	profile := firstProfile(ctx, t, client, frameID)
	name := testPrefix() + "-meal-category"

	category, err := client.CreateMealCategory(ctx, frameID, MealCategoryData{Name: name, Color: profile.Color})
	if noteKnownGap(t, err, "CreateMealCategory") {
		t.Skip("CreateMealCategory is not served by the live API")
	}
	if err != nil {
		t.Fatalf("CreateMealCategory: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteMealCategory(ctx, frameID, category.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteMealCategory: %v", err)
		}
	})

	updated, err := client.UpdateMealCategory(ctx, frameID, category.ID, MealCategoryData{Name: name + "-updated"})
	if err != nil {
		t.Fatalf("UpdateMealCategory: %v", err)
	}
	if updated.Name != name+"-updated" {
		t.Errorf("updated name = %q", updated.Name)
	}

	if err := client.DeleteMealCategory(ctx, frameID, category.ID); err != nil {
		t.Fatalf("DeleteMealCategory: %v", err)
	}
}

func TestIntegration_UpForGrabsChoreClaim(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	profile := firstProfile(ctx, t, client, frameID)

	chore, err := client.CreateUpForGrabsChore(ctx, frameID, ChoreData{
		Title:   testPrefix() + "-grabs",
		DueDate: futureDate(1),
	})
	if err != nil {
		t.Fatalf("CreateUpForGrabsChore: %v", err)
	}
	cleanupChore(ctx, t, client, frameID, chore.ID)

	claimed, err := client.ClaimChore(ctx, frameID, chore.ID, profile.ID)
	if err != nil {
		t.Fatalf("ClaimChore: %v", err)
	}
	if claimed.AssigneeID != profile.ID {
		t.Errorf("claimed assignee = %q, want %q", claimed.AssigneeID, profile.ID)
	}
}

// TestIntegration_SkipChore skips one occurrence of a repeating chore. The API
// rejects a skip on a one-time chore.
func TestIntegration_SkipChore(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	profile := firstProfile(ctx, t, client, frameID)
	chore, err := client.CreateChore(ctx, frameID, ChoreData{
		Title:      testPrefix() + "-skip",
		DueDate:    futureDate(1),
		AssigneeID: profile.ID,
		Frequency:  "daily",
	})
	if err != nil {
		t.Fatalf("CreateChore: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteRecurringChore(ctx, frameID, chore.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteRecurringChore: %v", err)
		}
	})

	// A repeating chore is skipped per occurrence, addressed as <id>-<date>.
	baseID, _, _ := strings.Cut(chore.ID, "-")
	if err := client.SkipChore(ctx, frameID, baseID+"-"+futureDate(1), ""); err != nil {
		t.Fatalf("SkipChore: %v", err)
	}
}

func TestIntegration_ChoreRotation(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	profile := firstProfile(ctx, t, client, frameID)

	result, err := client.CreateChoreRotation(ctx, frameID, RotationData{
		Chores:      []string{testPrefix() + "-rotation"},
		AssigneeIDs: []string{profile.ID},
		StartDate:   futureDate(14),
		Weeks:       2,
	})
	if result != nil {
		for _, c := range result.Chores {
			cleanupChore(ctx, t, client, frameID, c.ID)
		}
	}
	if err != nil {
		t.Fatalf("CreateChoreRotation: %v", err)
	}
	if len(result.Chores) != 2 {
		t.Errorf("created %d chores, want 2", len(result.Chores))
	}
}

func TestIntegration_ClearCompletedListItems(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	prefix := testPrefix()
	list := createTestList(ctx, t, client, frameID, prefix+"-clear")

	done, err := client.AddListItem(ctx, frameID, list.ID, ListItemData{Title: prefix + "-done"})
	if err != nil {
		t.Fatalf("AddListItem: %v", err)
	}
	if _, err := client.AddListItem(ctx, frameID, list.ID, ListItemData{Title: prefix + "-open"}); err != nil {
		t.Fatalf("AddListItem: %v", err)
	}
	if _, err := client.UpdateListItem(ctx, frameID, list.ID, done.ID, ListItemData{Completed: true}); err != nil {
		t.Fatalf("UpdateListItem: %v", err)
	}

	cleared, err := client.ClearCompletedListItems(ctx, frameID, list.ID)
	if err != nil {
		t.Fatalf("ClearCompletedListItems: %v", err)
	}
	if cleared != 1 {
		t.Errorf("cleared %d items, want 1", cleared)
	}
}

func TestIntegration_PlanMeals(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	mealCategoryID := firstMealCategory(ctx, t, client, frameID)
	recipe := createTestRecipe(ctx, t, client, frameID, testPrefix()+"-plan", mealCategoryID)
	start := futureDate(3)

	result, err := client.PlanMeals(ctx, frameID, MealPlanData{
		RecipeIDs:   []string{recipe.ID},
		CategoryIDs: []string{mealCategoryID},
		StartDate:   start,
	})
	if result != nil {
		for _, s := range result.Sittings {
			t.Cleanup(func() {
				if err := client.DeleteMealSitting(ctx, frameID, s.ID, start); err != nil && !IsNotFound(err) {
					t.Errorf("cleanup DeleteMealSitting: %v", err)
				}
			})
		}
	}
	if err != nil {
		t.Fatalf("PlanMeals: %v", err)
	}
	if len(result.Sittings) != 1 {
		t.Errorf("created %d sittings, want 1", len(result.Sittings))
	}
}

// createSummarySitting creates a sitting with a summary and no recipe, deleted
// when the test ends.
func createSummarySitting(ctx context.Context, t *testing.T, client *Client, frameID, summary, mealCategoryID, date string) *MealSitting {
	t.Helper()
	sitting, err := client.CreateMealSitting(ctx, frameID, MealSittingData{Summary: summary, MealCategoryID: mealCategoryID, Date: date})
	if err != nil {
		t.Fatalf("CreateMealSitting: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteMealSitting(ctx, frameID, sitting.ID, date); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteMealSitting: %v", err)
		}
	})
	return sitting
}

func TestIntegration_UpdateMealSitting(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	mealCategoryID := firstMealCategory(ctx, t, client, frameID)
	summary := testPrefix() + "-sitting"
	sitting := createSummarySitting(ctx, t, client, frameID, summary, mealCategoryID, futureDate(4))

	updated, err := client.UpdateMealSitting(ctx, frameID, sitting.ID, MealSittingData{Summary: summary + "-updated"})
	if noteKnownGap(t, err, "UpdateMealSitting") {
		t.Skip("UpdateMealSitting is not served by the live API")
	}
	if err != nil {
		t.Fatalf("UpdateMealSitting: %v", err)
	}
	if updated.Summary != summary+"-updated" {
		t.Errorf("updated summary = %q", updated.Summary)
	}
}

// TestIntegration_MealSittingDateMaxExclusive pins a live API behaviour: a
// sitting on date_max is not returned.
func TestIntegration_MealSittingDateMaxExclusive(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	mealCategoryID := firstMealCategory(ctx, t, client, frameID)
	dayBefore, day, dayAfter := futureDate(5), futureDate(6), futureDate(7)
	sitting := createSummarySitting(ctx, t, client, frameID, testPrefix()+"-exclusive", mealCategoryID, day)

	listed := func(dateMax string) bool {
		t.Helper()
		sittings, err := client.ListMealSittings(ctx, frameID, MealSittingListOptions{DateMin: dayBefore, DateMax: dateMax})
		if err != nil {
			t.Fatalf("ListMealSittings: %v", err)
		}
		for _, s := range sittings {
			if s.ID == sitting.ID {
				return true
			}
		}
		return false
	}

	if listed(day) {
		t.Error("sitting on date_max was returned, expected date_max to be exclusive")
	}
	if !listed(dayAfter) {
		t.Error("sitting before date_max was not returned")
	}
}

// TestIntegration_CalendarDateMaxExclusive pins the same behaviour for events.
func TestIntegration_CalendarDateMaxExclusive(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	frame, err := client.GetFrame(ctx, frameID)
	if err != nil {
		t.Fatalf("GetFrame: %v", err)
	}
	loc, err := time.LoadLocation(frame.TimeZone)
	if err != nil {
		t.Skipf("frame timezone %q not loadable: %v", frame.TimeZone, err)
	}

	day := time.Now().In(loc).AddDate(0, 0, 6)
	noon := time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, loc)
	event, err := client.CreateCalendarEvent(ctx, frameID, CalendarEventData{
		Title:   testPrefix() + "-exclusive",
		StartAt: noon.Format(time.RFC3339),
		EndAt:   noon.Add(time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("CreateCalendarEvent: %v", err)
	}
	t.Cleanup(func() {
		if err := client.DeleteCalendarEvent(ctx, frameID, event.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup DeleteCalendarEvent: %v", err)
		}
	})

	listed := func(dateMax time.Time) bool {
		t.Helper()
		events, err := client.ListCalendarEvents(ctx, frameID, noon.AddDate(0, 0, -1).Format(DateFormat), dateMax.Format(DateFormat), frame.TimeZone)
		if err != nil {
			t.Fatalf("ListCalendarEvents: %v", err)
		}
		for _, e := range events {
			if e.ID == event.ID {
				return true
			}
		}
		return false
	}

	if listed(noon) {
		t.Error("event on date_max was returned, expected date_max to be exclusive")
	}
	if !listed(noon.AddDate(0, 0, 1)) {
		t.Error("event before date_max was not returned")
	}
}

// TestIntegration_SittingSummaryWithRecipeRejected pins a live API behaviour:
// a sitting takes its name from the recipe, so sending both returns a 422.
func TestIntegration_SittingSummaryWithRecipeRejected(t *testing.T) {
	client, frameID := integrationClient(t)
	ctx := context.Background()

	mealCategoryID := firstMealCategory(ctx, t, client, frameID)
	prefix := testPrefix()
	recipe := createTestRecipe(ctx, t, client, frameID, prefix+"-recipe", mealCategoryID)
	date := futureDate(5)

	sitting, err := client.CreateMealSitting(ctx, frameID, MealSittingData{
		Summary:        prefix + "-summary",
		RecipeID:       recipe.ID,
		MealCategoryID: mealCategoryID,
		Date:           date,
	})
	if err == nil {
		_ = client.DeleteMealSitting(ctx, frameID, sitting.ID, date)
		t.Fatal("expected a validation error for a summary sent with a recipe")
	}
	if !IsValidation(err) || !strings.Contains(err.Error(), "blank") {
		t.Errorf("expected a 'must be blank' validation error, got: %v", err)
	}
}
