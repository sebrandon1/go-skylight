package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	homeNoTasks bool
	homeNoLists bool
	homeNoMeals bool
)

var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "Weekly combined view of events, tasks, and lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		monday, _ := weekStart("")
		sunday := monday.AddDate(0, 0, 6)
		today := time.Now().Format(lib.DateFormat)

		client, err := getClient()
		if err != nil {
			return err
		}

		frame, err := getFrameOrFail(client, frameID)
		if err != nil {
			return err
		}

		var (
			events   []lib.CalendarEvent
			chores   []lib.Chore
			lists    []lib.List
			meals    []lib.MealSitting
			evtErr   error
			choreErr error
			listErr  error
			mealErr  error
			wg       sync.WaitGroup
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			events, evtErr = client.ListCalendarEvents(frameID, monday.Format(lib.DateFormat), sunday.Format(lib.DateFormat), frame.TimeZone)
		}()

		if !homeNoTasks {
			wg.Add(1)
			go func() {
				defer wg.Done()
				chores, choreErr = client.ListChores(frameID, lib.ChoreListOptions{
					After:  today,
					Before: today,
					Status: lib.ChoreStatusPending,
				})
			}()
		}

		if !homeNoLists {
			wg.Add(1)
			go func() {
				defer wg.Done()
				lists, listErr = client.ListLists(frameID)
			}()
		}

		if !homeNoMeals {
			wg.Add(1)
			go func() {
				defer wg.Done()
				meals, mealErr = client.ListMealSittings(frameID, lib.MealSittingListOptions{
					DateMin: monday.Format(lib.DateFormat),
					DateMax: sunday.Format(lib.DateFormat),
				})
			}()
		}

		wg.Wait()

		if evtErr != nil {
			return fmt.Errorf("listing calendar events: %w", evtErr)
		}
		if choreErr != nil {
			return fmt.Errorf("listing chores: %w", choreErr)
		}
		if listErr != nil {
			return fmt.Errorf("listing lists: %w", listErr)
		}
		if mealErr != nil {
			return fmt.Errorf("listing meal sittings: %w", mealErr)
		}

		if outputFormat == outputJSON {
			printJSON(map[string]any{
				"week_start": monday.Format(lib.DateFormat),
				"week_end":   sunday.Format(lib.DateFormat),
				"events":     events,
				"tasks":      chores,
				"lists":      lists,
				"meals":      meals,
			})
			return nil
		}

		printHomeTable(events, chores, lists, meals, monday)
		return nil
	},
}

func printHomeTable(events []lib.CalendarEvent, chores []lib.Chore, lists []lib.List, meals []lib.MealSitting, monday time.Time) {
	fmt.Println("=== EVENTS THIS WEEK ===")
	printCalendarWeekTable(buildCalendarWeeklyView(events, monday))

	if len(chores) > 0 {
		fmt.Println("\n=== PENDING TASKS TODAY ===")
		printChoresTable(chores)
	}

	if len(lists) > 0 {
		fmt.Println("\n=== LISTS ===")
		printListsTable(lists)
	}

	if len(meals) > 0 {
		fmt.Println("\n=== MEALS THIS WEEK ===")
		printMealSittingsTable(meals)
	}
}

func init() {
	rootCmd.AddCommand(homeCmd)

	homeCmd.Flags().BoolVar(&homeNoTasks, "no-tasks", false, "Exclude pending tasks")
	homeCmd.Flags().BoolVar(&homeNoLists, "no-lists", false, "Exclude active lists")
	homeCmd.Flags().BoolVar(&homeNoMeals, "no-meals", false, "Exclude meal sittings")
}
