package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var frameCmd = &cobra.Command{
	Use:   "frame",
	Short: "Frame and device info commands",
}

var frameInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get frame information",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		frame, err := client.GetFrame(frameID)
		if err != nil {
			fmt.Printf("Error getting frame: %v\n", err)
			os.Exit(1)
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
			fmt.Printf("Error listing devices: %v\n", err)
			os.Exit(1)
		}

		printJSON(devices)
	},
}

var frameAvatarsCmd = &cobra.Command{
	Use:   "avatars",
	Short: "List available avatars",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()

		avatars, err := client.GetAvatars()
		if err != nil {
			fmt.Printf("Error getting avatars: %v\n", err)
			os.Exit(1)
		}

		printJSON(avatars)
	},
}

var frameColorsCmd = &cobra.Command{
	Use:   "colors",
	Short: "List available colors",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()

		colors, err := client.GetColors()
		if err != nil {
			fmt.Printf("Error getting colors: %v\n", err)
			os.Exit(1)
		}

		printJSON(colors)
	},
}

func init() {
	frameCmd.AddCommand(frameInfoCmd)
	frameCmd.AddCommand(frameDevicesCmd)
	frameCmd.AddCommand(frameAvatarsCmd)
	frameCmd.AddCommand(frameColorsCmd)
}
