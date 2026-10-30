package cmd

import (
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

// maxMessageLen is the Skylight app's character limit for a text message.
const maxMessageLen = 220

// defaultMessageListDays is how far ahead `message list` looks when --end-date is unset.
const defaultMessageListDays = 30

var (
	messageID        string
	messageText      string
	messageTo        []string
	messageSendAt    string
	messageRepeat    string
	messageStartDate string
	messageEndDate   string
)

var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "Household message commands",
	Long: `Send text messages to household members on a Skylight Calendar, now or
at a scheduled time. Requires Calendar Plus.

Recipients are family member category IDs (see "skylight profile list").
"message list" shows one row per delivery, so a recurring message appears
once for each occurrence in the date window. Recurring deliveries keep a
fixed UTC time, so they shift by an hour across a daylight saving change.

  # Send now, then schedule a weekly reminder
  skylight message send --text "Dinner is ready!" --to 123,456
  skylight message send --text "Soccer practice" --to 123 \
    --send-at 2026-11-07T08:00 --repeat "FREQ=WEEKLY;BYDAY=SA"`,
}

var messageListCmd = &cobra.Command{
	Use:   subList,
	Short: "List messages delivering in a date window",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}
		after, before, err := messageListWindow(time.Now())
		if err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		messages, err := client.ListMessages(cmd.Context(), frameID, after, before)
		if err != nil {
			return fmt.Errorf("listing messages: %w", err)
		}

		maybeLoadCatNames(cmd.Context(), client)
		printOutput(messages)
		return nil
	},
}

var messageSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send or schedule a message to household members",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}
		if err := validateMessageText(messageText); err != nil {
			return err
		}
		if messageRepeat != "" && messageSendAt == "" {
			return errors.New("--repeat requires --send-at")
		}
		to, err := messageRecipients()
		if err != nil {
			return err
		}
		data := lib.MessageData{Body: messageText, RRule: messageRepeat, CategoryIDs: to}
		if messageSendAt != "" {
			at, err := parseSendAt(messageSendAt)
			if err != nil {
				return err
			}
			data.DeliverAt = at
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		msg, err := client.SendMessage(cmd.Context(), frameID, data)
		if err != nil {
			return fmt.Errorf("sending message: %w", err)
		}

		maybeLoadCatNames(cmd.Context(), client)
		printOutput([]lib.Message{*msg})
		return nil
	},
}

var messageUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update a message",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		data := lib.MessageData{}
		if cmd.Flags().Changed("text") {
			if err := validateMessageText(messageText); err != nil {
				return err
			}
			data.Body = messageText
		}
		if cmd.Flags().Changed("to") {
			to, err := messageRecipients()
			if err != nil {
				return err
			}
			data.CategoryIDs = to
		}
		if cmd.Flags().Changed("send-at") {
			at, err := parseSendAt(messageSendAt)
			if err != nil {
				return err
			}
			data.DeliverAt = at
		}
		if cmd.Flags().Changed("repeat") {
			data.RRule = messageRepeat
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		msg, err := client.UpdateMessage(cmd.Context(), frameID, messageID, data)
		if err != nil {
			return fmt.Errorf("updating message: %w", err)
		}

		maybeLoadCatNames(cmd.Context(), client)
		printOutput([]lib.Message{*msg})
		return nil
	},
}

var messageDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a message",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete message %s", messageID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete message %s?", messageID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteMessage(cmd.Context(), frameID, messageID); err != nil {
			return fmt.Errorf("deleting message: %w", err)
		}

		printSuccess("Message deleted successfully")
		return nil
	},
}

// messageListWindow returns the after/before dates for `message list`,
// defaulting to today through defaultMessageListDays ahead.
func messageListWindow(now time.Time) (after, before string, err error) {
	after, before = messageStartDate, messageEndDate
	if err := validateDate(after); err != nil {
		return "", "", err
	}
	if err := validateDate(before); err != nil {
		return "", "", err
	}
	if after == "" {
		after = now.Format(lib.DateFormat)
	}
	if before == "" {
		start, _ := time.Parse(lib.DateFormat, after) // validated above
		before = start.AddDate(0, 0, defaultMessageListDays).Format(lib.DateFormat)
	}
	if before < after {
		return "", "", fmt.Errorf("--end-date %s is before --start-date %s", before, after)
	}
	return after, before, nil
}

// sendAtLayouts are the accepted --send-at formats. Layouts without a UTC
// offset are interpreted in the local time zone.
var sendAtLayouts = []string{"2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02 15:04"}

// parseSendAt converts a --send-at value to RFC 3339.
func parseSendAt(s string) (string, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format(time.RFC3339), nil
	}
	for _, layout := range sendAtLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t.Format(time.RFC3339), nil
		}
	}
	return "", fmt.Errorf("invalid --send-at %q: use YYYY-MM-DDTHH:MM[:SS] (local time) or RFC 3339", s)
}

// messageRecipients returns the --to category IDs, rejecting an empty list.
func messageRecipients() ([]string, error) {
	to := filterEmptyStrings(messageTo)
	if len(to) == 0 {
		return nil, errors.New("--to requires at least one category ID")
	}
	return to, nil
}

func validateMessageText(s string) error {
	if s == "" {
		return errors.New("--text must not be empty")
	}
	if n := utf8.RuneCountInString(s); n > maxMessageLen {
		return fmt.Errorf("--text is %d characters; the limit is %d", n, maxMessageLen)
	}
	return nil
}

func init() {
	messageCmd.AddCommand(messageListCmd)
	messageCmd.AddCommand(messageSendCmd)
	messageCmd.AddCommand(messageUpdateCmd)
	messageCmd.AddCommand(messageDeleteCmd)

	messageListCmd.Flags().StringVar(&messageStartDate, "start-date", "", "First delivery date to include, YYYY-MM-DD (default today)")
	messageListCmd.Flags().StringVar(&messageEndDate, "end-date", "", fmt.Sprintf("Last delivery date to include, YYYY-MM-DD (default start + %d days)", defaultMessageListDays))

	messageSendCmd.Flags().StringVar(&messageText, "text", "", fmt.Sprintf("Message text (max %d characters)", maxMessageLen))
	messageSendCmd.Flags().StringSliceVar(&messageTo, "to", nil, "Recipient category IDs (comma-separated)")
	messageSendCmd.Flags().StringVar(&messageSendAt, "send-at", "", "Schedule delivery: YYYY-MM-DDTHH:MM (local time) or RFC 3339 (default: send now)")
	messageSendCmd.Flags().StringVar(&messageRepeat, "repeat", "", `Recurrence rule for a scheduled message, e.g. "FREQ=WEEKLY;BYDAY=SA"`)
	markFlagRequired(messageSendCmd, "text")
	markFlagRequired(messageSendCmd, "to")

	messageUpdateCmd.Flags().StringVar(&messageID, "message-id", "", "Message ID to update")
	messageUpdateCmd.Flags().StringVar(&messageText, "text", "", fmt.Sprintf("Message text (max %d characters)", maxMessageLen))
	messageUpdateCmd.Flags().StringSliceVar(&messageTo, "to", nil, "Recipient category IDs (comma-separated; replaces the list)")
	messageUpdateCmd.Flags().StringVar(&messageSendAt, "send-at", "", "Delivery time: YYYY-MM-DDTHH:MM (local time) or RFC 3339")
	messageUpdateCmd.Flags().StringVar(&messageRepeat, "repeat", "", `Recurrence rule, e.g. "FREQ=WEEKLY;BYDAY=SA"`)
	markFlagRequired(messageUpdateCmd, "message-id")

	messageDeleteCmd.Flags().StringVar(&messageID, "message-id", "", "Message ID")
	messageDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	messageDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(messageDeleteCmd, "message-id")
}
