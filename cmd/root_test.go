package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSetVersion(t *testing.T) {
	SetVersion("1.2.3")
	if rootCmd.Version != "1.2.3" {
		t.Errorf("rootCmd.Version = %q, want %q", rootCmd.Version, "1.2.3")
	}
	// Reset for other tests
	SetVersion("dev")
}

func TestExecuteHelp(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	if err := Execute(); err != nil {
		t.Errorf("Execute --help failed: %v", err)
	}
}

func TestRootCommandExists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	if rootCmd.Use != "skylight" {
		t.Errorf("Expected Use 'skylight', got '%s'", rootCmd.Use)
	}
}

func TestRootCommandShort(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
}

func TestGetCommandExists(t *testing.T) {
	if getCmd == nil {
		t.Fatal("getCmd should not be nil")
	}
	if getCmd.Use != "get" {
		t.Errorf("Expected Use 'get', got '%s'", getCmd.Use)
	}
}

func TestRootPersistentFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	tests := []struct {
		name string
		flag string
	}{
		{"email flag", "email"},
		{"password flag", "password"},
		{"token flag", "token"},
		{"user-id flag", "user-id"},
		{"frame-id flag", "frame-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' to be registered", tt.flag)
			}
		})
	}
}

func TestRootSubcommands(t *testing.T) {
	subcommands := rootCmd.Commands()

	foundGet := false
	foundLogin := false
	for _, cmd := range subcommands {
		if cmd.Use == "get" {
			foundGet = true
		}
		if cmd.Use == "login" {
			foundLogin = true
		}
	}

	if !foundGet {
		t.Error("Expected 'get' subcommand under root")
	}
	if !foundLogin {
		t.Error("Expected 'login' subcommand under root")
	}
}

func TestLoginCommandExists(t *testing.T) {
	if loginCmd == nil {
		t.Fatal("loginCmd should not be nil")
	}
	if loginCmd.Use != "login" {
		t.Errorf("Expected Use 'login', got '%s'", loginCmd.Use)
	}
	if loginCmd.Run == nil {
		t.Error("loginCmd should have a Run function")
	}
}

func TestExecuteReturnsNoError(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute with --help failed: %v", err)
	}
}

// Calendar subcommand tests

func TestCalendarCommandExists(t *testing.T) {
	if calendarCmd == nil {
		t.Fatal("calendarCmd should not be nil")
	}
	if calendarCmd.Use != "calendar" {
		t.Errorf("Expected Use 'calendar', got '%s'", calendarCmd.Use)
	}
}

func TestCalendarSubcommands(t *testing.T) {
	subcommands := calendarCmd.Commands()

	expected := map[string]bool{
		"list":    false,
		"create":  false,
		"delete":  false,
		"sources": false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under calendar", name)
		}
	}
}

func TestCalendarListFlags(t *testing.T) {
	flags := calendarListCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"start-date flag", "start-date"},
		{"end-date flag", "end-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on calendar list command", tt.flag)
			}
		})
	}
}

func TestCalendarCreateFlags(t *testing.T) {
	flags := calendarCreateCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"title flag", "title"},
		{"start-at flag", "start-at"},
		{"end-at flag", "end-at"},
		{"all-day flag", "all-day"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on calendar create command", tt.flag)
			}
		})
	}
}

func TestCalendarDeleteFlags(t *testing.T) {
	f := calendarDeleteCmd.Flags().Lookup("event-id")
	if f == nil {
		t.Error("Expected flag 'event-id' on calendar delete command")
	}
}

// Chore subcommand tests

func TestChoreCommandExists(t *testing.T) {
	if choreCmd == nil {
		t.Fatal("choreCmd should not be nil")
	}
	if choreCmd.Use != "chore" {
		t.Errorf("Expected Use 'chore', got '%s'", choreCmd.Use)
	}
}

func TestChoreSubcommands(t *testing.T) {
	subcommands := choreCmd.Commands()

	expected := map[string]bool{
		"list":     false,
		"create":   false,
		"delete":   false,
		"complete": false,
		"skip":     false,
		"claim":    false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under chore", name)
		}
	}
}

func TestChoreListFlags(t *testing.T) {
	flags := choreListCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"date flag", "date"},
		{"status flag", "status"},
		{"assignee-id flag", "assignee-id"},
		{"after flag", "after"},
		{"before flag", "before"},
		{"include-late flag", "include-late"},
		{"up-for-grabs flag", "up-for-grabs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on chore list command", tt.flag)
			}
		})
	}
}

func TestChoreCreateFlags(t *testing.T) {
	flags := choreCreateCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"title flag", "title"},
		{"date flag", "date"},
		{"assignee-id flag", "assignee-id"},
		{"points flag", "points"},
		{"recurring flag", "recurring"},
		{"up-for-grabs flag", "up-for-grabs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on chore create command", tt.flag)
			}
		})
	}
}

func TestChoreDeleteFlags(t *testing.T) {
	f := choreDeleteCmd.Flags().Lookup("chore-id")
	if f == nil {
		t.Error("Expected flag 'chore-id' on chore delete command")
	}
}

func TestChoreCompleteFlags(t *testing.T) {
	f := choreCompleteCmd.Flags().Lookup("chore-id")
	if f == nil {
		t.Error("Expected flag 'chore-id' on chore complete command")
	}
}

func TestChoreSkipFlags(t *testing.T) {
	f := choreSkipCmd.Flags().Lookup("chore-id")
	if f == nil {
		t.Error("Expected flag 'chore-id' on chore skip command")
	}
}

func TestChoreClaimFlags(t *testing.T) {
	flags := choreClaimCmd.Flags()
	for _, name := range []string{"chore-id", "assignee-id"} {
		if flags.Lookup(name) == nil {
			t.Errorf("Expected flag %q on chore claim command", name)
		}
	}
}

// List subcommand tests

func TestListCommandExists(t *testing.T) {
	if listCmd == nil {
		t.Fatal("listCmd should not be nil")
	}
	if listCmd.Use != "list" {
		t.Errorf("Expected Use 'list', got '%s'", listCmd.Use)
	}
}

func TestListSubcommands(t *testing.T) {
	subcommands := listCmd.Commands()

	expected := map[string]bool{
		"all":         false,
		"info":        false,
		"create":      false,
		"delete":      false,
		"add-item":    false,
		"delete-item": false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under list", name)
		}
	}
}

func TestListGetFlags(t *testing.T) {
	f := listGetCmd.Flags().Lookup("list-id")
	if f == nil {
		t.Error("Expected flag 'list-id' on list info command")
	}
}

func TestListCreateFlags(t *testing.T) {
	flags := listCreateCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"title flag", "title"},
		{"color flag", "color"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on list create command", tt.flag)
			}
		})
	}
}

func TestListDeleteFlags(t *testing.T) {
	f := listDeleteCmd.Flags().Lookup("list-id")
	if f == nil {
		t.Error("Expected flag 'list-id' on list delete command")
	}
}

func TestListAddItemFlags(t *testing.T) {
	flags := listAddItemCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"list-id flag", "list-id"},
		{"title flag", "title"},
		{"position flag", "position"}, // #274
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on list add-item command", tt.flag)
			}
		})
	}
}

func TestListUpdateItemFlags(t *testing.T) {
	// #274: update-item exposes --position
	if listUpdateItemCmd.Flags().Lookup("position") == nil {
		t.Error("Expected flag 'position' on list update-item command")
	}
}

func TestListDeleteItemFlags(t *testing.T) {
	flags := listDeleteItemCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"list-id flag", "list-id"},
		{"item-id flag", "item-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on list delete-item command", tt.flag)
			}
		})
	}
}

// Reward subcommand tests

func TestRewardCommandExists(t *testing.T) {
	if rewardCmd == nil {
		t.Fatal("rewardCmd should not be nil")
	}
	if rewardCmd.Use != "reward" {
		t.Errorf("Expected Use 'reward', got '%s'", rewardCmd.Use)
	}
}

func TestRewardSubcommands(t *testing.T) {
	subcommands := rewardCmd.Commands()

	expected := map[string]bool{
		"list":     false,
		"create":   false,
		"delete":   false,
		"redeem":   false,
		"unredeem": false,
		"points":   false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under reward", name)
		}
	}
}

func TestRewardCreateFlags(t *testing.T) {
	flags := rewardCreateCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"title flag", "title"},
		{"points flag", "points"},
		{"emoji-icon flag", "emoji-icon"},
		{"no-respawn flag", "no-respawn"},
		{"category-ids flag", "category-ids"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on reward create command", tt.flag)
			}
		})
	}
}

func TestRewardUpdateFlags(t *testing.T) {
	flags := rewardUpdateCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"reward-id flag", "reward-id"},
		{"title flag", "title"},
		{"points flag", "points"},
		{"emoji-icon flag", "emoji-icon"},
		{"no-respawn flag", "no-respawn"},
		{"category-ids flag", "category-ids"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on reward update command", tt.flag)
			}
		})
	}
}

func TestRewardDeleteFlags(t *testing.T) {
	f := rewardDeleteCmd.Flags().Lookup("reward-id")
	if f == nil {
		t.Error("Expected flag 'reward-id' on reward delete command")
	}
}

func TestRewardRedeemFlags(t *testing.T) {
	f := rewardRedeemCmd.Flags().Lookup("reward-id")
	if f == nil {
		t.Error("Expected flag 'reward-id' on reward redeem command")
	}
}

func TestRewardUnredeemFlags(t *testing.T) {
	f := rewardUnredeemCmd.Flags().Lookup("reward-id")
	if f == nil {
		t.Error("Expected flag 'reward-id' on reward unredeem command")
	}
}

// Meal subcommand tests

func TestMealCommandExists(t *testing.T) {
	if mealCmd == nil {
		t.Fatal("mealCmd should not be nil")
	}
	if mealCmd.Use != "meal" {
		t.Errorf("Expected Use 'meal', got '%s'", mealCmd.Use)
	}
}

func TestMealSubcommands(t *testing.T) {
	subcommands := mealCmd.Commands()

	expected := map[string]bool{
		"categories":     false,
		"recipes":        false,
		"recipe-info":    false,
		"create-recipe":  false,
		"delete-recipe":  false,
		"sittings":       false,
		"create-sitting": false,
		"add-to-grocery": false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under meal", name)
		}
	}
}

func TestMealRecipeInfoFlags(t *testing.T) {
	f := mealRecipeInfoCmd.Flags().Lookup("recipe-id")
	if f == nil {
		t.Error("Expected flag 'recipe-id' on meal recipe-info command")
	}
}

func TestMealCreateRecipeFlags(t *testing.T) {
	f := mealCreateRecipeCmd.Flags().Lookup("title")
	if f == nil {
		t.Error("Expected flag 'title' on meal create-recipe command")
	}
}

func TestMealDeleteRecipeFlags(t *testing.T) {
	f := mealDeleteRecipeCmd.Flags().Lookup("recipe-id")
	if f == nil {
		t.Error("Expected flag 'recipe-id' on meal delete-recipe command")
	}
}

func TestMealCreateSittingFlags(t *testing.T) {
	flags := mealCreateSittingCmd.Flags()

	tests := []struct {
		name string
		flag string
	}{
		{"recipe-id flag", "recipe-id"},
		{"date flag", "date"},
		{"meal-category-id flag", "meal-category-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' on meal create-sitting command", tt.flag)
			}
		})
	}
}

func TestMealAddToGroceryFlags(t *testing.T) {
	f := mealAddToGroceryCmd.Flags().Lookup("recipe-id")
	if f == nil {
		t.Error("Expected flag 'recipe-id' on meal add-to-grocery command")
	}
}

// Category command tests

func TestCategoryCommandExists(t *testing.T) {
	if categoryCmd == nil {
		t.Fatal("categoryCmd should not be nil")
	}
	if categoryCmd.Use != "category" {
		t.Errorf("Expected Use 'category', got '%s'", categoryCmd.Use)
	}
	if !categoryCmd.HasSubCommands() {
		t.Error("categoryCmd should have subcommands")
	}
}

// Frame subcommand tests

func TestFrameCommandExists(t *testing.T) {
	if frameCmd == nil {
		t.Fatal("frameCmd should not be nil")
	}
	if frameCmd.Use != "frame" {
		t.Errorf("Expected Use 'frame', got '%s'", frameCmd.Use)
	}
}

func TestFrameSubcommands(t *testing.T) {
	subcommands := frameCmd.Commands()

	expected := map[string]bool{
		"info":    false,
		"devices": false,
		"avatars": false,
		"colors":  false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under frame", name)
		}
	}
}

// Verify all commands have Short descriptions

func TestAllCommandsHaveShortDescriptions(t *testing.T) {
	cmds := map[string]*cobra.Command{
		"root":                rootCmd,
		"get":                 getCmd,
		"login":               loginCmd,
		"calendar":            calendarCmd,
		"calendar list":       calendarListCmd,
		"calendar create":     calendarCreateCmd,
		"calendar delete":     calendarDeleteCmd,
		"calendar sources":    sourceCalendarsCmd,
		"chore":               choreCmd,
		"chore list":          choreListCmd,
		"chore create":        choreCreateCmd,
		"chore delete":        choreDeleteCmd,
		"chore complete":      choreCompleteCmd,
		"list":                listCmd,
		"list all":            listListCmd,
		"list info":           listGetCmd,
		"list create":         listCreateCmd,
		"list delete":         listDeleteCmd,
		"list add-item":       listAddItemCmd,
		"list delete-item":    listDeleteItemCmd,
		"reward":              rewardCmd,
		"reward list":         rewardListCmd,
		"reward create":       rewardCreateCmd,
		"reward delete":       rewardDeleteCmd,
		"reward redeem":       rewardRedeemCmd,
		"reward unredeem":     rewardUnredeemCmd,
		"reward points":       rewardPointsCmd,
		"meal":                mealCmd,
		"meal categories":     mealCategoriesCmd,
		"meal recipes":        mealRecipesCmd,
		"meal recipe-info":    mealRecipeInfoCmd,
		"meal create-recipe":  mealCreateRecipeCmd,
		"meal delete-recipe":  mealDeleteRecipeCmd,
		"meal sittings":       mealSittingsCmd,
		"meal create-sitting": mealCreateSittingCmd,
		"meal get-sitting":    mealGetSittingCmd,
		"meal add-to-grocery": mealAddToGroceryCmd,
		"category":            categoryCmd,
		"category list":       categoryListCmd,
		"category create":     categoryCreateCmd,
		"category delete":     categoryDeleteCmd,
		"category update":     categoryUpdateCmd,
		"frame":               frameCmd,
		"frame info":          frameInfoCmd,
		"frame devices":       frameDevicesCmd,
		"frame avatars":       frameAvatarsCmd,
		"frame colors":        frameColorsCmd,
		"config":              configCmd,
		"config show":         configShowCmd,
		"config get":          configGetCmd,
		"config set":          configSetCmd,
	}

	for name, c := range cmds {
		t.Run(name, func(t *testing.T) {
			if c.Short == "" {
				t.Errorf("Command '%s' should have a Short description", name)
			}
		})
	}
}

// Verify leaf commands have Run functions

func TestLeafCommandsHaveRunFunctions(t *testing.T) {
	leafCmds := map[string]*cobra.Command{
		"login":               loginCmd,
		"calendar list":       calendarListCmd,
		"calendar create":     calendarCreateCmd,
		"calendar delete":     calendarDeleteCmd,
		"calendar sources":    sourceCalendarsCmd,
		"chore list":          choreListCmd,
		"chore create":        choreCreateCmd,
		"chore delete":        choreDeleteCmd,
		"chore complete":      choreCompleteCmd,
		"list all":            listListCmd,
		"list info":           listGetCmd,
		"list create":         listCreateCmd,
		"list delete":         listDeleteCmd,
		"list add-item":       listAddItemCmd,
		"list delete-item":    listDeleteItemCmd,
		"reward list":         rewardListCmd,
		"reward create":       rewardCreateCmd,
		"reward delete":       rewardDeleteCmd,
		"reward redeem":       rewardRedeemCmd,
		"reward unredeem":     rewardUnredeemCmd,
		"reward points":       rewardPointsCmd,
		"meal categories":     mealCategoriesCmd,
		"meal recipes":        mealRecipesCmd,
		"meal recipe-info":    mealRecipeInfoCmd,
		"meal create-recipe":  mealCreateRecipeCmd,
		"meal delete-recipe":  mealDeleteRecipeCmd,
		"meal sittings":       mealSittingsCmd,
		"meal create-sitting": mealCreateSittingCmd,
		"meal get-sitting":    mealGetSittingCmd,
		"meal add-to-grocery": mealAddToGroceryCmd,
		"category list":       categoryListCmd,
		"category create":     categoryCreateCmd,
		"category delete":     categoryDeleteCmd,
		"category update":     categoryUpdateCmd,
		"frame info":          frameInfoCmd,
		"frame devices":       frameDevicesCmd,
		"frame avatars":       frameAvatarsCmd,
		"frame colors":        frameColorsCmd,
		"config show":         configShowCmd,
		"config get":          configGetCmd,
		"config set":          configSetCmd,
	}

	for name, c := range leafCmds {
		t.Run(name, func(t *testing.T) {
			if c.Run == nil && c.RunE == nil {
				t.Errorf("Leaf command '%s' should have a Run or RunE function", name)
			}
		})
	}
}

// Verify parent commands do not have Run functions

func TestParentCommandsHaveNoRunFunction(t *testing.T) {
	parentCmds := map[string]*cobra.Command{
		"root":     rootCmd,
		"get":      getCmd,
		"calendar": calendarCmd,
		"chore":    choreCmd,
		"list":     listCmd,
		"reward":   rewardCmd,
		"meal":     mealCmd,
		"frame":    frameCmd,
		"config":   configCmd,
	}

	for name, c := range parentCmds {
		t.Run(name, func(t *testing.T) {
			if c.Run != nil {
				t.Errorf("Parent command '%s' should not have a Run function", name)
			}
		})
	}
}

func TestGetCommandHelpDoesNotError(t *testing.T) {
	rootCmd.SetArgs([]string{"get", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute 'get --help' failed: %v", err)
	}
}

func TestCalendarHelpDoesNotError(t *testing.T) {
	rootCmd.SetArgs([]string{"get", "calendar", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute 'get calendar --help' failed: %v", err)
	}
}

// TestResourceCommandsAtRootLevel verifies that all resource commands are
// registered directly under rootCmd, not only nested under get.
func TestResourceCommandsAtRootLevel(t *testing.T) {
	expected := map[string]bool{
		"calendar": false,
		"chore":    false,
		"list":     false,
		"reward":   false,
		"meal":     false,
		"category": false,
		"frame":    false,
	}
	for _, cmd := range rootCmd.Commands() {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected '%s' as a top-level command under root", name)
		}
	}
}

// TestResourceCommandPaths guards against the Cobra dual-parent bug where
// adding the same *cobra.Command instance to two parents causes CommandPath()
// to reflect the wrong ancestor (e.g. "skylight get chore" instead of "skylight chore").
func TestResourceCommandPaths(t *testing.T) {
	tests := []struct {
		wantPath string
		cmd      *cobra.Command
	}{
		{"skylight calendar", calendarCmd},
		{"skylight chore", choreCmd},
		{"skylight list", listCmd},
		{"skylight reward", rewardCmd},
		{"skylight meal", mealCmd},
		{"skylight category", categoryCmd},
		{"skylight frame", frameCmd},
	}
	for _, tt := range tests {
		t.Run(tt.wantPath, func(t *testing.T) {
			if got := tt.cmd.CommandPath(); got != tt.wantPath {
				t.Errorf("CommandPath() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

// TestGetCommandIsHidden verifies that get is hidden now that resource commands
// live at the top level.
func TestGetCommandIsHidden(t *testing.T) {
	if !getCmd.Hidden {
		t.Error("getCmd should be hidden after flattening resource commands to top level")
	}
}

func TestConfigCommandExists(t *testing.T) {
	if configCmd == nil {
		t.Fatal("configCmd should not be nil")
	}
	if configCmd.Use != "config" {
		t.Errorf("Expected Use 'config', got '%s'", configCmd.Use)
	}

	expected := map[string]bool{"show": false, "get": false, "set": false}
	for _, sub := range configCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under config", name)
		}
	}
}
