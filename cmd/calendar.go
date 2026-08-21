package cmd

import (
	"fmt"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	calendarStartDate     string
	calendarEndDate       string
	calendarEventID       string
	calendarTitle         string
	calendarStartAt       string
	calendarEndAt         string
	calendarAllDay        bool
	calendarColor         string
	calendarCategoryID    string
	calendarWeekDate      string
	calendarDayDate       string
	calendarCountdownDate string
	calendarSourceID      string
)

var calendarCmd = &cobra.Command{
	Use:   exportResourceCalendar,
	Short: "Calendar event management commands",
	Long: `Create, list, update, and delete calendar events on a Skylight frame.

Events created here appear alongside any connected source calendars
(Google, Apple, etc.) — see "calendar sources" to list those. Use
"calendar week" for a 7-day Mon-Sun view instead of a raw list.

  # List this week's events, then create a new one
  skylight calendar week
  skylight calendar create --title "Dentist" --start-at 2026-06-05T09:00:00Z --end-at 2026-06-05T10:00:00Z`,
}

var calendarListCmd = &cobra.Command{
	Use:   subList,
	Short: "List calendar events",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		for _, f := range []struct {
			name string
			val  string
		}{{"start-date", calendarStartDate}, {"end-date", calendarEndDate}} {
			if cmd.Flags().Changed(f.name) {
				if err := validateDate(f.val); err != nil {
					return err
				}
			}
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		frame, err := getFrameOrFail(ctx, client, frameID)
		if err != nil {
			return err
		}

		events, err := client.ListCalendarEvents(ctx, frameID, calendarStartDate, calendarEndDate, frame.TimeZone)
		if err != nil {
			return fmt.Errorf("listing calendar events: %w", err)
		}

		printOutput(events)
		return nil
	},
}

var calendarCreateCmd = &cobra.Command{
	Use:   subCreate,
	Short: "Create a calendar event",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		allDay := calendarAllDay
		event, err := client.CreateCalendarEvent(cmd.Context(), frameID, lib.CalendarEventData{
			Title:      calendarTitle,
			StartAt:    calendarStartAt,
			EndAt:      calendarEndAt,
			AllDay:     &allDay,
			Color:      calendarColor,
			CategoryID: calendarCategoryID,
		})
		if err != nil {
			return fmt.Errorf("creating calendar event: %w", err)
		}

		printOutput([]lib.CalendarEvent{*event})
		return nil
	},
}

var calendarGetCmd = &cobra.Command{
	Use:   subGet,
	Short: "Get a single calendar event by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		event, err := client.GetCalendarEvent(cmd.Context(), frameID, calendarEventID)
		if err != nil {
			return fmt.Errorf("getting calendar event: %w", err)
		}

		printOutput(event)
		return nil
	},
}

var calendarDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a calendar event",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete calendar event %s", calendarEventID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete calendar event %s?", calendarEventID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteCalendarEvent(cmd.Context(), frameID, calendarEventID); err != nil {
			return fmt.Errorf("deleting calendar event: %w", err)
		}

		printSuccess("Calendar event deleted successfully")
		return nil
	},
}

var sourceCalendarsCmd = &cobra.Command{
	Use:   "sources",
	Short: "List source calendars",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		calendars, err := client.ListSourceCalendars(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing source calendars: %w", err)
		}

		printOutput(calendars)
		return nil
	},
}

var calendarUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update a calendar event",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.CalendarEventData{}
		if cmd.Flags().Changed(subTitle) {
			data.Title = calendarTitle
		}
		if cmd.Flags().Changed("start-at") {
			data.StartAt = calendarStartAt
		}
		if cmd.Flags().Changed("end-at") {
			data.EndAt = calendarEndAt
		}
		if cmd.Flags().Changed("all-day") {
			allDay := calendarAllDay
			data.AllDay = &allDay
		}
		if cmd.Flags().Changed("color") {
			data.Color = calendarColor
		}
		if cmd.Flags().Changed("category-id") {
			data.CategoryID = calendarCategoryID
		}

		event, err := client.UpdateCalendarEvent(cmd.Context(), frameID, calendarEventID, data)
		if err != nil {
			return fmt.Errorf("updating calendar event: %w", err)
		}

		printOutput([]lib.CalendarEvent{*event})
		return nil
	},
}

var calendarCreateCountdownCmd = &cobra.Command{
	Use:   "create-countdown",
	Short: "Create a countdown event",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if err := validateDate(calendarCountdownDate); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		allDayTrue := true
		event, err := client.CreateCalendarEvent(cmd.Context(), frameID, lib.CalendarEventData{
			Title:     calendarTitle,
			StartAt:   calendarCountdownDate,
			AllDay:    &allDayTrue,
			EventType: lib.CalendarEventTypeCountdown,
		})
		if err != nil {
			return fmt.Errorf("creating countdown event: %w", err)
		}

		printOutput([]lib.CalendarEvent{*event})
		return nil
	},
}

var calendarSourceEnableCmd = &cobra.Command{
	Use:   "source-enable",
	Short: "Enable a source calendar",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.UpdateSourceCalendar(cmd.Context(), frameID, calendarSourceID, true); err != nil {
			return fmt.Errorf("enabling source calendar: %w", err)
		}

		printSuccess("Source calendar enabled successfully")
		return nil
	},
}

var calendarSourceDisableCmd = &cobra.Command{
	Use:   "source-disable",
	Short: "Disable a source calendar",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.UpdateSourceCalendar(cmd.Context(), frameID, calendarSourceID, false); err != nil {
			return fmt.Errorf("disabling source calendar: %w", err)
		}

		printSuccess("Source calendar disabled successfully")
		return nil
	},
}

var calendarWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show a 7-day Mon-Sun view of calendar events",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if cmd.Flags().Changed(subDate) {
			if err := validateDate(calendarWeekDate); err != nil {
				return err
			}
		}

		monday, err := weekStart(calendarWeekDate)
		if err != nil {
			return fmt.Errorf("computing week start: %w", err)
		}
		sunday := monday.AddDate(0, 0, 6)

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		frame, err := getFrameOrFail(ctx, client, frameID)
		if err != nil {
			return err
		}

		events, err := client.ListCalendarEvents(
			ctx,
			frameID,
			monday.Format(lib.DateFormat),
			sunday.Format(lib.DateFormat),
			frame.TimeZone,
		)
		if err != nil {
			return fmt.Errorf("listing calendar events: %w", err)
		}

		days := buildCalendarWeeklyView(events, monday)
		printOutput(days)
		return nil
	},
}

var calendarDayCmd = &cobra.Command{
	Use:   "day",
	Short: "Show a single day's calendar events",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		var dateStr string
		if cmd.Flags().Changed(subDate) {
			if err := validateDate(calendarDayDate); err != nil {
				return err
			}
			dateStr = calendarDayDate
		} else {
			dateStr = time.Now().Format(lib.DateFormat)
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		frame, err := getFrameOrFail(ctx, client, frameID)
		if err != nil {
			return err
		}

		events, err := client.ListCalendarEvents(ctx, frameID, dateStr, dateStr, frame.TimeZone)
		if err != nil {
			return fmt.Errorf("listing calendar events: %w", err)
		}

		printOutput(events)
		return nil
	},
}

func init() {
	calendarCmd.AddCommand(calendarListCmd)
	calendarCmd.AddCommand(calendarGetCmd)
	calendarCmd.AddCommand(calendarCreateCmd)
	calendarCmd.AddCommand(calendarCreateCountdownCmd)
	calendarCmd.AddCommand(calendarUpdateCmd)
	calendarCmd.AddCommand(calendarDeleteCmd)
	calendarCmd.AddCommand(sourceCalendarsCmd)
	calendarCmd.AddCommand(calendarSourceEnableCmd)
	calendarCmd.AddCommand(calendarSourceDisableCmd)
	calendarCmd.AddCommand(calendarWeekCmd)
	calendarCmd.AddCommand(calendarDayCmd)

	calendarDayCmd.Flags().StringVar(&calendarDayDate, subDate, "", "Date to show (YYYY-MM-DD, default today)")

	calendarGetCmd.Flags().StringVar(&calendarEventID, "event-id", "", "Event ID to get")
	markFlagRequired(calendarGetCmd, "event-id")

	calendarListCmd.Flags().StringVar(&calendarStartDate, "start-date", "", "Start date filter")
	calendarListCmd.Flags().StringVar(&calendarEndDate, "end-date", "", "End date filter")

	calendarCreateCmd.Flags().StringVar(&calendarTitle, subTitle, "", "Event title")
	calendarCreateCmd.Flags().StringVar(&calendarStartAt, "start-at", "", "Event start time")
	calendarCreateCmd.Flags().StringVar(&calendarEndAt, "end-at", "", "Event end time")
	calendarCreateCmd.Flags().BoolVar(&calendarAllDay, "all-day", false, "All day event")
	calendarCreateCmd.Flags().StringVar(&calendarColor, "color", "", "Event color")
	calendarCreateCmd.Flags().StringVar(&calendarCategoryID, "category-id", "", "Category ID to assign to this event")
	markFlagRequired(calendarCreateCmd, subTitle)
	markFlagRequired(calendarCreateCmd, "start-at")

	calendarUpdateCmd.Flags().StringVar(&calendarEventID, "event-id", "", "Event ID to update")
	calendarUpdateCmd.Flags().StringVar(&calendarTitle, subTitle, "", "Event title")
	calendarUpdateCmd.Flags().StringVar(&calendarStartAt, "start-at", "", "Event start time")
	calendarUpdateCmd.Flags().StringVar(&calendarEndAt, "end-at", "", "Event end time")
	calendarUpdateCmd.Flags().BoolVar(&calendarAllDay, "all-day", false, "All day event")
	calendarUpdateCmd.Flags().StringVar(&calendarColor, "color", "", "Event color")
	calendarUpdateCmd.Flags().StringVar(&calendarCategoryID, "category-id", "", "Category ID to assign to this event")
	markFlagRequired(calendarUpdateCmd, "event-id")

	calendarDeleteCmd.Flags().StringVar(&calendarEventID, "event-id", "", "Event ID to delete")
	calendarDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	calendarDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(calendarDeleteCmd, "event-id")

	calendarCreateCountdownCmd.Flags().StringVar(&calendarTitle, subTitle, "", "Countdown event title")
	calendarCreateCountdownCmd.Flags().StringVar(&calendarCountdownDate, subDate, "", "Target date (YYYY-MM-DD)")
	markFlagRequired(calendarCreateCountdownCmd, subTitle)
	markFlagRequired(calendarCreateCountdownCmd, subDate)

	calendarWeekCmd.Flags().StringVar(&calendarWeekDate, subDate, "", "Week containing this date (YYYY-MM-DD, default current week)")

	calendarSourceEnableCmd.Flags().StringVar(&calendarSourceID, "source-id", "", "Source calendar ID")
	markFlagRequired(calendarSourceEnableCmd, "source-id")

	calendarSourceDisableCmd.Flags().StringVar(&calendarSourceID, "source-id", "", "Source calendar ID")
	markFlagRequired(calendarSourceDisableCmd, "source-id")
}
