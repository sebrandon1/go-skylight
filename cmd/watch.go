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
	watchResourceMeals    = "meals"
	watchResourcePhotos   = "photos"
)

var (
	watchInterval  int
	watchResources string
	watchPersist   bool

	allWatchResources = []string{watchResourceRewards, watchResourceChores, watchResourceCalendar, watchResourceLists, watchResourceRoutines, watchResourceMeals, watchResourcePhotos}
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
  - meals:    newly scheduled meal sittings (today through end of week)
  - photos:   newly uploaded photos and videos

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
			seenRewardIDs:      make(map[string]struct{}),
			seenChoreIDs:       make(map[string]struct{}),
			seenEventIDs:       make(map[string]struct{}),
			seenListIDs:        make(map[string]struct{}),
			seenRoutineIDs:     make(map[string]struct{}),
			seenMealSittingIDs: make(map[string]struct{}),
			seenPhotoIDs:       make(map[string]struct{}),
			timezone:           frame.TimeZone,
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
	seenRewardIDs      map[string]struct{}
	seenChoreIDs       map[string]struct{}
	seenEventIDs       map[string]struct{}
	seenListIDs        map[string]struct{}
	seenRoutineIDs     map[string]struct{}
	seenMealSittingIDs map[string]struct{}
	seenPhotoIDs       map[string]struct{}
	seeding            bool
	timezone           string
}

func printRedemptionEvent(e lib.RedemptionEvent) {
	ts := e.ObservedAt.Format("15:04:05")
	if outputFormat == outputJSON {
		printJSON(map[string]any{
			subType: "reward_redeemed", "id": e.RewardID, subTitle: e.RewardName,
			subPoints: e.Points, "category_id": e.CategoryID, "child_name": e.ChildName, "ts": ts,
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
	return parseResourceFilter(s, allWatchResources)
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
		case watchResourceMeals:
			wg.Add(1)
			go func() {
				defer wg.Done()
				pollMealSittings(ctx, client, state, now, ts)
			}()
		case watchResourcePhotos:
			wg.Add(1)
			go func() {
				defer wg.Done()
				pollPhotos(ctx, client, state, ts)
			}()
		}
	}
	wg.Wait()
}

func pollRewards(ctx context.Context, client *lib.Client, state *watchState, ts string) {
	// seen set is replaced each tick so memory stays bounded.
	state.seenRewardIDs = pollAndDiff(ts,
		func() ([]lib.Reward, error) { return client.ListRewards(ctx, frameID) },
		func(r lib.Reward) string { return r.ID },
		func(r lib.Reward) bool { return r.Redeemed },
		func(r lib.Reward) {
			if outputFormat == outputJSON {
				printJSON(map[string]any{
					subType: "reward_redeemed", "id": r.ID, subTitle: r.Title,
					subPoints: r.Points, "category_id": r.CategoryID, "ts": ts,
				})
			} else {
				fmt.Printf("[%s] REWARD REDEEMED  %s (%d pts) — category %s\n",
					ts, r.Title, r.Points, r.CategoryID)
			}
		},
		state.seenRewardIDs, state.seeding, watchResourceRewards,
	)
}

func pollLists(ctx context.Context, client *lib.Client, state *watchState, ts string) {
	state.seenListIDs = pollAndDiff(ts,
		func() ([]lib.List, error) { return client.ListLists(ctx, frameID) },
		func(l lib.List) string { return l.ID },
		nil,
		func(l lib.List) {
			if outputFormat == outputJSON {
				printJSON(map[string]any{subType: "list_created", "id": l.ID, subTitle: l.Title, "ts": ts})
			} else {
				fmt.Printf("[%s] LIST CREATED     %s\n", ts, l.Title)
			}
		},
		state.seenListIDs, state.seeding, watchResourceLists,
	)
}

func pollRoutines(ctx context.Context, client *lib.Client, state *watchState, ts string) {
	state.seenRoutineIDs = pollAndDiff(ts,
		func() ([]lib.Routine, error) { return client.ListRoutines(ctx, frameID) },
		func(r lib.Routine) string { return r.ID },
		nil,
		func(r lib.Routine) {
			if outputFormat == outputJSON {
				printJSON(map[string]any{subType: "routine_created", "id": r.ID, subTitle: r.Title, "ts": ts})
			} else {
				fmt.Printf("[%s] ROUTINE CREATED  %s\n", ts, r.Title)
			}
		},
		state.seenRoutineIDs, state.seeding, watchResourceRoutines,
	)
}

func pollMealSittings(ctx context.Context, client *lib.Client, state *watchState, now time.Time, ts string) {
	today := now.Format(lib.DateFormat)
	monday, _ := weekStart("")
	sunday := monday.AddDate(0, 0, 6).Format(lib.DateFormat)
	state.seenMealSittingIDs = pollAndDiff(ts,
		func() ([]lib.MealSitting, error) {
			return client.ListMealSittings(ctx, frameID, lib.MealSittingListOptions{DateMin: today, DateMax: sunday})
		},
		func(s lib.MealSitting) string { return s.ID },
		nil,
		func(s lib.MealSitting) {
			if outputFormat == outputJSON {
				printJSON(map[string]any{subType: "meal_scheduled", "id": s.ID, "summary": s.Summary, subDate: s.Date, "ts": ts})
			} else {
				fmt.Printf("[%s] MEAL SCHEDULED   %s on %s\n", ts, s.Summary, s.Date)
			}
		},
		state.seenMealSittingIDs, state.seeding, watchResourceMeals,
	)
}

func pollPhotos(ctx context.Context, client *lib.Client, state *watchState, ts string) {
	state.seenPhotoIDs = pollAndDiff(ts,
		func() ([]lib.Photo, error) {
			var all []lib.Photo
			pageToken := ""
			for {
				page, next, err := client.ListPhotos(ctx, frameID, lib.PhotoListOptions{PageToken: pageToken})
				if err != nil {
					return nil, err
				}
				all = append(all, page...)
				if next == "" {
					return all, nil
				}
				pageToken = next
			}
		},
		func(p lib.Photo) string { return p.ID },
		nil,
		func(p lib.Photo) {
			if outputFormat == outputJSON {
				printJSON(map[string]any{subType: "photo_added", "id": p.ID, "asset_type": p.AssetType, "ts": ts})
			} else {
				fmt.Printf("[%s] PHOTO ADDED      %s (%s)\n", ts, p.ID, p.AssetType)
			}
		},
		state.seenPhotoIDs, state.seeding, watchResourcePhotos,
	)
}

func pollChores(ctx context.Context, client *lib.Client, state *watchState, now time.Time, ts string) {
	today := now.Format(lib.DateFormat)
	state.seenChoreIDs = pollAndDiff(ts,
		func() ([]lib.Chore, error) {
			return client.ListChores(ctx, frameID, lib.ChoreListOptions{After: today, Before: today, Status: lib.ChoreStatusComplete})
		},
		func(c lib.Chore) string { return c.ID },
		nil,
		func(c lib.Chore) {
			if outputFormat == outputJSON {
				printJSON(map[string]any{
					subType: "chore_completed", "id": c.ID, subTitle: c.Title,
					"assignee_id": c.AssigneeID, "ts": ts,
				})
			} else {
				fmt.Printf("[%s] CHORE COMPLETED  %s — assignee %s\n", ts, c.Title, c.AssigneeID)
			}
		},
		state.seenChoreIDs, state.seeding, watchResourceChores,
	)
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
				subType: "event_soon", "id": e.ID, subTitle: e.Title,
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
	watchCmd.Flags().StringVar(&watchResources, "resources", resourceAll, "Comma-separated resources to watch: rewards,chores,calendar,lists,routines,meals,photos")
	watchCmd.Flags().BoolVar(&watchPersist, "persist", false, "Persist reward deduplication state to disk across restarts (~/.skylight/poller-state.json)")
	registerEnumFlagCompletion(watchCmd, "resources",
		"rewards", "chores", "calendar", "lists", "routines", "meals", "photos", "all")
}
