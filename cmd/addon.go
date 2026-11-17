package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

const (
	flagDisneyProfilePictures = "disney-profile-pictures"
	flagDisneyScreensaver     = "disney-screensaver"
)

var (
	addonDisneyProfilePictures bool
	addonDisneyScreensaver     bool
)

var addonCmd = &cobra.Command{
	Use:   "addon",
	Short: "Manage frame add-ons and household add-on settings",
}

var addonListCmd = &cobra.Command{
	Use:   subList,
	Short: "List available add-ons and their enabled state",
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
			return fmt.Errorf("listing addons: %w", err)
		}

		if !frame.Plus {
			fmt.Fprintln(os.Stderr, "Warning: Skylight Plus subscription required to access add-ons.")
		}
		if len(frame.FeatureBundle) == 0 {
			fmt.Println("No add-ons found.")
			return nil
		}

		printOutput(frame.FeatureBundle)
		return nil
	},
}

var addonSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Show the household add-on settings (Disney Mode switches)",
	Long: `Show the add-on settings you can change for the household.

The add-ons in "addon list" come with the frame's plan and can't be turned on
or off through the API. What can be changed are the Disney Mode switches:
Disney avatars for profile photos and the Disney screensaver album.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}
		client, err := getClient()
		if err != nil {
			return err
		}
		cfg, err := client.GetHouseholdConfig(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("getting add-on settings: %w", err)
		}
		printOutput(cfg)
		return nil
	},
}

var addonSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Change household add-on settings (Disney Mode switches)",
	Example: `  skylight addon set --disney-screensaver=false
  skylight addon set --disney-profile-pictures=true --disney-screensaver=true`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}
		var update lib.HouseholdConfigUpdate
		if cmd.Flags().Changed(flagDisneyProfilePictures) {
			v := addonDisneyProfilePictures
			update.DisneyProfilePictures = &v
		}
		if cmd.Flags().Changed(flagDisneyScreensaver) {
			v := addonDisneyScreensaver
			update.DisneyScreensaver = &v
		}
		if update.DisneyProfilePictures == nil && update.DisneyScreensaver == nil {
			return fmt.Errorf("nothing to change: pass --%s and/or --%s", flagDisneyProfilePictures, flagDisneyScreensaver)
		}
		client, err := getClient()
		if err != nil {
			return err
		}
		cfg, err := client.UpdateHouseholdConfig(cmd.Context(), frameID, update)
		if err != nil {
			return fmt.Errorf("updating add-on settings: %w", err)
		}
		printOutput(cfg)
		return nil
	},
}

func enabledLabel(b bool) string {
	if b {
		return boolYes
	}
	return boolNo
}

func printHouseholdConfigTable(cfg *lib.HouseholdConfig) {
	w := newTableWriter()
	fmt.Fprintln(w, "SETTING\tENABLED")
	fmt.Fprintf(w, "%s\t%s\n", flagDisneyProfilePictures, enabledLabel(cfg.DisneyProfilePictures))
	fmt.Fprintf(w, "%s\t%s\n", flagDisneyScreensaver, enabledLabel(cfg.DisneyScreensaver))
	w.Flush()
}

func init() {
	addonCmd.AddCommand(addonListCmd)
	addonCmd.AddCommand(addonSettingsCmd)
	addonCmd.AddCommand(addonSetCmd)
	addonSetCmd.Flags().BoolVar(&addonDisneyProfilePictures, flagDisneyProfilePictures, false, "Use Disney avatars for profile photos")
	addonSetCmd.Flags().BoolVar(&addonDisneyScreensaver, flagDisneyScreensaver, false, "Show the Disney album in the screensaver")
}
