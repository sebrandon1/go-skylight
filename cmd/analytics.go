package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var analyticsDays int

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Family activity statistics over a time period",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()
		client := getClient()

		now := time.Now()
		start := now.AddDate(0, 0, -analyticsDays)
		startStr := start.Format(lib.DateFormat)
		endStr := now.Format(lib.DateFormat)

		frame := getFrameOrFail(client, frameID)

		var (
			categories []lib.Category
			chores     []lib.Chore
			rewards    []lib.Reward
			points     []lib.RewardPointEntry
			events     []lib.CalendarEvent

			catErr    error
			choreErr  error
			rewardErr error
			pointsErr error
			eventErr  error

			wg sync.WaitGroup
		)

		wg.Add(5)
		go func() {
			defer wg.Done()
			categories, catErr = client.ListCategories(frameID)
		}()
		go func() {
			defer wg.Done()
			chores, choreErr = client.ListChores(frameID, lib.ChoreListOptions{
				After:  startStr,
				Before: endStr,
			})
		}()
		go func() {
			defer wg.Done()
			rewards, rewardErr = client.ListRewards(frameID)
		}()
		go func() {
			defer wg.Done()
			points, pointsErr = client.GetRewardPoints(frameID)
		}()
		go func() {
			defer wg.Done()
			events, eventErr = client.ListCalendarEvents(frameID, startStr, endStr, frame.TimeZone)
		}()
		wg.Wait()

		if catErr != nil {
			fatal("listing categories", catErr)
		}
		if choreErr != nil {
			fatal("listing chores", choreErr)
		}
		if rewardErr != nil {
			fatal("listing rewards", rewardErr)
		}
		if pointsErr != nil {
			fatal("getting reward points", pointsErr)
		}
		if eventErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: calendar events unavailable: %v\n", eventErr)
			events = nil
		}

		catNames := buildCatNames(categories)
		stats := computeAnalytics(chores, rewards, points, events, catNames, start, now)

		if outputFormat == outputJSON {
			printJSON(stats)
			return
		}

		printAnalyticsText(stats)
	},
}

type AnalyticsStats struct {
	PeriodDays    int                   `json:"period_days"`
	StartDate     string                `json:"start_date"`
	EndDate       string                `json:"end_date"`
	Assignees     []AssigneeStats       `json:"assignees"`
	TopChores     []ChoreFrequency      `json:"top_chores"`
	Rewards       RewardStats           `json:"rewards"`
	CalendarStats CalendarActivityStats `json:"calendar"`
}

type AssigneeStats struct {
	Name            string  `json:"name"`
	TotalChores     int     `json:"total_chores"`
	CompletedChores int     `json:"completed_chores"`
	CompletionRate  float64 `json:"completion_rate"`
	PointBalance    int     `json:"point_balance"`
}

type ChoreFrequency struct {
	Title     string `json:"title"`
	Count     int    `json:"count"`
	Completed int    `json:"completed"`
}

type RewardStats struct {
	Total    int `json:"total"`
	Redeemed int `json:"redeemed"`
}

type CalendarActivityStats struct {
	TotalEvents  int     `json:"total_events"`
	EventsPerDay float64 `json:"events_per_day"`
}

type choreCount struct {
	total     int
	completed int
}

func incrChoreCount(m map[string]*choreCount, key, status string) {
	if m[key] == nil {
		m[key] = &choreCount{}
	}
	m[key].total++
	if status == lib.ChoreStatusComplete {
		m[key].completed++
	}
}

func tallyChores(chores []lib.Chore) (assigneeTotals map[string]*choreCount, choreTitleMap map[string]*choreCount) {
	assigneeTotals = make(map[string]*choreCount)
	choreTitleMap = make(map[string]*choreCount)
	for _, c := range chores {
		if c.AssigneeID != "" {
			incrChoreCount(assigneeTotals, c.AssigneeID, c.Status)
		}
		incrChoreCount(choreTitleMap, c.Title, c.Status)
	}
	return
}

func buildAssigneeStats(assigneeTotals map[string]*choreCount, catNames map[string]string, pointsByCategory map[int]int) []AssigneeStats {
	var assignees []AssigneeStats
	for id, counts := range assigneeTotals {
		name := catNames[id]
		if name == "" {
			name = id
		}
		var rate float64
		if counts.total > 0 {
			rate = float64(counts.completed) / float64(counts.total) * 100
		}
		var pointBalance int
		if catID, err := strconv.Atoi(id); err == nil {
			pointBalance = pointsByCategory[catID]
		} else {
			// Category IDs are strings in the chores/categories API but ints on
			// reward points — non-numeric IDs would silently show 0 points (#253).
			fmt.Fprintf(os.Stderr, "Warning: analytics: skipping point balance for category id %q: not numeric\n", id)
		}
		assignees = append(assignees, AssigneeStats{
			Name:            name,
			TotalChores:     counts.total,
			CompletedChores: counts.completed,
			CompletionRate:  rate,
			PointBalance:    pointBalance,
		})
	}
	sort.Slice(assignees, func(i, j int) bool { return assignees[i].Name < assignees[j].Name })
	return assignees
}

func buildTopChores(choreTitleMap map[string]*choreCount) []ChoreFrequency {
	var topChores []ChoreFrequency
	for title, counts := range choreTitleMap {
		topChores = append(topChores, ChoreFrequency{Title: title, Count: counts.total, Completed: counts.completed})
	}
	sort.Slice(topChores, func(i, j int) bool { return topChores[i].Count > topChores[j].Count })
	if len(topChores) > 5 {
		topChores = topChores[:5]
	}
	return topChores
}

func computeAnalytics(
	chores []lib.Chore,
	rewards []lib.Reward,
	points []lib.RewardPointEntry,
	events []lib.CalendarEvent,
	catNames map[string]string,
	start, end time.Time,
) AnalyticsStats {
	days := int(end.Sub(start).Hours()/24) + 1

	pointsByCategory := make(map[int]int, len(points))
	for _, p := range points {
		pointsByCategory[p.CategoryID] = p.CurrentPointBalance
	}

	assigneeTotals, choreTitleMap := tallyChores(chores)
	assignees := buildAssigneeStats(assigneeTotals, catNames, pointsByCategory)
	topChores := buildTopChores(choreTitleMap)

	var redeemed int
	for _, r := range rewards {
		if r.Redeemed {
			redeemed++
		}
	}

	var eventsPerDay float64
	if days > 0 {
		eventsPerDay = float64(len(events)) / float64(days)
	}

	return AnalyticsStats{
		PeriodDays: days,
		StartDate:  start.Format(lib.DateFormat),
		EndDate:    end.Format(lib.DateFormat),
		Assignees:  assignees,
		TopChores:  topChores,
		Rewards: RewardStats{
			Total:    len(rewards),
			Redeemed: redeemed,
		},
		CalendarStats: CalendarActivityStats{
			TotalEvents:  len(events),
			EventsPerDay: eventsPerDay,
		},
	}
}

func printAnalyticsText(s AnalyticsStats) {
	fmt.Printf("Analytics: %s → %s (%d days)\n\n", s.StartDate, s.EndDate, s.PeriodDays)

	fmt.Println("Family Members:")
	if len(s.Assignees) == 0 {
		fmt.Println("  (none)")
	} else {
		w := newTableWriter()
		fmt.Fprintln(w, "NAME\tCOMPLETED/TOTAL\tCOMPLETION RATE\tPOINTS")
		for _, a := range s.Assignees {
			fmt.Fprintf(w, "%s\t%d/%d\t%.1f%%\t%d\n",
				a.Name, a.CompletedChores, a.TotalChores, a.CompletionRate, a.PointBalance)
		}
		w.Flush()
	}

	fmt.Println("\nTop Chores:")
	if len(s.TopChores) == 0 {
		fmt.Println("  (none)")
	} else {
		w := newTableWriter()
		fmt.Fprintln(w, "TITLE\tTIMES\tCOMPLETED")
		for _, c := range s.TopChores {
			fmt.Fprintf(w, "%s\t%d\t%d\n", c.Title, c.Count, c.Completed)
		}
		w.Flush()
	}

	fmt.Printf("\nRewards:   %d total, %d redeemed\n", s.Rewards.Total, s.Rewards.Redeemed)
	fmt.Printf("Calendar:  %d events (%.1f/day)\n", s.CalendarStats.TotalEvents, s.CalendarStats.EventsPerDay)
}

func init() {
	rootCmd.AddCommand(analyticsCmd)
	analyticsCmd.Flags().IntVar(&analyticsDays, "days", 30, "Number of days to include in the report")
}
