package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	alarmDeviceID  string
	alarmID        string
	alarmLabel     string
	alarmTime      string
	alarmDays      []string
	alarmRRule     string
	alarmSound     string
	alarmVolume    int
	alarmSnoozable bool
	alarmEnabled   bool
)

const (
	flagAlarmLabel     = "label"
	flagAlarmTime      = "time"
	flagAlarmDays      = "days"
	flagAlarmRRule     = "rrule"
	flagAlarmSound     = "sound"
	flagAlarmVolume    = "volume"
	flagAlarmSnoozable = "snoozable"
	flagAlarmEnabled   = "enabled"
	flagAlarmID        = "alarm-id"
	flagAlarmDeviceID  = "device-id"
)

var alarmCmd = &cobra.Command{
	Use:   "alarm",
	Short: "Skylight Buddy alarm management commands",
	Long: `Manage alarms on a Skylight Buddy device.

When --device-id is omitted, the frame's Buddy is used automatically; if the
frame has more than one Buddy, pass --device-id (see "frame devices").

  # List alarms, then add a weekday alarm
  skylight alarm list --output table
  skylight alarm create --label School --time 06:30 --days mon,tue,wed,thu,fri

  # Turn an alarm off
  skylight alarm update --alarm-id 123 --enabled=false`,
}

var alarmListCmd = &cobra.Command{
	Use:   subList,
	Short: "List alarms on a Buddy device",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, deviceID, err := alarmClientAndDevice(cmd.Context())
		if err != nil {
			return err
		}

		alarms, err := client.ListAlarms(cmd.Context(), frameID, deviceID)
		if err != nil {
			return fmt.Errorf("listing alarms: %w", err)
		}

		printOutput(alarms)
		return nil
	},
}

var alarmCreateCmd = &cobra.Command{
	Use:   subCreate,
	Short: "Create an alarm on a Buddy device",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := alarmDataFromFlags(cmd)
		if err != nil {
			return err
		}
		if data.Enabled == nil {
			enabled := true
			data.Enabled = &enabled
		}

		client, deviceID, err := alarmClientAndDevice(cmd.Context())
		if err != nil {
			return err
		}

		alarm, err := client.CreateAlarm(cmd.Context(), frameID, deviceID, data)
		if err != nil {
			return fmt.Errorf("creating alarm: %w", err)
		}
		if alarm == nil {
			printSuccess(`Alarm created; run "alarm list" to see it`)
			return nil
		}

		printOutput([]lib.Alarm{*alarm})
		return nil
	},
}

var alarmUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update an alarm on a Buddy device",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := alarmDataFromFlags(cmd)
		if err != nil {
			return err
		}
		if data == (lib.AlarmData{}) {
			return fmt.Errorf("nothing to update: set at least one of --label, --time, --days, --rrule, --sound, --volume, --snoozable, --enabled")
		}

		client, deviceID, err := alarmClientAndDevice(cmd.Context())
		if err != nil {
			return err
		}

		alarm, err := client.UpdateAlarm(cmd.Context(), frameID, deviceID, alarmID, data)
		if err != nil {
			return fmt.Errorf("updating alarm: %w", err)
		}
		if alarm == nil {
			printSuccess("Alarm updated")
			return nil
		}

		printOutput([]lib.Alarm{*alarm})
		return nil
	},
}

var alarmDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete an alarm from a Buddy device",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete alarm %s", alarmID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete alarm %s?", alarmID)) {
			return nil
		}

		client, deviceID, err := alarmClientAndDevice(cmd.Context())
		if err != nil {
			return err
		}

		if err := client.DeleteAlarm(cmd.Context(), frameID, deviceID, alarmID); err != nil {
			return fmt.Errorf("deleting alarm: %w", err)
		}

		printSuccess("Alarm deleted successfully")
		return nil
	},
}

func alarmClientAndDevice(ctx context.Context) (*lib.Client, string, error) {
	if err := requireFrameID(); err != nil {
		return nil, "", err
	}

	client, err := getClient()
	if err != nil {
		return nil, "", err
	}

	if alarmDeviceID != "" {
		return client, alarmDeviceID, nil
	}

	devices, err := client.ListDevices(ctx, frameID)
	if err != nil {
		return nil, "", fmt.Errorf("listing devices to find Buddy: %w", err)
	}
	deviceID, err := pickBuddyDevice(devices)
	if err != nil {
		return nil, "", err
	}
	return client, deviceID, nil
}

func pickBuddyDevice(devices []lib.Device) (string, error) {
	var buddies []string
	for _, d := range devices {
		if d.Role == lib.DeviceRoleBuddy {
			buddies = append(buddies, d.ID)
		}
	}
	switch len(buddies) {
	case 0:
		return "", fmt.Errorf("no Buddy device found on frame %s; pass --device-id", frameID)
	case 1:
		return buddies[0], nil
	default:
		return "", fmt.Errorf("multiple Buddy devices found (%s); pass --device-id", strings.Join(buddies, ", "))
	}
}

func alarmDataFromFlags(cmd *cobra.Command) (lib.AlarmData, error) {
	var data lib.AlarmData
	flags := cmd.Flags()

	if flags.Changed(flagAlarmLabel) {
		data.Label = &alarmLabel
	}
	if flags.Changed(flagAlarmTime) {
		t, err := time.Parse("15:04", alarmTime)
		if err != nil {
			return data, fmt.Errorf("invalid --time %q: use 24-hour HH:MM (e.g. 06:30)", alarmTime)
		}
		normalized := t.Format("15:04")
		data.Time = &normalized
	}
	if flags.Changed(flagAlarmDays) {
		rule, err := lib.AlarmRRule(alarmDays)
		if err != nil {
			return data, fmt.Errorf("invalid --days: %w", err)
		}
		data.RRule = &rule
	}
	if flags.Changed(flagAlarmRRule) {
		rule := strings.TrimPrefix(alarmRRule, "RRULE:")
		data.RRule = &rule
	}
	if flags.Changed(flagAlarmSound) {
		data.Sound = &alarmSound
	}
	if flags.Changed(flagAlarmVolume) {
		if alarmVolume < 0 || alarmVolume > 100 {
			return data, fmt.Errorf("invalid --volume %d: must be between 0 and 100", alarmVolume)
		}
		data.Volume = &alarmVolume
	}
	if flags.Changed(flagAlarmSnoozable) {
		data.Snoozable = &alarmSnoozable
	}
	if flags.Changed(flagAlarmEnabled) {
		data.Enabled = &alarmEnabled
	}
	return data, nil
}

func addAlarmFieldFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringVar(&alarmLabel, flagAlarmLabel, "", "Alarm label")
	f.StringVar(&alarmTime, flagAlarmTime, "", "Alarm time in 24-hour HH:MM (e.g. 06:30)")
	f.StringSliceVar(&alarmDays, flagAlarmDays, nil, "Days the alarm repeats (e.g. mon,tue,wed,thu,fri)")
	f.StringVar(&alarmRRule, flagAlarmRRule, "", "Raw recurrence rule (e.g. FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=SA,SU)")
	f.StringVar(&alarmSound, flagAlarmSound, "", "Alarm sound")
	f.IntVar(&alarmVolume, flagAlarmVolume, 0, "Alarm volume (0-100)")
	f.BoolVar(&alarmSnoozable, flagAlarmSnoozable, false, "Whether the alarm can be snoozed")
	f.BoolVar(&alarmEnabled, flagAlarmEnabled, false, "Whether the alarm is enabled (create defaults to enabled)")
	cmd.MarkFlagsMutuallyExclusive(flagAlarmDays, flagAlarmRRule)
}

func init() {
	alarmCmd.PersistentFlags().StringVar(&alarmDeviceID, flagAlarmDeviceID, "", "Buddy device ID (default: the frame's only Buddy)")

	alarmCmd.AddCommand(alarmListCmd)
	alarmCmd.AddCommand(alarmCreateCmd)
	alarmCmd.AddCommand(alarmUpdateCmd)
	alarmCmd.AddCommand(alarmDeleteCmd)

	addAlarmFieldFlags(alarmCreateCmd)
	markFlagRequired(alarmCreateCmd, flagAlarmTime)

	addAlarmFieldFlags(alarmUpdateCmd)
	alarmUpdateCmd.Flags().StringVar(&alarmID, flagAlarmID, "", "Alarm ID to update")
	markFlagRequired(alarmUpdateCmd, flagAlarmID)

	alarmDeleteCmd.Flags().StringVar(&alarmID, flagAlarmID, "", "Alarm ID to delete")
	alarmDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	alarmDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(alarmDeleteCmd, flagAlarmID)

	rootCmd.AddCommand(alarmCmd)
}
