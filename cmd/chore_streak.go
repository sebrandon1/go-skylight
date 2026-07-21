package cmd

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var streakDays int

// ChoreStreakStats holds per-assignee chore completion statistics over a window.
type ChoreStreakStats struct {
	AssigneeID      string  `json:"assignee_id"`
	AssigneeName    string  `json:"assignee_name"`
	CurrentStreak   int     `json:"current_streak"`
	LongestStreak   int     `json:"longest_streak"`
	TotalChores     int     `json:"total_chores"`
	CompletedChores int     `json:"completed_chores"`
	CompletionRate  float64 `json:"completion_rate_pct"`
}

type choreDay struct {
	date       string
	assigneeID string
}

// computeChoreStreaks derives per-assignee streak stats from a flat chore list.
// dates must be a sorted (ascending) slice of YYYY-MM-DD strings covering the window.
// catNames maps category ID → display name; missing IDs fall back to the raw ID.
// Days with no chores for a given assignee are skipped (they neither break nor extend streaks).
func computeChoreStreaks(chores []lib.Chore, dates []string, catNames map[string]string) []ChoreStreakStats {
	total := map[choreDay]int{}
	done := map[choreDay]int{}
	assigneeSet := map[string]bool{}

	for _, c := range chores {
		if c.AssigneeID == "" {
			continue
		}
		key := choreDay{c.DueDate, c.AssigneeID}
		total[key]++
		if c.Status == lib.ChoreStatusComplete {
			done[key]++
		}
		assigneeSet[c.AssigneeID] = true
	}

	results := make([]ChoreStreakStats, 0, len(assigneeSet))
	for aid := range assigneeSet {
		var totalC, doneC, longestStreak, curStreak int

		for _, date := range dates {
			key := choreDay{date, aid}
			t := total[key]
			d := done[key]
			totalC += t
			doneC += d

			if t == 0 {
				continue
			}
			if d == t {
				curStreak++
				if curStreak > longestStreak {
					longestStreak = curStreak
				}
			} else {
				curStreak = 0
			}
		}

		// Current streak: walk backwards from the most-recent date.
		currentStreak := 0
		for i := len(dates) - 1; i >= 0; i-- {
			key := choreDay{dates[i], aid}
			t := total[key]
			if t == 0 {
				continue
			}
			if done[key] == t {
				currentStreak++
			} else {
				break
			}
		}

		rate := 0.0
		if totalC > 0 {
			rate = float64(doneC) / float64(totalC) * 100
		}

		name := catNames[aid]
		if name == "" {
			name = aid
		}
		results = append(results, ChoreStreakStats{
			AssigneeID:      aid,
			AssigneeName:    name,
			CurrentStreak:   currentStreak,
			LongestStreak:   longestStreak,
			TotalChores:     totalC,
			CompletedChores: doneC,
			CompletionRate:  rate,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].AssigneeName < results[j].AssigneeName
	})

	return results
}

var choreStreakCmd = &cobra.Command{
	Use:   "streak",
	Short: "Show chore completion streaks per family member",
	Long: `Compute chore completion streaks and stats for each family member.

Reports current streak, longest streak, and overall completion rate over
the requested window (default: last 30 days). Days with no assigned chores
for a given member are ignored and do not break streaks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		now := time.Now()
		end := now.Format(lib.DateFormat)
		start := now.AddDate(0, 0, -(streakDays - 1)).Format(lib.DateFormat)

		var (
			categories []lib.Category
			catErr     error
			chores     []lib.Chore
			choreErr   error
			wg         sync.WaitGroup
		)

		wg.Add(2)
		go func() {
			defer wg.Done()
			categories, catErr = client.ListCategories(frameID)
		}()
		go func() {
			defer wg.Done()
			chores, choreErr = client.ListChores(frameID, lib.ChoreListOptions{
				After:       start,
				Before:      end,
				IncludeLate: true,
			})
		}()
		wg.Wait()

		if catErr != nil {
			return fmt.Errorf("listing categories: %w", catErr)
		}
		if choreErr != nil {
			return fmt.Errorf("listing chores: %w", choreErr)
		}

		catNames := buildCatNames(categories)

		dates := make([]string, 0, streakDays)
		for i := -(streakDays - 1); i <= 0; i++ {
			dates = append(dates, now.AddDate(0, 0, i).Format(lib.DateFormat))
		}

		results := computeChoreStreaks(chores, dates, catNames)

		printOutput(results)
		return nil
	},
}

func init() {
	choreCmd.AddCommand(choreStreakCmd)
	choreStreakCmd.Flags().IntVar(&streakDays, "days", 30, "Number of days to analyze (default 30)")
}
