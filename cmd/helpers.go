package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

const (
	boolYes = "yes"
	boolNo  = "no"
)

// dryRun and yes are shared by all destructive commands. Only one command
// runs per invocation, so sharing a backing var across multiple flag
// registrations is safe.
var (
	dryRun bool
	yes    bool
)

// confirmAction prompts on stderr to keep stdout pipeable.
func confirmAction(prompt string) bool {
	if yes {
		return true
	}
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	r := strings.ToLower(sc.Text())
	return r == "y" || r == boolYes
}

// activeCatNames holds the category ID → name map used by table renderers to
// resolve raw IDs to human-readable names. Set by list commands that support
// table output before calling printOutput.
var activeCatNames map[string]string

// loadCatNames fetches category names for table rendering. Errors are silently
// ignored so that a failed category lookup never blocks the primary output.
func loadCatNames(ctx context.Context, client *lib.Client) {
	cats, err := client.ListCategories(ctx, frameID)
	if err != nil {
		return
	}
	activeCatNames = buildCatNames(cats)
}

// maybeLoadCatNames populates activeCatNames only when table output is active,
// so JSON-mode invocations incur no extra API call.
func maybeLoadCatNames(ctx context.Context, client *lib.Client) {
	if outputFormat == outputTable {
		loadCatNames(ctx, client)
	}
}

// printSuccess prints a human-readable confirmation message, unless --quiet
// was set, in which case it's suppressed. Use for post-mutation confirmations
// (e.g. "X deleted successfully") — never for primary command output.
func printSuccess(msg string) {
	if quiet {
		return
	}
	fmt.Println(msg)
}

// printSuccessf is printSuccess with fmt.Printf-style formatting, for
// confirmations that carry dynamic content (e.g. an item count).
func printSuccessf(format string, args ...any) {
	if quiet {
		return
	}
	fmt.Printf(format, args...)
}

// printDryRun prints a "Dry run: would ..." preview for --dry-run commands.
// Deliberately not gated by --quiet: it's the only output a dry-run
// invocation produces, so suppressing it would leave no signal that
// anything was evaluated at all — unlike printSuccess/printSuccessf, which
// gate confirmations layered on top of output that already happened.
func printDryRun(format string, args ...any) {
	fmt.Printf("Dry run: would "+format+"\n", args...)
}

// markFlagRequired marks a flag as required, panicking if the flag name
// doesn't exist on the command. A failure here is a programmer error
// (typo'd flag name) that should be caught immediately rather than
// silently leaving the flag optional.
func markFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Sprintf("marking flag %q required on %q: %v", name, cmd.CommandPath(), err))
	}
}

func getFrameOrFail(ctx context.Context, client *lib.Client, id string) (*lib.Frame, error) {
	frame, err := client.GetFrame(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting frame info: %w", err)
	}
	return frame, nil
}

// validateDate returns an error if date is non-empty and not in YYYY-MM-DD format.
func validateDate(date string) error {
	if date == "" {
		return nil
	}
	if _, err := time.Parse(lib.DateFormat, date); err != nil {
		return fmt.Errorf("invalid date %q: use YYYY-MM-DD format", date)
	}
	return nil
}

// validateEnum returns an error if value is non-empty and not present in allowed.
func validateEnum(value string, allowed []string) error {
	if value == "" || slices.Contains(allowed, value) {
		return nil
	}
	return fmt.Errorf("invalid value %q: must be one of %s", value, strings.Join(allowed, ", "))
}

// pointEntry is a reward point balance with its category ID resolved to a
// human-readable name.
type pointEntry struct {
	Name    string `json:"name"`
	Balance int    `json:"balance"`
}

// resolveRewardPointNames maps each point entry's numeric CategoryID to a
// human-readable name via categories, falling back to the numeric ID as a
// string when no matching category is found.
func resolveRewardPointNames(points []lib.RewardPointEntry, categories []lib.Category) []pointEntry {
	catNames := buildCatNames(categories)

	entries := make([]pointEntry, 0, len(points))
	for _, p := range points {
		key := strconv.Itoa(p.CategoryID)
		name := catNames[key]
		if name == "" {
			name = key
		}
		entries = append(entries, pointEntry{Name: name, Balance: p.CurrentPointBalance})
	}
	return entries
}

func resolveCatName(id string) string {
	if id == "" {
		return ""
	}
	if n := activeCatNames[id]; n != "" {
		return n
	}
	return id
}

func buildCatNames(categories []lib.Category) map[string]string {
	m := make(map[string]string, len(categories))
	for _, c := range categories {
		m[c.ID] = c.Name
	}
	return m
}

func printJSON(data any) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("marshaling JSON: %v", err))
	}
	fmt.Println(string(output))
}

// printOutput prints data in the format specified by --output (json or table).
// Defaults to JSON for types that don't have a dedicated table renderer.
func printOutput(data any) {
	if outputFormat == outputTable && printTableOutput(data) {
		return
	}
	printJSON(data)
}

func printTableOutput(data any) bool {
	switch v := data.(type) {
	case []lib.Chore:
		printChoresTable(v)
	case []lib.Reward:
		printRewardsTable(v)
	case []lib.Frame:
		printFramesTable(v)
	case []lib.CalendarEvent:
		printCalendarTable(v)
	case []lib.Category:
		printCategoriesTable(v)
	case []ChoreStreakStats:
		printChoreStreakTable(v)
	case []WeeklyChoreDay:
		printChoreWeekTable(v)
	case []WeeklyCalendarDay:
		printCalendarWeekTable(v)
	case []lib.Bounty:
		printBountiesTable(v)
	default:
		return printTableOutputMore(data)
	}
	return true
}

func printTableOutputMore(data any) bool {
	switch v := data.(type) {
	case []lib.SourceCalendar:
		printSourceCalendarsTable(v)
	case []lib.Device:
		printDevicesTable(v)
	case []lib.Avatar:
		printAvatarsTable(v)
	case []lib.Color:
		printColorsTable(v)
	case []lib.List:
		printListsTable(v)
	case []lib.MealCategory:
		printMealCategoriesTable(v)
	case []lib.Recipe:
		printRecipesTable(v)
	case []lib.MealSitting:
		printMealSittingsTable(v)
	case []lib.Photo:
		printPhotosTable(v)
	case []lib.Routine:
		printRoutinesTable(v)
	case []pointEntry:
		printPointsTable(v)
	case []lib.ListItem:
		printListItemsTable(v)
	case map[string]lib.FeatureState:
		printFeatureBundleTable(v)
	default:
		return false
	}
	return true
}

func newTableWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

func filterEmptyStrings(ss []string) []string {
	out := ss[:0:0]
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
