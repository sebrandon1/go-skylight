package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	calendarStartDate string
	calendarEndDate   string
	calendarEventID   string
	calendarTitle     string
	calendarStartAt   string
	calendarEndAt     string
	calendarAllDay    bool
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

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		events, err := client.ListCalendarEvents(frameID, calendarStartDate, calendarEndDate)
		if err != nil {
			fmt.Printf("Error listing calendar events: %v\n", err)
			os.Exit(1)
		}

		printJSON(events)
	},
}

var calendarCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a calendar event",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		event, err := client.CreateCalendarEvent(frameID, lib.CalendarEventData{
			Title:   calendarTitle,
			StartAt: calendarStartAt,
			EndAt:   calendarEndAt,
			AllDay:  calendarAllDay,
		})
		if err != nil {
			fmt.Printf("Error creating calendar event: %v\n", err)
			os.Exit(1)
		}

		printJSON(event)
	},
}

var calendarDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a calendar event",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.DeleteCalendarEvent(frameID, calendarEventID)
		if err != nil {
			fmt.Printf("Error deleting calendar event: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Calendar event deleted successfully")
	},
}

var sourceCalendarsCmd = &cobra.Command{
	Use:   "sources",
	Short: "List source calendars",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		calendars, err := client.ListSourceCalendars(frameID)
		if err != nil {
			fmt.Printf("Error listing source calendars: %v\n", err)
			os.Exit(1)
		}

		printJSON(calendars)
	},
}

func init() {
	calendarCmd.AddCommand(calendarListCmd)
	calendarCmd.AddCommand(calendarCreateCmd)
	calendarCmd.AddCommand(calendarDeleteCmd)
	calendarCmd.AddCommand(sourceCalendarsCmd)

	calendarListCmd.Flags().StringVar(&calendarStartDate, "start-date", "", "Start date filter")
	calendarListCmd.Flags().StringVar(&calendarEndDate, "end-date", "", "End date filter")

	calendarCreateCmd.Flags().StringVar(&calendarTitle, "title", "", "Event title")
	calendarCreateCmd.Flags().StringVar(&calendarStartAt, "start-at", "", "Event start time")
	calendarCreateCmd.Flags().StringVar(&calendarEndAt, "end-at", "", "Event end time")
	calendarCreateCmd.Flags().BoolVar(&calendarAllDay, "all-day", false, "All day event")

	calendarDeleteCmd.Flags().StringVar(&calendarEventID, "event-id", "", "Event ID to delete")
}
