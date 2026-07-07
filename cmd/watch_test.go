package cmd

import (
	"strings"
	"testing"
)

func TestParseWatchResources_All(t *testing.T) {
	cases := []string{"", "all"}
	for _, c := range cases {
		got := parseWatchResources(c)
		if len(got) != len(allWatchResources) {
			t.Errorf("parseWatchResources(%q): expected %d resources, got %d: %v", c, len(allWatchResources), len(got), got)
		}
	}
}

func TestParseWatchResources_Specific(t *testing.T) {
	got := parseWatchResources("rewards,chores")
	if len(got) != 2 {
		t.Fatalf("expected 2 resources, got %d: %v", len(got), got)
	}
	if got[0] != "rewards" || got[1] != "chores" {
		t.Errorf("unexpected resources: %v", got)
	}
}

func TestParseWatchResources_Single(t *testing.T) {
	got := parseWatchResources("calendar")
	if len(got) != 1 || got[0] != "calendar" {
		t.Errorf("expected [calendar], got %v", got)
	}
}

func TestParseWatchResources_TrimsSpaces(t *testing.T) {
	got := parseWatchResources(" rewards , chores ")
	if len(got) != 2 {
		t.Fatalf("expected 2 resources, got %d: %v", len(got), got)
	}
	if got[0] != "rewards" || got[1] != "chores" {
		t.Errorf("expected trimmed resources, got %v", got)
	}
}

func TestWatchState_TracksSeenRewards(t *testing.T) {
	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	if _, seen := state.seenRewardIDs["r1"]; seen {
		t.Error("expected r1 not yet seen")
	}
	state.seenRewardIDs["r1"] = struct{}{}
	if _, seen := state.seenRewardIDs["r1"]; !seen {
		t.Error("expected r1 to be seen after marking")
	}
}

func TestWatchState_TracksSeenChores(t *testing.T) {
	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	if _, seen := state.seenChoreIDs["c1"]; seen {
		t.Error("expected c1 not yet seen")
	}
	state.seenChoreIDs["c1"] = struct{}{}
	if _, seen := state.seenChoreIDs["c1"]; !seen {
		t.Error("expected c1 to be seen after marking")
	}
}

func TestWatchState_TracksSeenEvents(t *testing.T) {
	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	if _, seen := state.seenEventIDs["e1"]; seen {
		t.Error("expected e1 not yet seen")
	}
	state.seenEventIDs["e1"] = struct{}{}
	if _, seen := state.seenEventIDs["e1"]; !seen {
		t.Error("expected e1 to be seen after marking")
	}
}

func TestParseWatchResources_IgnoresInvalid(t *testing.T) {
	got := parseWatchResources("rewards,bogus,chores")
	if len(got) != 2 {
		t.Fatalf("expected 2 valid resources, got %d: %v", len(got), got)
	}
	if got[0] != "rewards" || got[1] != "chores" {
		t.Errorf("expected [rewards chores], got %v", got)
	}
}

func TestWatchCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "watch" {
			found = true
			break
		}
	}
	if !found {
		t.Error("watch command not registered on root")
	}
}

func TestWatchCmdHasFlags(t *testing.T) {
	if f := watchCmd.Flags().Lookup("interval"); f == nil {
		t.Error("expected --interval flag on watch command")
	}
	if f := watchCmd.Flags().Lookup("resources"); f == nil {
		t.Error("expected --resources flag on watch command")
	}
}

func TestWatchCmdLong_ContainsResources(t *testing.T) {
	long := watchCmd.Long
	for _, r := range []string{"rewards", "chores", "calendar", "lists", "routines"} {
		if !strings.Contains(long, r) {
			t.Errorf("expected %q mentioned in watch command long description", r)
		}
	}
}
