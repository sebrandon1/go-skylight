package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	currentAlbumID         int
	screensaverShowWeather bool
	screensaverShowEvents  bool
	frameDeviceID          string
	hideTabs               []string
	showTabs               []string
)

var frameCmd = &cobra.Command{
	Use:   subFrame,
	Short: "Frame and device info commands",
	Long: `Inspect and configure the Skylight frame device itself.

Covers account-level frame listing, per-frame info/devices, the
available avatar/color palettes used elsewhere (e.g. family member
categories), and switching the active photo slideshow album.

  # Find your frame ID, then check its info
  skylight frame list
  skylight frame info --frame-id 12345678`,
}

var frameListCmd = &cobra.Command{
	Use:   subList,
	Short: "List all frames",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}

		frames, err := client.ListFrames(cmd.Context())
		if err != nil {
			return fmt.Errorf("listing frames: %w", err)
		}

		printOutput(frames)
		return nil
	},
}

var frameInfoCmd = &cobra.Command{
	Use:   subInfo,
	Short: "Get frame information",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		frame, err := client.GetFrame(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("getting frame: %w", err)
		}

		printJSON(frame)
		return nil
	},
}

var frameDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		devices, err := client.ListDevices(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing devices: %w", err)
		}

		printOutput(devices)
		return nil
	},
}

var frameTabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "Show which navigation tabs are visible on each device",
	Long: `Show the navigation tab visibility of each device on the frame.

Tab visibility is set per device. Use frame update --hide-tabs/--show-tabs
to change it.

  skylight frame tabs --frame-id 12345678
  skylight frame tabs --frame-id 12345678 --device-id 987 -o table`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ids := []string{frameDeviceID}
		if frameDeviceID == "" {
			devices, err := client.ListDevices(cmd.Context(), frameID)
			if err != nil {
				return fmt.Errorf("listing devices: %w", err)
			}
			ids = deviceIDs(devices)
		}

		configs := make([]lib.DeviceConfig, len(ids))
		fns := make([]func() error, len(ids))
		for i, id := range ids {
			fns[i] = func() error {
				cfg, err := client.GetDeviceConfig(cmd.Context(), frameID, id)
				if err != nil {
					return fmt.Errorf("getting tabs for device %s: %w", id, err)
				}
				configs[i] = *cfg
				return nil
			}
		}
		if err := runConcurrent(fns...); err != nil {
			return err
		}

		printOutput(configs)
		return nil
	},
}

var frameAvatarsCmd = &cobra.Command{
	Use:   "avatars",
	Short: "List available avatars",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}

		avatars, err := client.GetAvatars(cmd.Context())
		if err != nil {
			return fmt.Errorf("getting avatars: %w", err)
		}

		printOutput(avatars)
		return nil
	},
}

var frameColorsCmd = &cobra.Command{
	Use:   "colors",
	Short: "List available colors",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}

		colors, err := client.GetColors(cmd.Context())
		if err != nil {
			return fmt.Errorf("getting colors: %w", err)
		}

		printOutput(colors)
		return nil
	},
}

var frameListAlbumsCmd = &cobra.Command{
	Use:   "list-albums",
	Short: "List photo albums for a frame",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		albums, err := client.ListAlbums(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing albums: %w", err)
		}

		printOutput(albums)
		return nil
	},
}

var frameSetAlbumCmd = &cobra.Command{
	Use:   "set-album",
	Short: "Set the active slideshow album by album ID (-1 for all photos)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.SetCurrentAlbum(cmd.Context(), frameID, currentAlbumID); err != nil {
			return fmt.Errorf("setting current album: %w", err)
		}

		// #265: honor global --quiet like other mutation confirmations
		printSuccessf("Current album set to %d\n", currentAlbumID)
		return nil
	},
}

var frameUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update frame screensaver settings and device tab visibility",
	Long: `Patch frame-level screensaver settings and per-device navigation tabs.

Only flags that are explicitly provided are sent to the API, so you can
update a single setting without affecting the others.

Tabs: ` + strings.Join(lib.HideableTabs(), ", ") + `. Tab visibility is set
per device; --device-id is needed only when the frame has more than one
device. Hiding a tab does not delete any of its data.

  # Enable weather on the photo screensaver
  skylight frame update --frame-id 12345678 --screensaver-show-weather=true

  # Disable calendar events on the photo screensaver
  skylight frame update --frame-id 12345678 --screensaver-show-events=false

  # Hide the meals and recipes tabs, and show photos again
  skylight frame update --frame-id 12345678 --hide-tabs meals,recipes --show-tabs photos`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		tabs, err := tabChanges(hideTabs, showTabs)
		if err != nil {
			return err
		}

		opts := lib.UpdateFrameSettingsOptions{}
		if cmd.Flags().Changed(flagScreensaverShowWeather) {
			opts.ScreensaverShowWeather = &screensaverShowWeather
		}
		if cmd.Flags().Changed(flagScreensaverShowEvents) {
			opts.ScreensaverShowEvents = &screensaverShowEvents
		}
		if opts == (lib.UpdateFrameSettingsOptions{}) && len(tabs) == 0 {
			return fmt.Errorf("nothing to update: pass --%s, --%s, --%s or --%s",
				flagScreensaverShowWeather, flagScreensaverShowEvents, flagHideTabs, flagShowTabs)
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if opts != (lib.UpdateFrameSettingsOptions{}) {
			if err := client.UpdateFrameSettings(cmd.Context(), frameID, opts); err != nil {
				return fmt.Errorf("updating frame settings: %w", err)
			}
			printSuccessf("Frame settings updated\n")
		}

		if len(tabs) > 0 {
			deviceID, err := resolveDeviceID(cmd.Context(), client, frameDeviceID)
			if err != nil {
				return err
			}
			if _, err := client.UpdateDeviceTabs(cmd.Context(), frameID, deviceID, tabs); err != nil {
				return fmt.Errorf("updating device tabs: %w", err)
			}
			printSuccessf("Tabs updated on device %s\n", deviceID)
		}
		return nil
	},
}

// tabChanges merges --hide-tabs and --show-tabs into a tab -> shown map,
// rejecting unknown tab names and tabs listed in both.
func tabChanges(hide, show []string) (map[string]bool, error) {
	valid := lib.HideableTabs()
	tabs := make(map[string]bool, len(hide)+len(show))
	for _, set := range []struct {
		names []string
		shown bool
	}{{hide, false}, {show, true}} {
		for _, name := range set.names {
			name = strings.ToLower(strings.TrimSpace(name))
			if name == "" {
				continue
			}
			if !slices.Contains(valid, name) {
				return nil, fmt.Errorf("unknown tab %q: must be one of %s", name, strings.Join(valid, ", "))
			}
			if prev, ok := tabs[name]; ok && prev != set.shown {
				return nil, fmt.Errorf("tab %q is in both --%s and --%s", name, flagHideTabs, flagShowTabs)
			}
			tabs[name] = set.shown
		}
	}
	return tabs, nil
}

// resolveDeviceID returns deviceID when set; otherwise the frame's only
// device. A frame with several devices needs an explicit --device-id.
func resolveDeviceID(ctx context.Context, client *lib.Client, deviceID string) (string, error) {
	if deviceID != "" {
		return deviceID, nil
	}
	devices, err := client.ListDevices(ctx, frameID)
	if err != nil {
		return "", fmt.Errorf("listing devices: %w", err)
	}
	switch len(devices) {
	case 0:
		return "", fmt.Errorf("frame %s has no devices", frameID)
	case 1:
		return devices[0].ID, nil
	}
	return "", fmt.Errorf("frame has %d devices (%s): pass --%s", len(devices), strings.Join(deviceIDs(devices), ", "), flagDeviceID)
}

func deviceIDs(devices []lib.Device) []string {
	ids := make([]string, len(devices))
	for i, d := range devices {
		ids[i] = d.ID
	}
	return ids
}

func init() {
	frameCmd.AddCommand(frameListCmd)
	frameCmd.AddCommand(frameInfoCmd)
	frameCmd.AddCommand(frameDevicesCmd)
	frameCmd.AddCommand(frameTabsCmd)
	frameCmd.AddCommand(frameAvatarsCmd)
	frameCmd.AddCommand(frameColorsCmd)
	frameCmd.AddCommand(frameListAlbumsCmd)
	frameCmd.AddCommand(frameSetAlbumCmd)
	frameCmd.AddCommand(frameUpdateCmd)

	frameSetAlbumCmd.Flags().IntVar(&currentAlbumID, "album-id", 0, "Album ID to display (-1 for all photos)")
	markFlagRequired(frameSetAlbumCmd, "album-id")

	frameTabsCmd.Flags().StringVar(&frameDeviceID, flagDeviceID, "", "Only show this device (default: all devices on the frame)")

	frameUpdateCmd.Flags().BoolVar(&screensaverShowWeather, flagScreensaverShowWeather, false, "Show weather on the photo screensaver (requires Calendar Plus)")
	frameUpdateCmd.Flags().BoolVar(&screensaverShowEvents, flagScreensaverShowEvents, false, "Show upcoming calendar events on the photo screensaver (requires Calendar Plus)")
	frameUpdateCmd.Flags().StringSliceVar(&hideTabs, flagHideTabs, nil, "Tabs to hide on the device (comma-separated: "+strings.Join(lib.HideableTabs(), ", ")+")")
	frameUpdateCmd.Flags().StringSliceVar(&showTabs, flagShowTabs, nil, "Tabs to show on the device (comma-separated)")
	frameUpdateCmd.Flags().StringVar(&frameDeviceID, flagDeviceID, "", "Device to change tabs on (required if the frame has more than one device)")
}
