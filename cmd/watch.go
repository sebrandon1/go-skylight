package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

const (
	watchResourceRewards  = "rewards"
	watchResourceChores   = "chores"
	watchResourceCalendar = "calendar"
	watchResourceLists    = "lists"
	watchResourceRoutines = "routines"
)

var (
	watchInterval  int
	watchResources string
	watchPersist   bool

	allWatchResources = []string{watchResourceRewards, watchResourceChores, watchResourceCalendar, watchResourceLists, watchResourceRoutines}

	validWatchResources = func() map[string]struct{} {
		m := make(map[string]struct{}, len(allWatchResources))
		for _, r := range allWatchResources {
			m[r] = struct{}{}
		}
		return m
	}()
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Poll for changes and print events as they happen",
	Long: `Poll Skylight resources at a regular interval and print new events.

Tracks previously-seen IDs in memory and emits only newly-observed changes:
  - rewards:  newly redeemed rewards
  - chores:   chores newly marked complete
  - calendar: events starting within the next hour
  - lists:    newly created lists
  - routines: newly created routines

Use --persist to persist reward deduplication state to disk
(~/.skylight/poller-state.json) so restarts do not re-emit already-seen
reward redemptions. Has no effect unless rewards is in --resources.

Press Ctrl+C to stop.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if watchInterval < 1 {
			return fmt.Errorf("--interval must be at least 1")
		}

		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		resources, err := parseWatchResources(watchResources)
		if err != nil {
			return err
		}

		ctx := cmd.Context()

		ticker := time.NewTicker(time.Duration(watchInterval) * time.Second)
		defer ticker.Stop()

		frame, err := getFrameOrFail(ctx, client, frameID)
		if err != nil {
			return err
		}

		state := &watchState{
			seenRewardIDs:  make(map[string]struct{}),
			seenChoreIDs:   make(map[string]struct{}),
			seenEventIDs:   make(map[string]struct{}),
			seenListIDs:    make(map[string]struct{}),
			seenRoutineIDs: make(map[string]struct{}),
			timezone:       frame.TimeZone,
		}

		pollResources := resources

		var poller *lib.RewardsPoller
		if watchPersist {
			if slices.Contains(resources, watchResourceRewards) {
				poller = lib.NewRewardsPoller(client, frameID, time.Duration(watchInterval)*time.Second, "")
				poller.Start(ctx)
				defer poller.Stop()
				pollResources = slices.DeleteFunc(slices.Clone(resources), func(r string) bool { return r == watchResourceRewards })
			} else {
				fmt.Fprintln(os.Stderr, "Warning: --persist has no effect without rewards in --resources")
			}
		}

		// Seed in-memory state silently so existing items aren't reported as new.
		state.seeding = true
		poll(ctx, client, state, pollResources)
		state.seeding = false

		// nil channel is never selected in a select statement — used when poller is inactive.
		var pollerEvents <-chan lib.RedemptionEvent
		persistLabel := ""
		if poller != nil {
			pollerEvents = poller.Events()
			persistLabel = ", rewards persisted"
		}

		fmt.Printf("Watching %s (interval: %ds%s). Press Ctrl+C to stop.\n\n",
			strings.Join(resources, ", "), watchInterval, persistLabel)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nStopped.")
				return nil
			case <-ticker.C:
				poll(ctx, client, state, pollResources)
			case e := <-pollerEvents:
				printRedemptionEvent(e)
			}
		}
	},
}

type watchState struct {
	seenRewardIDs  map[string]struct{}
	seenChoreIDs   map[string]struct{}
	seenEventIDs   map[string]struct{}
	seenListIDs    map[string]struct{}
	seenRoutineIDs map[string]struct{}
	seeding        bool
	timezone       string
}

func printRedemptionEvent(e lib.RedemptionEvent) {
	ts := e.ObservedAt.Format("15:04:05")
	if outputFormat == outputJSON {
		printJSON(map[string]any{
			"type": "reward_redeemed", "id": e.RewardID, "title": e.RewardName,
			"points": e.Points, "category_id": e.CategoryID, "child_name": e.ChildName, "ts": ts,
		})
	} else {
		name := e.ChildName
		if name == "" {
			name = e.CategoryID
		}
		fmt.Printf("[%s] REWARD REDEEMED  %s (%d pts) — %s\n", ts, e.RewardName, e.Points, name)
	}
}

func parseWatchResources(s string) ([]string, error) {
	if s == "" || s == resourceAll {
		return allWatchResources, nil
	}
	var out []string
	for _, r := range strings.Split(s, ",") {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if _, ok := validWatchResources[r]; !ok {
			fmt.Fprintf(os.Stderr, "Warning: unknown resource %q (valid: %s)\n", r, strings.Join(allWatchResources, ", "))
			continue
		}
		out = append(out, r)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid resources specified")
	}
	return out, nil
}

func poll(ctx context.Context, client *lib.Client, state *watchState, resources []string) {
	now := time.Now()
	ts := now.Format("15:04:05")

	var wg sync.WaitGroup
	for _, r := range resources {
		switch r {
		case watchResourceRewards:
			wg.Add(1)
			go func() {
				defer wg.Done()
				pollRewards(ctx, client, state, ts)
			}()
		case watchResourceChores:
			wg.Add(1)
			go func() {
				defer wg.Done()
				pollChores(ctx, client, state, now, ts)
			}()
		case watchResourceCalendar:
			wg.Add(1)
			go func() {
				defer wg.Done()
				pollCalendar(ctx, client, state, now, ts)
			}()
		case watchResourceLists:
			wg.Add(1)
			go func() {
				defer wg.Done()
				pollLists(ctx, client, state, ts)
			}()
		case watchResourceRoutines:
			wg.Add(1)
			go func() {
				defer wg.Done()
				pollRoutines(ctx, client, state, ts)
			}()
		}
	}
	wg.Wait()
}

func pollRewards(ctx context.Context, client *lib.Client, state *watchState, ts string) {
	rewards, err := client.ListRewards(ctx, frameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing rewards: %v\n", ts, err)
		return
	}
	// Replace seen set with currently redeemed IDs so memory stays bounded.
	current := make(map[string]struct{})
	for _, r := range rewards {
		if !r.Redeemed {
			continue
		}
		current[r.ID] = struct{}{}
		if _, seen := state.seenRewardIDs[r.ID]; seen {
			continue
		}
		if state.seeding {
			continue
		}
		if outputFormat == outputJSON {
			printJSON(map[string]any{
				"type": "reward_redeemed", "id": r.ID, "title": r.Title,
				"points": r.Points, "category_id": r.CategoryID, "ts": ts,
			})
		} else {
			fmt.Printf("[%s] REWARD REDEEMED  %s (%d pts) — category %s\n",
				ts, r.Title, r.Points, r.CategoryID)
		}
	}
	state.seenRewardIDs = current
}

func pollLists(ctx context.Context, client *lib.Client, state *watchState, ts string) {
	lists, err := client.ListLists(ctx, frameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing lists: %v\n", ts, err)
		return
	}
	current := make(map[string]struct{})
	for _, l := range lists {
		current[l.ID] = struct{}{}
		if _, seen := state.seenListIDs[l.ID]; seen {
			continue
		}
		if state.seeding {
			continue
		}
		if outputFormat == outputJSON {
			printJSON(map[string]any{
				"type": "list_created", "id": l.ID, "title": l.Title, "ts": ts,
			})
		} else {
			fmt.Printf("[%s] LIST CREATED     %s\n", ts, l.Title)
		}
	}
	state.seenListIDs = current
}

func pollRoutines(ctx context.Context, client *lib.Client, state *watchState, ts string) {
	routines, err := client.ListRoutines(ctx, frameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing routines: %v\n", ts, err)
		return
	}
	current := make(map[string]struct{})
	for _, r := range routines {
		current[r.ID] = struct{}{}
		if _, seen := state.seenRoutineIDs[r.ID]; seen {
			continue
		}
		if state.seeding {
			continue
		}
		if outputFormat == outputJSON {
			printJSON(map[string]any{
				"type": "routine_created", "id": r.ID, "title": r.Title, "ts": ts,
			})
		} else {
			fmt.Printf("[%s] ROUTINE CREATED  %s\n", ts, r.Title)
		}
	}
	state.seenRoutineIDs = current
}

func pollChores(ctx context.Context, client *lib.Client, state *watchState, now time.Time, ts string) {
	today := now.Format(lib.DateFormat)
	chores, err := client.ListChores(ctx, frameID, lib.ChoreListOptions{After: today, Before: today, Status: lib.ChoreStatusComplete})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing chores: %v\n", ts, err)
		return
	}
	current := make(map[string]struct{})
	for _, c := range chores {
		current[c.ID] = struct{}{}
		if _, seen := state.seenChoreIDs[c.ID]; seen {
			continue
		}
		if state.seeding {
			continue
		}
		if outputFormat == outputJSON {
			printJSON(map[string]any{
				"type": "chore_completed", "id": c.ID, "title": c.Title,
				"assignee_id": c.AssigneeID, "ts": ts,
			})
		} else {
			fmt.Printf("[%s] CHORE COMPLETED  %s — assignee %s\n",
				ts, c.Title, c.AssigneeID)
		}
	}
	state.seenChoreIDs = current
}

func pollCalendar(ctx context.Context, client *lib.Client, state *watchState, now time.Time, ts string) {
	today := now.Format(lib.DateFormat)
	events, err := client.ListCalendarEvents(ctx, frameID, today, today, state.timezone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing calendar events: %v\n", ts, err)
		return
	}
	// Only track events currently in the "soon" window so the set stays small.
	current := make(map[string]struct{})
	for _, e := range events {
		if e.AllDay || e.StartAt == "" {
			continue
		}
		start, err := time.Parse(time.RFC3339, e.StartAt)
		if err != nil {
			continue
		}
		diff := time.Until(start)
		if diff <= 0 || diff > time.Hour {
			continue
		}
		current[e.ID] = struct{}{}
		if _, seen := state.seenEventIDs[e.ID]; seen {
			continue
		}
		if state.seeding {
			continue
		}
		if outputFormat == outputJSON {
			printJSON(map[string]any{
				"type": "event_soon", "id": e.ID, "title": e.Title,
				"start_at": e.StartAt, "minutes_until": int(diff.Minutes()), "ts": ts,
			})
		} else {
			fmt.Printf("[%s] EVENT SOON       %s starts in %d min\n",
				ts, e.Title, int(diff.Minutes()))
		}
	}
	state.seenEventIDs = current
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().IntVar(&watchInterval, "interval", 60, "Poll interval in seconds")
	watchCmd.Flags().StringVar(&watchResources, "resources", resourceAll, "Comma-separated resources to watch: rewards,chores,calendar,lists,routines")
	watchCmd.Flags().BoolVar(&watchPersist, "persist", false, "Persist reward deduplication state to disk across restarts (~/.skylight/poller-state.json)")
	registerEnumFlagCompletion(watchCmd, "resources",
		"rewards", "chores", "calendar", "lists", "routines", "all")
}
