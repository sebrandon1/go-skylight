package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var currentAlbumID int

var frameCmd = &cobra.Command{
	Use:   "frame",
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
	Use:   "list",
	Short: "List all frames",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()

		frames, err := client.ListFrames()
		if err != nil {
			fatal("listing frames", err)
		}

		printOutput(frames)
	},
}

var frameInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get frame information",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		frame, err := client.GetFrame(frameID)
		if err != nil {
			fatal("getting frame", err)
		}

		printJSON(frame)
	},
}

var frameDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List devices",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		devices, err := client.ListDevices(frameID)
		if err != nil {
			fatal("listing devices", err)
		}

		printOutput(devices)
	},
}

var frameAvatarsCmd = &cobra.Command{
	Use:   "avatars",
	Short: "List available avatars",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()

		avatars, err := client.GetAvatars()
		if err != nil {
			fatal("getting avatars", err)
		}

		printOutput(avatars)
	},
}

var frameColorsCmd = &cobra.Command{
	Use:   "colors",
	Short: "List available colors",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()

		colors, err := client.GetColors()
		if err != nil {
			fatal("getting colors", err)
		}

		printOutput(colors)
	},
}

var frameSetAlbumCmd = &cobra.Command{
	Use:   "set-album",
	Short: "Set the active slideshow album by album ID (-1 for all photos)",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		if err := client.SetCurrentAlbum(frameID, currentAlbumID); err != nil {
			fatal("setting current album", err)
		}

		fmt.Printf("Current album set to %d\n", currentAlbumID)
	},
}

func init() {
	frameCmd.AddCommand(frameListCmd)
	frameCmd.AddCommand(frameInfoCmd)
	frameCmd.AddCommand(frameDevicesCmd)
	frameCmd.AddCommand(frameAvatarsCmd)
	frameCmd.AddCommand(frameColorsCmd)
	frameCmd.AddCommand(frameSetAlbumCmd)

	frameSetAlbumCmd.Flags().IntVar(&currentAlbumID, "album-id", 0, "Album ID to display (-1 for all photos)")
	markFlagRequired(frameSetAlbumCmd, "album-id")
}
