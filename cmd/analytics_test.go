package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

func analyticsTimeRange() (time.Time, time.Time) {
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	return start, end
}

func TestComputeAnalytics_Empty(t *testing.T) {
	start, end := analyticsTimeRange()
	stats := computeAnalytics(nil, nil, nil, nil, nil, start, end)
	if stats.PeriodDays < 30 {
		t.Errorf("PeriodDays: want >=30, got %d", stats.PeriodDays)
	}
	if len(stats.Assignees) != 0 {
		t.Errorf("expected 0 assignees, got %d", len(stats.Assignees))
	}
	if len(stats.TopChores) != 0 {
		t.Errorf("expected 0 top chores, got %d", len(stats.TopChores))
	}
	if stats.Rewards.Total != 0 || stats.Rewards.Redeemed != 0 {
		t.Errorf("expected 0 rewards, got %+v", stats.Rewards)
	}
	if stats.CalendarStats.TotalEvents != 0 {
		t.Errorf("expected 0 events, got %d", stats.CalendarStats.TotalEvents)
	}
}

func TestComputeAnalytics_ChoreCompletion(t *testing.T) {
	chores := []lib.Chore{
		{AssigneeID: "1", Title: "Vacuum", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", Title: "Dishes", Status: lib.ChoreStatusPending},
		{AssigneeID: "2", Title: "Vacuum", Status: lib.ChoreStatusComplete},
		{AssigneeID: "2", Title: "Vacuum", Status: lib.ChoreStatusComplete},
	}
	catNames := map[string]string{"1": "Alice", "2": "Bob"}

	start, end := analyticsTimeRange()
	stats := computeAnalytics(chores, nil, nil, nil, catNames, start, end)

	if len(stats.Assignees) != 2 {
		t.Fatalf("expected 2 assignees, got %d", len(stats.Assignees))
	}

	// Sorted alphabetically: Alice first
	alice := stats.Assignees[0]
	if alice.Name != "Alice" {
		t.Errorf("expected Alice first, got %s", alice.Name)
	}
	if alice.TotalChores != 2 || alice.CompletedChores != 1 {
		t.Errorf("Alice: want 2 total / 1 completed, got %d/%d", alice.TotalChores, alice.CompletedChores)
	}
	if alice.CompletionRate != 50.0 {
		t.Errorf("Alice CompletionRate: want 50.0, got %.1f", alice.CompletionRate)
	}

	bob := stats.Assignees[1]
	if bob.TotalChores != 2 || bob.CompletedChores != 2 {
		t.Errorf("Bob: want 2 total / 2 completed, got %d/%d", bob.TotalChores, bob.CompletedChores)
	}
	if bob.CompletionRate != 100.0 {
		t.Errorf("Bob CompletionRate: want 100.0, got %.1f", bob.CompletionRate)
	}
}

func TestComputeAnalytics_TopChores(t *testing.T) {
	chores := []lib.Chore{
		{Title: "Vacuum", Status: lib.ChoreStatusComplete},
		{Title: "Vacuum", Status: lib.ChoreStatusComplete},
		{Title: "Vacuum", Status: lib.ChoreStatusPending},
		{Title: "Dishes", Status: lib.ChoreStatusComplete},
		{Title: "Dishes", Status: lib.ChoreStatusComplete},
		{Title: "Laundry", Status: lib.ChoreStatusPending},
	}

	start, end := analyticsTimeRange()
	stats := computeAnalytics(chores, nil, nil, nil, nil, start, end)

	if len(stats.TopChores) == 0 {
		t.Fatal("expected top chores, got none")
	}
	if stats.TopChores[0].Title != "Vacuum" {
		t.Errorf("expected Vacuum as top chore, got %s", stats.TopChores[0].Title)
	}
	if stats.TopChores[0].Count != 3 {
		t.Errorf("Vacuum count: want 3, got %d", stats.TopChores[0].Count)
	}
}

func TestComputeAnalytics_TopChoresLimit(t *testing.T) {
	var chores []lib.Chore
	for i := 0; i < 10; i++ {
		chores = append(chores, lib.Chore{Title: "chore", Status: lib.ChoreStatusComplete})
	}
	for _, title := range []string{"A", "B", "C", "D", "E", "F", "G"} {
		chores = append(chores, lib.Chore{Title: title})
	}

	start, end := analyticsTimeRange()
	stats := computeAnalytics(chores, nil, nil, nil, nil, start, end)

	if len(stats.TopChores) > 5 {
		t.Errorf("expected at most 5 top chores, got %d", len(stats.TopChores))
	}
}

func TestComputeAnalytics_RewardStats(t *testing.T) {
	rewards := []lib.Reward{
		{ID: "r1", Redeemed: true},
		{ID: "r2", Redeemed: true},
		{ID: "r3", Redeemed: false},
	}

	start, end := analyticsTimeRange()
	stats := computeAnalytics(nil, rewards, nil, nil, nil, start, end)

	if stats.Rewards.Total != 3 {
		t.Errorf("Total: want 3, got %d", stats.Rewards.Total)
	}
	if stats.Rewards.Redeemed != 2 {
		t.Errorf("Redeemed: want 2, got %d", stats.Rewards.Redeemed)
	}
}

func TestComputeAnalytics_CalendarDensity(t *testing.T) {
	events := []lib.CalendarEvent{{ID: "e1"}, {ID: "e2"}, {ID: "e3"}}

	start, end := analyticsTimeRange()
	stats := computeAnalytics(nil, nil, nil, events, nil, start, end)

	if stats.CalendarStats.TotalEvents != 3 {
		t.Errorf("TotalEvents: want 3, got %d", stats.CalendarStats.TotalEvents)
	}
	expectedRate := 3.0 / float64(stats.PeriodDays)
	if stats.CalendarStats.EventsPerDay != expectedRate {
		t.Errorf("EventsPerDay: want %.4f, got %.4f", expectedRate, stats.CalendarStats.EventsPerDay)
	}
}

func TestComputeAnalytics_AssigneeFallsBackToID(t *testing.T) {
	chores := []lib.Chore{
		{AssigneeID: "99", Title: "Chore", Status: lib.ChoreStatusComplete},
	}

	start, end := analyticsTimeRange()
	stats := computeAnalytics(chores, nil, nil, nil, map[string]string{}, start, end)

	if len(stats.Assignees) != 1 {
		t.Fatalf("expected 1 assignee, got %d", len(stats.Assignees))
	}
	if stats.Assignees[0].Name != "99" {
		t.Errorf("expected fallback to ID, got %s", stats.Assignees[0].Name)
	}
}

func TestPrintAnalyticsText_Output(t *testing.T) {
	stats := AnalyticsStats{
		PeriodDays: 7,
		StartDate:  "2026-04-22",
		EndDate:    "2026-04-29",
		Assignees: []AssigneeStats{
			{Name: "Alice", TotalChores: 5, CompletedChores: 4, CompletionRate: 80.0, PointBalance: 100},
		},
		TopChores: []ChoreFrequency{
			{Title: "Vacuum", Count: 3, Completed: 2},
		},
		Rewards:       RewardStats{Total: 5, Redeemed: 2},
		CalendarStats: CalendarActivityStats{TotalEvents: 7, EventsPerDay: 1.0},
	}

	output := captureStdout(func() { printAnalyticsText(stats) })

	if !strings.Contains(output, "Alice") {
		t.Errorf("expected 'Alice' in output, got: %s", output)
	}
	if !strings.Contains(output, "Vacuum") {
		t.Errorf("expected 'Vacuum' in output, got: %s", output)
	}
	if !strings.Contains(output, "80.0%") {
		t.Errorf("expected completion rate in output, got: %s", output)
	}
	if !strings.Contains(output, "7 events") {
		t.Errorf("expected event count in output, got: %s", output)
	}
}

func TestPrintAnalyticsText_EmptyAssignees(t *testing.T) {
	stats := AnalyticsStats{PeriodDays: 30, StartDate: "2026-04-01", EndDate: "2026-05-01"}

	output := captureStdout(func() { printAnalyticsText(stats) })

	if !strings.Contains(output, "(none)") {
		t.Errorf("expected '(none)' for empty assignees, got: %s", output)
	}
}
