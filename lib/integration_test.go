//go:build integration

package lib

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

const integrationTestPrefix = "integration-test-"

var (
	sharedClient  *Client
	sharedFrameID string
	clientOnce    sync.Once
	clientErr     error
)

// TestMain sweeps leftover integration-test-* resources before running tests
// so that interrupted prior runs don't accumulate artifacts on the live frame.
func TestMain(m *testing.M) {
	sweepIntegrationTestResources()
	os.Exit(m.Run())
}

func sweepIntegrationTestResources() {
	email := os.Getenv("SKYLIGHT_EMAIL")
	password := os.Getenv("SKYLIGHT_PASSWORD")
	frameID := os.Getenv("SKYLIGHT_FRAME_ID")
	loadSkylightConfig(&email, &password, &frameID)
	if email == "" || password == "" || frameID == "" {
		return
	}

	fingerprint := integrationTestPrefix + frameID
	tok, err := LoginHeadless(email, password, fingerprint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sweep: auth failed: %v\n", err)
		return
	}
	c, err := NewClientWithToken("integration", tok.AccessToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sweep: client init failed: %v\n", err)
		return
	}

	ctx := context.Background()
	now := time.Now()

	// Sweep chores (past year + today)
	chores, err := c.ListChores(ctx, frameID, ChoreListOptions{
		After:       now.AddDate(-1, 0, 0).Format(DateFormat),
		Before:      now.AddDate(0, 0, 1).Format(DateFormat),
		IncludeLate: true,
	})
	if err == nil {
		for _, ch := range chores {
			if strings.HasPrefix(ch.Title, integrationTestPrefix) {
				if ch.Recurring {
					_ = c.DeleteRecurringChore(ctx, frameID, ch.ID)
				} else {
					_ = c.DeleteChore(ctx, frameID, ch.ID)
				}
			}
		}
	}

	// Sweep rewards
	rewards, err := c.ListRewards(ctx, frameID)
	if err == nil {
		for _, r := range rewards {
			if strings.HasPrefix(r.Title, integrationTestPrefix) {
				_ = c.DeleteReward(ctx, frameID, r.ID)
			}
		}
	}

	// Sweep lists
	lists, err := c.ListLists(ctx, frameID)
	if err == nil {
		for _, l := range lists {
			if strings.HasPrefix(l.Title, integrationTestPrefix) {
				_ = c.DeleteList(ctx, frameID, l.ID)
			}
		}
	}

	// Sweep recipes
	recipes, err := c.ListRecipes(ctx, frameID)
	if err == nil {
		for _, r := range recipes {
			if strings.HasPrefix(r.Title, integrationTestPrefix) {
				_ = c.DeleteRecipe(ctx, frameID, r.ID)
			}
		}
	}

	// Sweep calendar events (past year through next 30 days)
	events, err := c.ListCalendarEvents(ctx, frameID,
		now.AddDate(-1, 0, 0).Format(DateFormat),
		now.AddDate(0, 0, 30).Format(DateFormat),
		"",
	)
	if err == nil {
		for _, e := range events {
			if strings.HasPrefix(e.Title, integrationTestPrefix) {
				_ = c.DeleteCalendarEvent(ctx, frameID, e.ID)
			}
		}
	}
}

func integrationClient(t *testing.T) (*Client, string) {
	t.Helper()

	email := os.Getenv("SKYLIGHT_EMAIL")
	password := os.Getenv("SKYLIGHT_PASSWORD")
	frameID := os.Getenv("SKYLIGHT_FRAME_ID")

	loadSkylightConfig(&email, &password, &frameID)

	if frameID == "" {
		t.Skip("skipping: set SKYLIGHT_FRAME_ID or add it to ~/.skylight/config")
	}
	if email == "" || password == "" {
		t.Skip("skipping: set SKYLIGHT_EMAIL and SKYLIGHT_PASSWORD or add them to ~/.skylight/config")
	}

	clientOnce.Do(func() {
		fingerprint := integrationTestPrefix + frameID
		tok, err := LoginHeadless(email, password, fingerprint)
		if err != nil {
			clientErr = fmt.Errorf("LoginHeadless: %w", err)
			return
		}

		sharedClient, clientErr = NewClientWithToken("integration", tok.AccessToken,
			WithRateLimit(rate.Limit(2), 5),
			WithRetry(3, 500*time.Millisecond, 10*time.Second),
		)
		sharedFrameID = frameID
	})

	if clientErr != nil {
		t.Fatalf("integration auth: %v", clientErr)
	}

	return sharedClient, sharedFrameID
}

func TestIntegration_GetFrame(t *testing.T) {
	client, frameID := integrationClient(t)

	frame, err := client.GetFrame(context.Background(), frameID)
	if err != nil {
		t.Fatalf("GetFrame: %v", err)
	}

	if frame.ID != frameID {
		t.Errorf("expected frame ID %s, got %s", frameID, frame.ID)
	}

	if frame.Name == "" {
		t.Error("expected non-empty frame name")
	}

	t.Logf("frame: %s (%s)", frame.Name, frame.ID)
}

func TestIntegration_ListDevices(t *testing.T) {
	client, frameID := integrationClient(t)

	devices, err := client.ListDevices(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}

	t.Logf("devices: %d", len(devices))

	for _, d := range devices {
		if d.ID == "" {
			t.Error("device has empty ID")
		}
	}
}

func TestIntegration_ListCategories(t *testing.T) {
	client, frameID := integrationClient(t)

	categories, err := client.ListCategories(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	if len(categories) == 0 {
		t.Fatal("expected at least one category (family member)")
	}

	t.Logf("categories: %d", len(categories))

	for _, c := range categories {
		if c.ID == "" {
			t.Error("category has empty ID")
		}
		if c.Name == "" {
			t.Error("category has empty Name")
		}
	}
}

func TestIntegration_ListRewards(t *testing.T) {
	client, frameID := integrationClient(t)

	rewards, err := client.ListRewards(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListRewards: %v", err)
	}

	t.Logf("rewards: %d", len(rewards))

	for _, r := range rewards {
		if r.ID == "" {
			t.Error("reward has empty ID")
		}
		if r.Title == "" {
			t.Error("reward has empty Title")
		}
	}
}

func TestIntegration_ListChores(t *testing.T) {
	client, frameID := integrationClient(t)

	now := time.Now()
	opts := ChoreListOptions{
		After:  now.Format(DateFormat),
		Before: now.AddDate(0, 0, 30).Format(DateFormat),
	}

	chores, err := client.ListChores(context.Background(), frameID, opts)
	if err != nil {
		t.Fatalf("ListChores: %v", err)
	}

	t.Logf("chores: %d", len(chores))

	for _, c := range chores {
		if c.ID == "" {
			t.Error("chore has empty ID")
		}
		if c.Title == "" {
			t.Error("chore has empty Title")
		}
	}
}

func TestIntegration_ListCalendarEvents(t *testing.T) {
	client, frameID := integrationClient(t)

	now := time.Now()
	start := now.Format(DateFormat)
	end := now.AddDate(0, 0, 7).Format(DateFormat)

	events, err := client.ListCalendarEvents(context.Background(), frameID, start, end, "")
	if err != nil {
		if strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "Internal Server Error") {
			t.Skipf("ListCalendarEvents: skipping due to API instability: %v", err)
		}
		t.Fatalf("ListCalendarEvents: %v", err)
	}

	t.Logf("calendar events: %d", len(events))

	for _, e := range events {
		if e.ID == "" {
			t.Error("event has empty ID")
		}
		if e.Title == "" {
			t.Error("event has empty Title")
		}
	}
}

func TestIntegration_ListLists(t *testing.T) {
	client, frameID := integrationClient(t)

	lists, err := client.ListLists(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListLists: %v", err)
	}

	t.Logf("lists: %d", len(lists))

	for _, l := range lists {
		if l.ID == "" {
			t.Error("list has empty ID")
		}
		if l.Title == "" {
			t.Error("list has empty Title")
		}
	}
}

func TestIntegration_ListRecipes(t *testing.T) {
	client, frameID := integrationClient(t)

	recipes, err := client.ListRecipes(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListRecipes: %v", err)
	}

	t.Logf("recipes: %d", len(recipes))

	for _, r := range recipes {
		if r.ID == "" {
			t.Error("recipe has empty ID")
		}
		if r.Title == "" {
			t.Error("recipe has empty Title")
		}
	}
}

func TestIntegration_ListMealCategories(t *testing.T) {
	client, frameID := integrationClient(t)

	categories, err := client.ListMealCategories(context.Background(), frameID)
	if err != nil {
		t.Fatalf("ListMealCategories: %v", err)
	}

	t.Logf("meal categories: %d", len(categories))

	for _, c := range categories {
		if c.ID == "" {
			t.Error("meal category has empty ID")
		}
	}
}
