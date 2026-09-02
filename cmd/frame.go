package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	currentAlbumID         int
	screensaverShowWeather bool
	screensaverShowEvents  bool
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
	Short: "Update frame screensaver settings",
	Long: `Patch frame-level screensaver settings.

Only flags that are explicitly provided are sent to the API, so you can
update a single setting without affecting the others.

  # Enable weather on the photo screensaver
  skylight frame update --frame-id 12345678 --screensaver-show-weather=true

  # Disable calendar events on the photo screensaver
  skylight frame update --frame-id 12345678 --screensaver-show-events=false`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		opts := lib.UpdateFrameSettingsOptions{}
		if cmd.Flags().Changed("screensaver-show-weather") {
			opts.ScreensaverShowWeather = &screensaverShowWeather
		}
		if cmd.Flags().Changed("screensaver-show-events") {
			opts.ScreensaverShowEvents = &screensaverShowEvents
		}

		if err := client.UpdateFrameSettings(cmd.Context(), frameID, opts); err != nil {
			return fmt.Errorf("updating frame settings: %w", err)
		}

		printSuccessf("Frame settings updated\n")
		return nil
	},
}

func init() {
	frameCmd.AddCommand(frameListCmd)
	frameCmd.AddCommand(frameInfoCmd)
	frameCmd.AddCommand(frameDevicesCmd)
	frameCmd.AddCommand(frameAvatarsCmd)
	frameCmd.AddCommand(frameColorsCmd)
	frameCmd.AddCommand(frameSetAlbumCmd)
	frameCmd.AddCommand(frameUpdateCmd)

	frameSetAlbumCmd.Flags().IntVar(&currentAlbumID, "album-id", 0, "Album ID to display (-1 for all photos)")
	markFlagRequired(frameSetAlbumCmd, "album-id")

	frameUpdateCmd.Flags().BoolVar(&screensaverShowWeather, "screensaver-show-weather", false, "Show weather on the photo screensaver (requires Calendar Plus)")
	frameUpdateCmd.Flags().BoolVar(&screensaverShowEvents, "screensaver-show-events", false, "Show upcoming calendar events on the photo screensaver (requires Calendar Plus)")
}
