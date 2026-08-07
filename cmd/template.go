package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	templateName      string
	templateResources []string
	templateStartDate string
	templateDryRun    bool
)

func templateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".skylight", "templates"), nil
}

type choreTemplate struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Points      int    `json:"points,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Recurring   bool   `json:"recurring,omitempty"`
}

type rewardTemplate struct {
	Title     string `json:"title"`
	Points    int    `json:"points,omitempty"`
	EmojiIcon string `json:"emoji_icon,omitempty"`
}

type frameTemplate struct {
	Name      string           `json:"name"`
	CreatedAt string           `json:"created_at"`
	Chores    []choreTemplate  `json:"chores,omitempty"`
	Rewards   []rewardTemplate `json:"rewards,omitempty"`
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Chore and reward template management",
	Long: `Save and apply named templates of chore and reward configurations.

Templates are stored locally in ~/.skylight/templates/ as JSON files.
Use "template save" to capture the current frame's chores and rewards,
then "template apply" to recreate them on any frame at any time.

  # Save the current frame's chores and rewards as "weekly"
  skylight template save --name weekly

  # Apply the template to the current frame, starting Monday
  skylight template apply --name weekly --start-date 2026-08-10`,
}

var templateSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save current frame chores/rewards as a named template",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		dir, err := templateDir()
		if err != nil {
			return err
		}

		saveChores := len(templateResources) == 0 || slices.Contains(templateResources, exportResourceChores)
		saveRewards := len(templateResources) == 0 || slices.Contains(templateResources, exportResourceRewards)

		tmpl := frameTemplate{
			Name:      templateName,
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		ctx := cmd.Context()
		if saveChores {
			chores, err := client.ListChores(ctx, frameID, lib.ChoreListOptions{})
			if err != nil {
				return fmt.Errorf("fetching chores: %w", err)
			}
			for _, c := range chores {
				tmpl.Chores = append(tmpl.Chores, choreTemplate{
					Title:       c.Title,
					Description: c.Description,
					Points:      c.Points,
					AssigneeID:  c.AssigneeID,
					DueDate:     c.DueDate,
					Recurring:   c.Recurring,
				})
			}
		}

		if saveRewards {
			rewards, err := client.ListRewards(ctx, frameID)
			if err != nil {
				return fmt.Errorf("fetching rewards: %w", err)
			}
			for _, r := range rewards {
				tmpl.Rewards = append(tmpl.Rewards, rewardTemplate{
					Title:     r.Title,
					Points:    r.Points,
					EmojiIcon: r.EmojiIcon,
				})
			}
		}

		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("creating template directory: %w", err)
		}

		data, err := json.MarshalIndent(tmpl, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding template: %w", err)
		}

		path := filepath.Join(dir, templateName+".json")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return fmt.Errorf("writing template: %w", err)
		}

		printSuccessf("Template %q saved (%d chore(s), %d reward(s))\n", templateName, len(tmpl.Chores), len(tmpl.Rewards))
		return nil
	},
}

var templateApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Create chores/rewards on the current frame from a saved template",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		dir, err := templateDir()
		if err != nil {
			return err
		}

		path := filepath.Join(dir, templateName+".json")
		raw, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("template %q not found (run 'template list' to see available templates)", templateName)
			}
			return fmt.Errorf("reading template: %w", err)
		}

		var tmpl frameTemplate
		if err := json.Unmarshal(raw, &tmpl); err != nil {
			return fmt.Errorf("parsing template: %w", err)
		}

		applyChores := len(templateResources) == 0 || slices.Contains(templateResources, exportResourceChores)
		applyRewards := len(templateResources) == 0 || slices.Contains(templateResources, exportResourceRewards)

		if templateDryRun {
			if applyChores {
				for _, c := range tmpl.Chores {
					printDryRun("create chore %q (points: %d)", c.Title, c.Points)
				}
			}
			if applyRewards {
				for _, r := range tmpl.Rewards {
					printDryRun("create reward %q (points: %d)", r.Title, r.Points)
				}
			}
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		var createdChores, createdRewards int

		if applyChores {
			for _, c := range tmpl.Chores {
				dueDate := c.DueDate
				if templateStartDate != "" && dueDate != "" {
					dueDate = templateStartDate
				}
				data := lib.ChoreData{
					Title:       c.Title,
					Description: c.Description,
					Points:      c.Points,
					AssigneeID:  c.AssigneeID,
					DueDate:     dueDate,
					Recurring:   c.Recurring,
				}
				if _, err := client.CreateChore(ctx, frameID, data); err != nil {
					return fmt.Errorf("creating chore %q: %w", c.Title, err)
				}
				createdChores++
			}
		}

		if applyRewards {
			for _, r := range tmpl.Rewards {
				data := lib.RewardData{
					Title:     r.Title,
					Points:    r.Points,
					EmojiIcon: r.EmojiIcon,
				}
				if _, err := client.CreateReward(ctx, frameID, data); err != nil {
					return fmt.Errorf("creating reward %q: %w", r.Title, err)
				}
				createdRewards++
			}
		}

		printSuccessf("Applied template %q: created %d chore(s), %d reward(s)\n", templateName, createdChores, createdRewards)
		return nil
	},
}

var templateListCmd = &cobra.Command{
	Use:   subList,
	Short: "List saved templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := templateDir()
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				printSuccessf("No templates saved yet. Use 'template save --name NAME' to create one.\n")
				return nil
			}
			return fmt.Errorf("reading template directory: %w", err)
		}

		type templateSummary struct {
			Name      string `json:"name"`
			Chores    int    `json:"chores"`
			Rewards   int    `json:"rewards"`
			CreatedAt string `json:"created_at"`
		}

		var summaries []templateSummary
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var tmpl frameTemplate
			if err := json.Unmarshal(raw, &tmpl); err != nil {
				continue
			}
			summaries = append(summaries, templateSummary{
				Name:      tmpl.Name,
				Chores:    len(tmpl.Chores),
				Rewards:   len(tmpl.Rewards),
				CreatedAt: tmpl.CreatedAt,
			})
		}

		printOutput(summaries)
		return nil
	},
}

var templateDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete a saved template",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := templateDir()
		if err != nil {
			return err
		}

		path := filepath.Join(dir, templateName+".json")

		if !confirmAction(fmt.Sprintf("Delete template %q?", templateName)) {
			return nil
		}

		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("template %q not found", templateName)
			}
			return fmt.Errorf("deleting template: %w", err)
		}

		printSuccessf("Template %q deleted\n", templateName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateSaveCmd)
	templateCmd.AddCommand(templateApplyCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateDeleteCmd)

	templateSaveCmd.Flags().StringVar(&templateName, "name", "", "Template name")
	templateSaveCmd.Flags().StringSliceVar(&templateResources, "resources", nil, "Resources to save: chores, rewards (default: both)")
	markFlagRequired(templateSaveCmd, "name")

	templateApplyCmd.Flags().StringVar(&templateName, "name", "", "Template name")
	templateApplyCmd.Flags().StringVar(&templateStartDate, "start-date", "", "Override due date for all chores (YYYY-MM-DD)")
	templateApplyCmd.Flags().StringSliceVar(&templateResources, "resources", nil, "Resources to apply: chores, rewards (default: both)")
	templateApplyCmd.Flags().BoolVar(&templateDryRun, "dry-run", false, "Preview without making API calls")
	markFlagRequired(templateApplyCmd, "name")

	templateDeleteCmd.Flags().StringVar(&templateName, "name", "", "Template name")
	templateDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(templateDeleteCmd, "name")
}
