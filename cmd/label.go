package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	labelID    string
	labelName  string
	labelColor string
)

var labelCmd = &cobra.Command{
	Use:   subLabel,
	Short: "Label management commands",
	Long: `Manage event and task labels on a Skylight frame.

Labels categorize calendar events and tasks by topic (e.g. Sports, School,
Work).

  # List all labels, then create a new one
  skylight label list --output table
  skylight label create --name "School" --color "#4A90D9"`,
}

var labelListCmd = &cobra.Command{
	Use:   subList,
	Short: "List event/task labels",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		labels, err := client.ListLabels(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing labels: %w", err)
		}

		printOutput(labels)
		return nil
	},
}

var labelCreateCmd = &cobra.Command{
	Use:   subCreate,
	Short: "Create an event/task label",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		linked := false // *bool required: omitempty would drop false, so nil means "no change" on updates
		data := lib.CategoryData{
			Name:            labelName,
			LinkedToProfile: &linked,
		}
		if cmd.Flags().Changed("color") {
			data.Color = labelColor
		}

		label, err := client.CreateCategory(cmd.Context(), frameID, data)
		if err != nil {
			return fmt.Errorf("creating label: %w", err)
		}

		printOutput([]lib.Category{*label})
		return nil
	},
}

var labelUpdateCmd = &cobra.Command{
	Use:   subUpdate,
	Short: "Update an event/task label",
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
			data.Name = labelName
		}
		if cmd.Flags().Changed("color") {
			data.Color = labelColor
		}

		label, err := client.UpdateCategory(cmd.Context(), frameID, labelID, data)
		if err != nil {
			return fmt.Errorf("updating label: %w", err)
		}

		printOutput([]lib.Category{*label})
		return nil
	},
}

var labelDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete an event/task label",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete label %s", labelID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete label %s?", labelID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteCategory(cmd.Context(), frameID, labelID); err != nil {
			return fmt.Errorf("deleting label: %w", err)
		}

		printSuccess("Label deleted successfully")
		return nil
	},
}

func init() {
	labelCmd.AddCommand(labelListCmd)
	labelCmd.AddCommand(labelCreateCmd)
	labelCmd.AddCommand(labelUpdateCmd)
	labelCmd.AddCommand(labelDeleteCmd)

	labelCreateCmd.Flags().StringVar(&labelName, "name", "", "Label name")
	labelCreateCmd.Flags().StringVar(&labelColor, "color", "", "Label color (hex, e.g. #FF0000)")
	markFlagRequired(labelCreateCmd, "name")

	labelUpdateCmd.Flags().StringVar(&labelID, "label-id", "", "Label ID to update")
	labelUpdateCmd.Flags().StringVar(&labelName, "name", "", "Label name")
	labelUpdateCmd.Flags().StringVar(&labelColor, "color", "", "Label color (hex, e.g. #FF0000)")
	markFlagRequired(labelUpdateCmd, "label-id")

	labelDeleteCmd.Flags().StringVar(&labelID, "label-id", "", "Label ID")
	labelDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	labelDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(labelDeleteCmd, "label-id")
}
