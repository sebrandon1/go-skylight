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
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		monday, _ := weekStart("")
		sunday := monday.AddDate(0, 0, 6)
		today := time.Now().Format(lib.DateFormat)

		client := getClient()

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

		// Frame info is only needed for TimeZone, used by the calendar events
		// call. Fetch it inside this goroutine (rather than serially before
		// the fan-out) so it runs concurrently with the chores/lists calls
		// instead of blocking them.
		wg.Add(1)
		go func() {
			defer wg.Done()
			frame := getFrameOrFail(client, frameID)
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
			fatal("listing calendar events", evtErr)
		}
		if choreErr != nil {
			fatal("listing chores", choreErr)
		}
		if listErr != nil {
			fatal("listing lists", listErr)
		}
		if mealErr != nil {
			fatal("listing meal sittings", mealErr)
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
			return
		}

		printHomeTable(events, chores, lists, meals, monday)
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
