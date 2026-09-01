package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	profileID    string
	profileName  string
	profileColor string
)

var profileCmd = &cobra.Command{
	Use:   subProfile,
	Short: "Profile (household member) management commands",
	Long: `Manage household member profiles on a Skylight frame.

Profiles represent individual family members and are used as assignee IDs on
chores, rewards, and routines.

  # List all profiles, then create a new one
  skylight profile list --output table
  skylight profile create --name "Alex" --color "#4A90D9"`,
}

var profileListCmd = &cobra.Command{
	Use:   subList,
	Short: "List household member profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		profiles, err := client.ListProfiles(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing profiles: %w", err)
		}

		printOutput(profiles)
		return nil
	},
}

var profileCreateCmd = &cobra.Command{
	Use:   subCreate,
	Short: "Create a household member profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		linked := true // *bool required: omitempty would drop false, so nil means "no change" on updates
		data := lib.CategoryData{
			Name:            profileName,
			LinkedToProfile: &linked,
		}
		if cmd.Flags().Changed("color") {
			data.Color = profileColor
		}

		profile, err := client.CreateCategory(cmd.Context(), frameID, data)
		if err != nil {
			return fmt.Errorf("creating profile: %w", err)
		}

		printOutput([]lib.Category{*profile})
		return nil
	},
}

var profileUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update a household member profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.CategoryData{}
		if cmd.Flags().Changed("name") {
			data.Name = profileName
		}
		if cmd.Flags().Changed("color") {
			data.Color = profileColor
		}

		profile, err := client.UpdateCategory(cmd.Context(), frameID, profileID, data)
		if err != nil {
			return fmt.Errorf("updating profile: %w", err)
		}

		printOutput([]lib.Category{*profile})
		return nil
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a household member profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete profile %s", profileID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete profile %s?", profileID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteCategory(cmd.Context(), frameID, profileID); err != nil {
			return fmt.Errorf("deleting profile: %w", err)
		}

		printSuccess("Profile deleted successfully")
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileUpdateCmd)
	profileCmd.AddCommand(profileDeleteCmd)

	profileCreateCmd.Flags().StringVar(&profileName, "name", "", "Profile name")
	profileCreateCmd.Flags().StringVar(&profileColor, "color", "", "Profile color (hex, e.g. #FF0000)")
	markFlagRequired(profileCreateCmd, "name")

	profileUpdateCmd.Flags().StringVar(&profileID, "profile-id", "", "Profile ID to update")
	profileUpdateCmd.Flags().StringVar(&profileName, "name", "", "Profile name")
	profileUpdateCmd.Flags().StringVar(&profileColor, "color", "", "Profile color (hex, e.g. #FF0000)")
	markFlagRequired(profileUpdateCmd, "profile-id")

	profileDeleteCmd.Flags().StringVar(&profileID, "profile-id", "", "Profile ID")
	profileDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	profileDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(profileDeleteCmd, "profile-id")
}
