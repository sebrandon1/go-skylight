package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

const (
	watchResourceRewards  = "rewards"
	watchResourceChores   = "chores"
	watchResourceCalendar = "calendar"
)

var (
	watchInterval  int
	watchResources string

	allWatchResources = []string{watchResourceRewards, watchResourceChores, watchResourceCalendar}

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
  - rewards: newly redeemed rewards
  - chores:  chores newly marked complete
  - calendar: events starting within the next hour

Press Ctrl+C to stop.`,
	Run: func(cmd *cobra.Command, args []string) {
		if watchInterval < 1 {
			fmt.Fprintln(os.Stderr, "Error: --interval must be at least 1")
			os.Exit(1)
		}

		requireFrameID()
		client := getClient()

		resources := parseWatchResources(watchResources)

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		ticker := time.NewTicker(time.Duration(watchInterval) * time.Second)
		defer ticker.Stop()

		state := &watchState{
			seenRewardIDs: make(map[string]struct{}),
			seenChoreIDs:  make(map[string]struct{}),
			seenEventIDs:  make(map[string]struct{}),
		}

		// Seed state silently so existing items aren't reported as new.
		state.seeding = true
		poll(client, state, resources)
		state.seeding = false

		fmt.Printf("Watching %s (interval: %ds). Press Ctrl+C to stop.\n\n",
			strings.Join(resources, ", "), watchInterval)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nStopped.")
				return
			case <-ticker.C:
				poll(client, state, resources)
			}
		}
	},
}

type watchState struct {
	seenRewardIDs map[string]struct{}
	seenChoreIDs  map[string]struct{}
	seenEventIDs  map[string]struct{}
	seeding       bool
}

func parseWatchResources(s string) []string {
	if s == "" || s == "all" {
		return allWatchResources
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
		fmt.Fprintln(os.Stderr, "Error: no valid resources specified")
		os.Exit(1)
	}
	return out
}

func poll(client *lib.Client, state *watchState, resources []string) {
	now := time.Now()
	ts := now.Format("15:04:05")
	for _, r := range resources {
		switch r {
		case watchResourceRewards:
			pollRewards(client, state, ts)
		case watchResourceChores:
			pollChores(client, state, now, ts)
		case watchResourceCalendar:
			pollCalendar(client, state, now, ts)
		}
	}
}

func pollRewards(client *lib.Client, state *watchState, ts string) {
	rewards, err := client.ListRewards(frameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing rewards: %v\n", ts, err)
		return
	}
	for _, r := range rewards {
		if !r.Redeemed {
			continue
		}
		if _, seen := state.seenRewardIDs[r.ID]; seen {
			continue
		}
		state.seenRewardIDs[r.ID] = struct{}{}
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
}

func pollChores(client *lib.Client, state *watchState, now time.Time, ts string) {
	today := now.Format(lib.DateFormat)
	chores, err := client.ListChores(frameID, lib.ChoreListOptions{After: today, Before: today, Status: lib.ChoreStatusComplete})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing chores: %v\n", ts, err)
		return
	}
	for _, c := range chores {
		if _, seen := state.seenChoreIDs[c.ID]; seen {
			continue
		}
		state.seenChoreIDs[c.ID] = struct{}{}
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
}

func pollCalendar(client *lib.Client, state *watchState, now time.Time, ts string) {
	today := now.Format(lib.DateFormat)
	events, err := client.ListCalendarEvents(frameID, today, today)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing calendar events: %v\n", ts, err)
		return
	}
	for _, e := range events {
		if _, seen := state.seenEventIDs[e.ID]; seen {
			continue
		}
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
		state.seenEventIDs[e.ID] = struct{}{}
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
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().IntVar(&watchInterval, "interval", 60, "Poll interval in seconds")
	watchCmd.Flags().StringVar(&watchResources, "resources", "all", "Comma-separated resources to watch: rewards,chores,calendar")
}
