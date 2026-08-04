package cmd

import (
	"fmt"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	homeNoTasks    bool
	homeNoLists    bool
	homeNoMeals    bool
	homeNoRoutines bool
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

		ctx := cmd.Context()
		frame, err := getFrameOrFail(ctx, client, frameID)
		if err != nil {
			return err
		}

		var (
			events   []lib.CalendarEvent
			chores   []lib.Chore
			lists    []lib.List
			meals    []lib.MealSitting
			routines []lib.Routine
		)

		fns := []func() error{
			func() error {
				var err error
				events, err = client.ListCalendarEvents(ctx, frameID, monday.Format(lib.DateFormat), sunday.Format(lib.DateFormat), frame.TimeZone)
				if err != nil {
					return fmt.Errorf("listing calendar events: %w", err)
				}
				return nil
			},
		}
		if !homeNoTasks {
			fns = append(fns, func() error {
				var err error
				chores, err = client.ListChores(ctx, frameID, lib.ChoreListOptions{
					After:  today,
					Before: today,
					Status: lib.ChoreStatusPending,
				})
				if err != nil {
					return fmt.Errorf("listing chores: %w", err)
				}
				return nil
			})
		}
		if !homeNoLists {
			fns = append(fns, func() error {
				var err error
				lists, err = client.ListLists(ctx, frameID)
				if err != nil {
					return fmt.Errorf("listing lists: %w", err)
				}
				return nil
			})
		}
		if !homeNoMeals {
			fns = append(fns, func() error {
				var err error
				meals, err = client.ListMealSittings(ctx, frameID, lib.MealSittingListOptions{
					DateMin: monday.Format(lib.DateFormat),
					DateMax: sunday.Format(lib.DateFormat),
				})
				if err != nil {
					return fmt.Errorf("listing meal sittings: %w", err)
				}
				return nil
			})
		}
		if !homeNoRoutines {
			fns = append(fns, func() error {
				var err error
				routines, err = client.ListRoutines(ctx, frameID)
				if err != nil && !lib.IsNotFound(err) {
					return fmt.Errorf("listing routines: %w", err)
				}
				return nil
			})
		}

		if err := runConcurrent(fns...); err != nil {
			return err
		}

		if outputFormat == outputJSON {
			printJSON(map[string]any{
				"week_start": monday.Format(lib.DateFormat),
				"week_end":   sunday.Format(lib.DateFormat),
				"events":     events,
				"tasks":      chores,
				"lists":      lists,
				"meals":      meals,
				"routines":   routines,
			})
			return nil
		}

		printHomeTable(events, chores, lists, meals, routines, monday)
		return nil
	},
}

func printHomeTable(events []lib.CalendarEvent, chores []lib.Chore, lists []lib.List, meals []lib.MealSitting, routines []lib.Routine, monday time.Time) {
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

	if len(routines) > 0 {
		fmt.Println("\n=== ROUTINES ===")
		printRoutinesTable(routines)
	}
}

func init() {
	rootCmd.AddCommand(homeCmd)

	homeCmd.Flags().BoolVar(&homeNoTasks, "no-tasks", false, "Exclude pending tasks")
	homeCmd.Flags().BoolVar(&homeNoLists, "no-lists", false, "Exclude active lists")
	homeCmd.Flags().BoolVar(&homeNoMeals, "no-meals", false, "Exclude meal sittings")
	homeCmd.Flags().BoolVar(&homeNoRoutines, "no-routines", false, "Exclude routines")
}
