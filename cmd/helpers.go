package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

const (
	boolYes = "yes"
	boolNo  = "no"
)

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %s: %v\n", msg, err)
	os.Exit(1)
}

func getFrameOrFail(client *lib.Client, id string) *lib.Frame {
	frame, err := client.GetFrame(id)
	if err != nil {
		fatal("getting frame info", err)
	}
	return frame
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
		fatal("marshaling JSON", err)
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
	default:
		return printTableOutputExtended(data)
	}
	return true
}

func printTableOutputExtended(data any) bool {
	switch v := data.(type) {
	case []lib.Bounty:
		printBountiesTable(v)
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
	default:
		return false
	}
	return true
}

func printChoresTable(chores []lib.Chore) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tDUE DATE\tPOINTS\tASSIGNEE")
	for _, c := range chores {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
			c.ID, c.Title, c.Status, c.DueDate, c.Points, c.AssigneeID)
	}
	w.Flush()
}

func printRewardsTable(rewards []lib.Reward) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tPOINTS\tEMOJI\tREDEEMED\tCATEGORY")
	for _, r := range rewards {
		redeemed := boolNo
		if r.Redeemed {
			redeemed = boolYes
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			r.ID, r.Title, r.Points, r.EmojiIcon, redeemed, r.CategoryID)
	}
	w.Flush()
}

func printFramesTable(frames []lib.Frame) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTIMEZONE")
	for _, f := range frames {
		fmt.Fprintf(w, "%s\t%s\t%s\n", f.ID, f.Name, f.TimeZone)
	}
	w.Flush()
}

func printCalendarTable(events []lib.CalendarEvent) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTART\tEND\tALL DAY")
	for _, e := range events {
		allDay := boolNo
		if e.AllDay {
			allDay = boolYes
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			e.ID, e.Title, e.StartAt, e.EndAt, allDay)
	}
	w.Flush()
}

func printCategoriesTable(cats []lib.Category) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCOLOR")
	for _, c := range cats {
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.ID, c.Name, c.Color)
	}
	w.Flush()
}

func printChoreStreakTable(stats []ChoreStreakStats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCURRENT STREAK\tLONGEST STREAK\tCOMPLETED\tTOTAL\tRATE")
	for _, s := range stats {
		fmt.Fprintf(w, "%s\t%d days\t%d days\t%d\t%d\t%.1f%%\n",
			s.AssigneeName, s.CurrentStreak, s.LongestStreak, s.CompletedChores, s.TotalChores, s.CompletionRate)
	}
	w.Flush()
}

func printBountiesTable(bounties []lib.Bounty) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CHORE ID\tCHORE TITLE\tPOINTS\tDUE DATE\tREWARD ID\tREWARD TITLE")
	for _, b := range bounties {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			b.Chore.ID, b.Chore.Title, b.Chore.Points, b.Chore.DueDate, b.Reward.ID, b.Reward.Title)
	}
	w.Flush()
}

func printSourceCalendarsTable(cals []lib.SourceCalendar) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPROVIDER\tCOLOR")
	for _, c := range cals {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ID, c.Name, c.Provider, c.Color)
	}
	w.Flush()
}

func printDevicesTable(devices []lib.Device) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tONLINE")
	for _, d := range devices {
		online := boolNo
		if d.Online {
			online = boolYes
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.ID, d.Name, online)
	}
	w.Flush()
}

func printAvatarsTable(avatars []lib.Avatar) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tIMAGE URL")
	for _, a := range avatars {
		fmt.Fprintf(w, "%s\t%s\t%s\n", a.ID, a.Name, a.ImageURL)
	}
	w.Flush()
}

func printColorsTable(colors []lib.Color) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tHEX")
	for _, c := range colors {
		fmt.Fprintf(w, "%s\t%s\n", c.Name, c.Hex)
	}
	w.Flush()
}

func printListsTable(lists []lib.List) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tCOLOR\tKIND\tITEMS")
	for _, l := range lists {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", l.ID, l.Title, l.Color, l.Kind, len(l.Items))
	}
	w.Flush()
}

func printMealCategoriesTable(cats []lib.MealCategory) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCOLOR")
	for _, c := range cats {
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.ID, c.Name, c.Color)
	}
	w.Flush()
}

func printRecipesTable(recipes []lib.Recipe) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tURL")
	for _, r := range recipes {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.ID, r.Title, r.MealCategoryID, r.URL)
	}
	w.Flush()
}

func printMealSittingsTable(sittings []lib.MealSitting) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSUMMARY\tDATE\tRECIPE ID\tCATEGORY ID")
	for _, s := range sittings {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.ID, s.Summary, s.Date, s.RecipeID, s.MealCategoryID)
	}
	w.Flush()
}

func printPhotosTable(photos []lib.Photo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tSTATUS\tCREATED")
	for _, p := range photos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.AssetType, p.Status, p.CreatedAt)
	}
	w.Flush()
}

func printRoutinesTable(routines []lib.Routine) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tASSIGNEE\tSTEPS")
	for _, r := range routines {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", r.ID, r.Title, r.AssigneeID, len(r.Steps))
	}
	w.Flush()
}
