package lib

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestToCalendarEventNilCategory(t *testing.T) {
	entry := calendarAPIEntry{ID: "e1"}
	entry.Attributes.Summary = "Test"
	ev := entry.toCalendarEvent()
	if ev.CategoryID != "" {
		t.Errorf("CategoryID should be empty, got %q", ev.CategoryID)
	}
	if ev.Title != "Test" {
		t.Errorf("Title = %q, want %q", ev.Title, "Test")
	}
}

func TestToCalendarEventWithCategory(t *testing.T) {
	entry := calendarAPIEntry{ID: "e1"}
	entry.Attributes.Summary = "Event"
	entry.Relationships.Categories.Data = []apiRelationshipData{{ID: "cat1"}}
	ev := entry.toCalendarEvent()
	if ev.CategoryID != "cat1" {
		t.Errorf("CategoryID = %q, want %q", ev.CategoryID, "cat1")
	}
}

func TestToRewardNilCategory(t *testing.T) {
	entry := rewardAPIEntry{ID: "r1"}
	entry.Attributes.Name = "Prize"
	entry.Attributes.PointValue = 5
	r := entry.toReward()
	if r.CategoryID != "" {
		t.Errorf("CategoryID should be empty, got %q", r.CategoryID)
	}
	if r.Redeemed {
		t.Error("Redeemed should be false when RedeemedAt is nil")
	}
}

func TestToRewardWithCategory(t *testing.T) {
	entry := rewardAPIEntry{ID: "r1"}
	entry.Relationships.Category.Data = &apiRelationshipData{ID: "cat2"}
	redeemed := "2024-01-01T00:00:00Z"
	entry.Attributes.RedeemedAt = &redeemed
	r := entry.toReward()
	if r.CategoryID != "cat2" {
		t.Errorf("CategoryID = %q, want %q", r.CategoryID, "cat2")
	}
	if !r.Redeemed {
		t.Error("Redeemed should be true when RedeemedAt is set")
	}
}

func TestToRecipeNilMealCategory(t *testing.T) {
	entry := recipeAPIEntry{ID: "rec1"}
	entry.Attributes.Summary = "Pasta"
	entry.Attributes.URL = "http://example.com"
	r := entry.toRecipe()
	if r.MealCategoryID != "" {
		t.Errorf("MealCategoryID should be empty, got %q", r.MealCategoryID)
	}
	if r.Title != "Pasta" {
		t.Errorf("Title = %q, want %q", r.Title, "Pasta")
	}
}

func TestToRecipeWithMealCategory(t *testing.T) {
	entry := recipeAPIEntry{ID: "rec1"}
	entry.Relationships.MealCategory.Data = &apiRelationshipData{ID: "mc1"}
	r := entry.toRecipe()
	if r.MealCategoryID != "mc1" {
		t.Errorf("MealCategoryID = %q, want %q", r.MealCategoryID, "mc1")
	}
}

func TestToCategoryNilProfilePic(t *testing.T) {
	entry := categoryAPIEntry{ID: "cat1"}
	entry.Attributes.Label = "Alice"
	entry.Attributes.Color = "blue"
	c := entry.toCategory()
	if c.AvatarURL != "" {
		t.Errorf("AvatarURL should be empty, got %q", c.AvatarURL)
	}
}

func TestToCategoryWithProfilePic(t *testing.T) {
	entry := categoryAPIEntry{ID: "cat1"}
	url := "http://example.com/pic.jpg"
	entry.Attributes.ProfilePictureURLs.Large = &url
	c := entry.toCategory()
	if c.AvatarURL != url {
		t.Errorf("AvatarURL = %q, want %q", c.AvatarURL, url)
	}
}

func TestToCategoryLinkedToProfile(t *testing.T) {
	entry := categoryAPIEntry{ID: "cat1"}
	entry.Attributes.Label = "Mom"
	entry.Attributes.LinkedToProfile = true
	entry.Attributes.SelectedForChoreChart = true
	c := entry.toCategory()
	if !c.LinkedToProfile {
		t.Errorf("LinkedToProfile should be true")
	}
	if !c.SelectedForChoreChart {
		t.Errorf("SelectedForChoreChart should be true")
	}
}

func TestToCategoryWithProfilePicOriginalFallback(t *testing.T) {
	entry := categoryAPIEntry{ID: "cat1"}
	url := "http://example.com/original.jpg"
	entry.Attributes.ProfilePictureURLs.Original = &url
	c := entry.toCategory()
	if c.AvatarURL != url {
		t.Errorf("AvatarURL = %q, want %q", c.AvatarURL, url)
	}
}

func TestToCategoryWithXLProfilePic(t *testing.T) {
	entry := categoryAPIEntry{ID: "cat1"}
	xl := "http://example.com/xl.jpg"
	large := "http://example.com/large.jpg"
	entry.Attributes.ProfilePictureURLs.XL = &xl
	entry.Attributes.ProfilePictureURLs.Large = &large
	c := entry.toCategory()
	if c.AvatarURL != xl {
		t.Errorf("AvatarURL = %q, want XL %q", c.AvatarURL, xl)
	}
}

func TestToCategoryWithAvatarRelationship(t *testing.T) {
	entry := categoryAPIEntry{ID: "cat1"}
	entry.Relationships.Avatar.Data = &apiRelationshipData{ID: "av42"}
	c := entry.toCategory()
	if c.AvatarID != "av42" {
		t.Errorf("AvatarID = %q, want %q", c.AvatarID, "av42")
	}
	if c.HatID != "" {
		t.Errorf("HatID should be empty when hat relationship is nil, got %q", c.HatID)
	}
}

func TestToCategoryWithHatRelationship(t *testing.T) {
	entry := categoryAPIEntry{ID: "cat1"}
	entry.Relationships.Hat.Data = &apiRelationshipData{ID: "hat7"}
	c := entry.toCategory()
	if c.HatID != "hat7" {
		t.Errorf("HatID = %q, want %q", c.HatID, "hat7")
	}
	if c.AvatarID != "" {
		t.Errorf("AvatarID should be empty when avatar relationship is nil, got %q", c.AvatarID)
	}
}

func TestToCategoryWithBothRelationships(t *testing.T) {
	entry := categoryAPIEntry{ID: "cat1"}
	entry.Relationships.Avatar.Data = &apiRelationshipData{ID: "av1"}
	entry.Relationships.Hat.Data = &apiRelationshipData{ID: "hat1"}
	c := entry.toCategory()
	if c.AvatarID != "av1" {
		t.Errorf("AvatarID = %q, want %q", c.AvatarID, "av1")
	}
	if c.HatID != "hat1" {
		t.Errorf("HatID = %q, want %q", c.HatID, "hat1")
	}
}

func TestToMealSittingNilRelationships(t *testing.T) {
	entry := mealSittingAPIEntry{ID: "ms1"}
	entry.Attributes.Summary = "Dinner"
	s := entry.toMealSitting()
	if s.MealCategoryID != "" {
		t.Errorf("MealCategoryID should be empty, got %q", s.MealCategoryID)
	}
	if s.RecipeID != "" {
		t.Errorf("RecipeID should be empty, got %q", s.RecipeID)
	}
}

func TestToMealSittingWithRelationships(t *testing.T) {
	entry := mealSittingAPIEntry{ID: "ms1"}
	entry.Relationships.MealCategory.Data = &apiRelationshipData{ID: "mc1"}
	entry.Relationships.MealRecipe.Data = &apiRelationshipData{ID: "rec1"}
	s := entry.toMealSitting()
	if s.MealCategoryID != "mc1" {
		t.Errorf("MealCategoryID = %q, want %q", s.MealCategoryID, "mc1")
	}
	if s.RecipeID != "rec1" {
		t.Errorf("RecipeID = %q, want %q", s.RecipeID, "rec1")
	}
}

func TestToChoreNilCategory(t *testing.T) {
	entry := choreAPIEntry{ID: "ch1"}
	entry.Attributes.Summary = "Clean"
	c := entry.toChore()
	if c.AssigneeID != "" {
		t.Errorf("AssigneeID should be empty, got %q", c.AssigneeID)
	}
}

func TestToChoreWithCategory(t *testing.T) {
	entry := choreAPIEntry{ID: "ch1"}
	entry.Relationships.Category.Data = &apiRelationshipData{ID: "cat1"}
	c := entry.toChore()
	if c.AssigneeID != "cat1" {
		t.Errorf("AssigneeID = %q, want %q", c.AssigneeID, "cat1")
	}
}

func TestToListItem(t *testing.T) {
	entry := listItemAPIEntry{ID: "li1"}
	entry.Attributes.Label = "Milk"
	entry.Attributes.Status = "completed"
	entry.Attributes.Position = 3
	item := entry.toListItem()
	if !item.Completed {
		t.Error("Completed should be true for status=completed")
	}
	if item.Position != 3 {
		t.Errorf("Position = %d, want 3", item.Position)
	}
}

func TestToListItemNotCompleted(t *testing.T) {
	entry := listItemAPIEntry{ID: "li1"}
	entry.Attributes.Status = "active"
	item := entry.toListItem()
	if item.Completed {
		t.Error("Completed should be false for status=active")
	}
}

func TestToSourceCalendar(t *testing.T) {
	entry := sourceCalendarAPIEntry{ID: "sc1"}
	entry.Attributes.Label = "Work"
	entry.Attributes.Kind = "google"
	sc := entry.toSourceCalendar()
	if sc.Name != "Work" {
		t.Errorf("Name = %q, want %q", sc.Name, "Work")
	}
	if sc.Provider != "google" {
		t.Errorf("Provider = %q, want %q", sc.Provider, "google")
	}
}

func TestToList(t *testing.T) {
	entry := listAPIEntry{ID: "l1"}
	entry.Attributes.Label = "Groceries"
	entry.Attributes.Color = "green"
	entry.Attributes.Kind = "shopping"
	l := entry.toList()
	if l.Title != "Groceries" {
		t.Errorf("Title = %q, want %q", l.Title, "Groceries")
	}
	if l.Kind != "shopping" {
		t.Errorf("Kind = %q, want %q", l.Kind, "shopping")
	}
}

func TestToMealCategory(t *testing.T) {
	entry := mealCategoryAPIEntry{ID: "mc1"}
	entry.Attributes.Label = "Breakfast"
	entry.Attributes.Color = "yellow"
	mc := entry.toMealCategory()
	if mc.Name != "Breakfast" {
		t.Errorf("Name = %q, want %q", mc.Name, "Breakfast")
	}
}

func TestToFrame(t *testing.T) {
	entry := frameAPIEntry{ID: "f1"}
	entry.Attributes.Name = "Living Room"
	entry.Attributes.TimeZone = "America/Chicago"
	f := entry.toFrame()
	if f.Name != "Living Room" {
		t.Errorf("Name = %q, want %q", f.Name, "Living Room")
	}
	if f.TimeZone != "America/Chicago" {
		t.Errorf("TimeZone = %q, want %q", f.TimeZone, "America/Chicago")
	}
}

const (
	testDeviceTimezone   = "America/Chicago"
	testDeviceBrightness = 200
	testDeviceSleepsAt   = "00:00"
	testDeviceWakesAt    = "06:00"
)

func TestToDevice(t *testing.T) {
	entry := deviceAPIEntry{ID: "d1"}
	entry.Attributes.Name = "Frame"
	entry.Attributes.Activated = true
	entry.Attributes.Timezone = testDeviceTimezone
	entry.Attributes.Brightness = testDeviceBrightness
	entry.Attributes.SleepsAt = testDeviceSleepsAt
	entry.Attributes.WakesAt = testDeviceWakesAt
	entry.Attributes.Nightlight = true
	d := entry.toDevice()
	if !d.Online {
		t.Error("Online should be true when Activated is true")
	}
	if d.Timezone != entry.Attributes.Timezone {
		t.Errorf("Timezone = %q, want %q", d.Timezone, entry.Attributes.Timezone)
	}
	if d.Brightness != entry.Attributes.Brightness {
		t.Errorf("Brightness = %d, want %d", d.Brightness, entry.Attributes.Brightness)
	}
	if d.SleepsAt != entry.Attributes.SleepsAt {
		t.Errorf("SleepsAt = %q, want %q", d.SleepsAt, entry.Attributes.SleepsAt)
	}
	if d.WakesAt != entry.Attributes.WakesAt {
		t.Errorf("WakesAt = %q, want %q", d.WakesAt, entry.Attributes.WakesAt)
	}
	if !d.Nightlight {
		t.Error("Nightlight should be true")
	}
}

func TestToAvatar(t *testing.T) {
	entry := avatarAPIEntry{ID: "a1"}
	entry.Attributes.Name = "Cat"
	entry.Attributes.ImageURL = "http://example.com/cat.png"
	a := entry.toAvatar()
	if a.Name != "Cat" {
		t.Errorf("Name = %q, want %q", a.Name, "Cat")
	}
	if a.ImageURL != "http://example.com/cat.png" {
		t.Errorf("ImageURL = %q, want %q", a.ImageURL, "http://example.com/cat.png")
	}
}

func TestToAvatarWithKindAndDisplayPosition(t *testing.T) {
	pos := 3
	entry := avatarAPIEntry{ID: "a2"}
	entry.Attributes.Name = "Raccoon"
	entry.Attributes.ImageURL = "http://example.com/raccoon.png"
	entry.Attributes.Kind = "default"
	entry.Attributes.DisplayPosition = &pos
	a := entry.toAvatar()
	if a.Kind != "default" {
		t.Errorf("Kind = %q, want %q", a.Kind, "default")
	}
	if a.DisplayPosition == nil || *a.DisplayPosition != 3 {
		t.Errorf("DisplayPosition = %v, want 3", a.DisplayPosition)
	}
}

func TestToAvatarNilDisplayPosition(t *testing.T) {
	entry := avatarAPIEntry{ID: "a3"}
	entry.Attributes.Kind = "default"
	a := entry.toAvatar()
	if a.DisplayPosition != nil {
		t.Errorf("DisplayPosition should be nil, got %v", a.DisplayPosition)
	}
}

func TestListItemCompletedFalseJSON(t *testing.T) {
	b, err := json.Marshal(ListItem{ID: "i1", Completed: false})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"completed":false`) {
		t.Errorf("expected completed:false in JSON, got: %s", b)
	}
}

func TestRewardRedeemedFalseJSON(t *testing.T) {
	b, err := json.Marshal(Reward{ID: "r1", Redeemed: false})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"redeemed":false`) {
		t.Errorf("expected redeemed:false in JSON, got: %s", b)
	}
}
