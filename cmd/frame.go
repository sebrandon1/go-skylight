package cmd

import (
	"github.com/spf13/cobra"
)

var frameCmd = &cobra.Command{
	Use:   "frame",
	Short: "Frame and device info commands",
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

func init() {
	frameCmd.AddCommand(frameListCmd)
	frameCmd.AddCommand(frameInfoCmd)
	frameCmd.AddCommand(frameDevicesCmd)
	frameCmd.AddCommand(frameAvatarsCmd)
	frameCmd.AddCommand(frameColorsCmd)
}
