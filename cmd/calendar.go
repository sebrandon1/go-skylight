package cmd

import (
	"fmt"
	"os"

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
	calendarWeekDate      string
	calendarCountdownDate string
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Calendar event management commands",
}

var calendarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List calendar events",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		for _, f := range []struct {
			name string
			val  string
		}{{"start-date", calendarStartDate}, {"end-date", calendarEndDate}} {
			if cmd.Flags().Changed(f.name) {
				if err := validateDate(f.val); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}

		client := getClient()

		frame := getFrameOrFail(client, frameID)

		events, err := client.ListCalendarEvents(frameID, calendarStartDate, calendarEndDate, frame.TimeZone)
		if err != nil {
			fatal("listing calendar events", err)
		}

		printOutput(events)
	},
}

var calendarCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a calendar event",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		event, err := client.CreateCalendarEvent(frameID, lib.CalendarEventData{
			Title:   calendarTitle,
			StartAt: calendarStartAt,
			EndAt:   calendarEndAt,
			AllDay:  calendarAllDay,
		})
		if err != nil {
			fatal("creating calendar event", err)
		}

		printJSON(event)
	},
}

var calendarDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a calendar event",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.DeleteCalendarEvent(frameID, calendarEventID)
		if err != nil {
			fatal("deleting calendar event", err)
		}

		fmt.Println("Calendar event deleted successfully")
	},
}

var sourceCalendarsCmd = &cobra.Command{
	Use:   "sources",
	Short: "List source calendars",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		calendars, err := client.ListSourceCalendars(frameID)
		if err != nil {
			fatal("listing source calendars", err)
		}

		printOutput(calendars)
	},
}

var calendarUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a calendar event",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.CalendarEventData{}
		if cmd.Flags().Changed("title") {
			data.Title = calendarTitle
		}
		if cmd.Flags().Changed("start-at") {
			data.StartAt = calendarStartAt
		}
		if cmd.Flags().Changed("end-at") {
			data.EndAt = calendarEndAt
		}
		if cmd.Flags().Changed("all-day") {
			data.AllDay = calendarAllDay
		}

		event, err := client.UpdateCalendarEvent(frameID, calendarEventID, data)
		if err != nil {
			fatal("updating calendar event", err)
		}

		printJSON(event)
	},
}

var calendarCreateCountdownCmd = &cobra.Command{
	Use:   "create-countdown",
	Short: "Create a countdown event",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if err := validateDate(calendarCountdownDate); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		client := getClient()

		event, err := client.CreateCalendarEvent(frameID, lib.CalendarEventData{
			Title:     calendarTitle,
			StartAt:   calendarCountdownDate,
			AllDay:    true,
			EventType: lib.CalendarEventTypeCountdown,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: creating countdown event: %v\n", err)
			os.Exit(1)
		}

		printJSON(event)
	},
}

var calendarWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show a 7-day Mon-Sun view of calendar events",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if cmd.Flags().Changed("date") {
			if err := validateDate(calendarWeekDate); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		monday, err := weekStart(calendarWeekDate)
		if err != nil {
			fatal("computing week start", err)
		}
		sunday := monday.AddDate(0, 0, 6)

		client := getClient()

		frame := getFrameOrFail(client, frameID)

		events, err := client.ListCalendarEvents(
			frameID,
			monday.Format(lib.DateFormat),
			sunday.Format(lib.DateFormat),
			frame.TimeZone,
		)
		if err != nil {
			fatal("listing calendar events", err)
		}

		days := buildCalendarWeeklyView(events, monday)
		printOutput(days)
	},
}

func init() {
	calendarCmd.AddCommand(calendarListCmd)
	calendarCmd.AddCommand(calendarCreateCmd)
	calendarCmd.AddCommand(calendarCreateCountdownCmd)
	calendarCmd.AddCommand(calendarUpdateCmd)
	calendarCmd.AddCommand(calendarDeleteCmd)
	calendarCmd.AddCommand(sourceCalendarsCmd)
	calendarCmd.AddCommand(calendarWeekCmd)

	calendarListCmd.Flags().StringVar(&calendarStartDate, "start-date", "", "Start date filter")
	calendarListCmd.Flags().StringVar(&calendarEndDate, "end-date", "", "End date filter")

	calendarCreateCmd.Flags().StringVar(&calendarTitle, "title", "", "Event title")
	calendarCreateCmd.Flags().StringVar(&calendarStartAt, "start-at", "", "Event start time")
	calendarCreateCmd.Flags().StringVar(&calendarEndAt, "end-at", "", "Event end time")
	calendarCreateCmd.Flags().BoolVar(&calendarAllDay, "all-day", false, "All day event")
	calendarCreateCmd.MarkFlagRequired("title")    //nolint:errcheck
	calendarCreateCmd.MarkFlagRequired("start-at") //nolint:errcheck

	calendarUpdateCmd.Flags().StringVar(&calendarEventID, "event-id", "", "Event ID to update")
	calendarUpdateCmd.Flags().StringVar(&calendarTitle, "title", "", "Event title")
	calendarUpdateCmd.Flags().StringVar(&calendarStartAt, "start-at", "", "Event start time")
	calendarUpdateCmd.Flags().StringVar(&calendarEndAt, "end-at", "", "Event end time")
	calendarUpdateCmd.Flags().BoolVar(&calendarAllDay, "all-day", false, "All day event")
	calendarUpdateCmd.MarkFlagRequired("event-id") //nolint:errcheck

	calendarDeleteCmd.Flags().StringVar(&calendarEventID, "event-id", "", "Event ID to delete")
	calendarDeleteCmd.MarkFlagRequired("event-id") //nolint:errcheck

	calendarCreateCountdownCmd.Flags().StringVar(&calendarTitle, "title", "", "Countdown event title")
	calendarCreateCountdownCmd.Flags().StringVar(&calendarCountdownDate, "date", "", "Target date (YYYY-MM-DD)")
	calendarCreateCountdownCmd.MarkFlagRequired("title") //nolint:errcheck
	calendarCreateCountdownCmd.MarkFlagRequired("date")  //nolint:errcheck

	calendarWeekCmd.Flags().StringVar(&calendarWeekDate, "date", "", "Week containing this date (YYYY-MM-DD, default current week)")
}
