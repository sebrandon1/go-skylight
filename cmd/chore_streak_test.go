package cmd

import (
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestComputeChoreStreaks_Empty(t *testing.T) {
	results := computeChoreStreaks(nil, []string{"2026-04-01"}, nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestComputeChoreStreaks_AllCompleted(t *testing.T) {
	dates := []string{"2026-04-27", "2026-04-28", "2026-04-29"}
	chores := []lib.Chore{
		{AssigneeID: "1", DueDate: "2026-04-27", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", DueDate: "2026-04-28", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", DueDate: "2026-04-29", Status: lib.ChoreStatusComplete},
	}
	catNames := map[string]string{"1": "Alice"}

	results := computeChoreStreaks(chores, dates, catNames)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	s := results[0]
	if s.CurrentStreak != 3 {
		t.Errorf("CurrentStreak: want 3, got %d", s.CurrentStreak)
	}
	if s.LongestStreak != 3 {
		t.Errorf("LongestStreak: want 3, got %d", s.LongestStreak)
	}
	if s.CompletionRate != 100.0 {
		t.Errorf("CompletionRate: want 100.0, got %.2f", s.CompletionRate)
	}
	if s.AssigneeName != "Alice" {
		t.Errorf("AssigneeName: want Alice, got %s", s.AssigneeName)
	}
}

func TestComputeChoreStreaks_BrokenStreak(t *testing.T) {
	dates := []string{"2026-04-27", "2026-04-28", "2026-04-29"}
	chores := []lib.Chore{
		{AssigneeID: "1", DueDate: "2026-04-27", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", DueDate: "2026-04-28", Status: lib.ChoreStatusPending},
		{AssigneeID: "1", DueDate: "2026-04-29", Status: lib.ChoreStatusComplete},
	}

	results := computeChoreStreaks(chores, dates, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	s := results[0]
	if s.CurrentStreak != 1 {
		t.Errorf("CurrentStreak: want 1, got %d", s.CurrentStreak)
	}
	if s.LongestStreak != 1 {
		t.Errorf("LongestStreak: want 1, got %d", s.LongestStreak)
	}
}

func TestComputeChoreStreaks_SkipDaysWithNoChores(t *testing.T) {
	// April 28 has no chores for assignee "1" — streak should not break.
	dates := []string{"2026-04-27", "2026-04-28", "2026-04-29"}
	chores := []lib.Chore{
		{AssigneeID: "1", DueDate: "2026-04-27", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", DueDate: "2026-04-29", Status: lib.ChoreStatusComplete},
	}

	results := computeChoreStreaks(chores, dates, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	s := results[0]
	// Both days with chores are complete, so current streak = 2, longest = 2
	if s.CurrentStreak != 2 {
		t.Errorf("CurrentStreak: want 2, got %d", s.CurrentStreak)
	}
	if s.LongestStreak != 2 {
		t.Errorf("LongestStreak: want 2, got %d", s.LongestStreak)
	}
}

func TestComputeChoreStreaks_MultipleAssignees(t *testing.T) {
	dates := []string{"2026-04-27", "2026-04-28", "2026-04-29"}
	chores := []lib.Chore{
		{AssigneeID: "1", DueDate: "2026-04-27", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", DueDate: "2026-04-28", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", DueDate: "2026-04-29", Status: lib.ChoreStatusComplete},
		{AssigneeID: "2", DueDate: "2026-04-27", Status: lib.ChoreStatusComplete},
		{AssigneeID: "2", DueDate: "2026-04-28", Status: lib.ChoreStatusPending},
		{AssigneeID: "2", DueDate: "2026-04-29", Status: lib.ChoreStatusPending},
	}
	catNames := map[string]string{"1": "Alice", "2": "Bob"}

	results := computeChoreStreaks(chores, dates, catNames)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Results are sorted by name: Alice first, Bob second.
	alice, bob := results[0], results[1]
	if alice.AssigneeName != "Alice" {
		t.Errorf("expected Alice first, got %s", alice.AssigneeName)
	}
	if alice.CurrentStreak != 3 || alice.LongestStreak != 3 {
		t.Errorf("Alice streak: want 3/3, got %d/%d", alice.CurrentStreak, alice.LongestStreak)
	}
	if bob.CurrentStreak != 0 {
		t.Errorf("Bob CurrentStreak: want 0, got %d", bob.CurrentStreak)
	}
	if bob.LongestStreak != 1 {
		t.Errorf("Bob LongestStreak: want 1, got %d", bob.LongestStreak)
	}
}

func TestComputeChoreStreaks_CompletionRate(t *testing.T) {
	dates := []string{"2026-04-27", "2026-04-28"}
	chores := []lib.Chore{
		{AssigneeID: "1", DueDate: "2026-04-27", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", DueDate: "2026-04-27", Status: lib.ChoreStatusComplete},
		{AssigneeID: "1", DueDate: "2026-04-28", Status: lib.ChoreStatusPending},
		{AssigneeID: "1", DueDate: "2026-04-28", Status: lib.ChoreStatusPending},
	}

	results := computeChoreStreaks(chores, dates, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	s := results[0]
	if s.TotalChores != 4 {
		t.Errorf("TotalChores: want 4, got %d", s.TotalChores)
	}
	if s.CompletedChores != 2 {
		t.Errorf("CompletedChores: want 2, got %d", s.CompletedChores)
	}
	if s.CompletionRate != 50.0 {
		t.Errorf("CompletionRate: want 50.0, got %.2f", s.CompletionRate)
	}
}

func TestComputeChoreStreaks_SkipsUnassigned(t *testing.T) {
	dates := []string{"2026-04-29"}
	chores := []lib.Chore{
		{AssigneeID: "", DueDate: "2026-04-29", Status: lib.ChoreStatusComplete},
	}

	results := computeChoreStreaks(chores, dates, nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results for unassigned chores, got %d", len(results))
	}
}

func TestComputeChoreStreaks_FallbackToID(t *testing.T) {
	dates := []string{"2026-04-29"}
	chores := []lib.Chore{
		{AssigneeID: "99", DueDate: "2026-04-29", Status: lib.ChoreStatusComplete},
	}

	results := computeChoreStreaks(chores, dates, map[string]string{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].AssigneeName != "99" {
		t.Errorf("AssigneeName should fall back to ID, got %s", results[0].AssigneeName)
	}
}
