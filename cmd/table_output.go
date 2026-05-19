package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
)

func printChoresTable(chores []lib.Chore) {
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tDUE DATE\tPOINTS\tASSIGNEE")
	for _, c := range chores {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
			c.ID, c.Title, c.Status, c.DueDate, c.Points, c.AssigneeID)
	}
	w.Flush()
}

func printRewardsTable(rewards []lib.Reward) {
	w := newTableWriter()
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
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tNAME\tTIMEZONE")
	for _, f := range frames {
		fmt.Fprintf(w, "%s\t%s\t%s\n", f.ID, f.Name, f.TimeZone)
	}
	w.Flush()
}

func printCalendarTable(events []lib.CalendarEvent) {
	w := newTableWriter()
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
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tNAME\tCOLOR")
	for _, c := range cats {
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.ID, c.Name, c.Color)
	}
	w.Flush()
}

func printChoreStreakTable(stats []ChoreStreakStats) {
	w := newTableWriter()
	fmt.Fprintln(w, "NAME\tCURRENT STREAK\tLONGEST STREAK\tCOMPLETED\tTOTAL\tRATE")
	for _, s := range stats {
		fmt.Fprintf(w, "%s\t%d days\t%d days\t%d\t%d\t%.1f%%\n",
			s.AssigneeName, s.CurrentStreak, s.LongestStreak, s.CompletedChores, s.TotalChores, s.CompletionRate)
	}
	w.Flush()
}

func printChoreWeekTable(days []WeeklyChoreDay) {
	w := newTableWriter()
	fmt.Fprintln(w, "DAY\tDATE\tTITLE\tSTATUS")
	for _, d := range days {
		if len(d.Chores) == 0 {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.Day, d.Display, "(no chores)", "—")
			continue
		}
		for i, c := range d.Chores {
			dayCol, dateCol := d.Day, d.Display
			if i > 0 {
				dayCol, dateCol = "", ""
			}
			status := "✗"
			if c.Status == lib.ChoreStatusComplete {
				status = "✓"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", dayCol, dateCol, c.Title, status)
		}
	}
	w.Flush()
}

func printCalendarWeekTable(days []WeeklyCalendarDay) {
	w := newTableWriter()
	fmt.Fprintln(w, "DAY\tDATE\tTITLE\tTIME\tALL DAY")
	for _, d := range days {
		if len(d.Events) == 0 {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.Day, d.Display, "(no events)", "—", "—")
			continue
		}
		for i, e := range d.Events {
			dayCol, dateCol := d.Day, d.Display
			if i > 0 {
				dayCol, dateCol = "", ""
			}
			timeCol := "All day"
			if !e.AllDay {
				timeCol = "—"
				if len(e.StartAt) >= 16 {
					timeCol = e.StartAt[11:16]
				}
			}
			allDayCol := boolNo
			if e.AllDay {
				allDayCol = boolYes
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", dayCol, dateCol, e.Title, timeCol, allDayCol)
		}
	}
	w.Flush()
}

func printBountiesTable(bounties []lib.Bounty) {
	w := newTableWriter()
	fmt.Fprintln(w, "CHORE ID\tCHORE TITLE\tPOINTS\tDUE DATE\tREWARD ID\tREWARD TITLE")
	for _, b := range bounties {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			b.Chore.ID, b.Chore.Title, b.Chore.Points, b.Chore.DueDate, b.Reward.ID, b.Reward.Title)
	}
	w.Flush()
}

func printSourceCalendarsTable(cals []lib.SourceCalendar) {
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tNAME\tPROVIDER\tCOLOR")
	for _, c := range cals {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ID, c.Name, c.Provider, c.Color)
	}
	w.Flush()
}

func printDevicesTable(devices []lib.Device) {
	w := newTableWriter()
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
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tNAME\tIMAGE URL")
	for _, a := range avatars {
		fmt.Fprintf(w, "%s\t%s\t%s\n", a.ID, a.Name, a.ImageURL)
	}
	w.Flush()
}

func printColorsTable(colors []lib.Color) {
	w := newTableWriter()
	fmt.Fprintln(w, "NAME\tHEX")
	for _, c := range colors {
		fmt.Fprintf(w, "%s\t%s\n", c.Name, c.Hex)
	}
	w.Flush()
}

func printListsTable(lists []lib.List) {
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tTITLE\tCOLOR\tKIND\tITEMS")
	for _, l := range lists {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", l.ID, l.Title, l.Color, l.Kind, len(l.Items))
	}
	w.Flush()
}

func printMealCategoriesTable(cats []lib.MealCategory) {
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tNAME\tCOLOR")
	for _, c := range cats {
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.ID, c.Name, c.Color)
	}
	w.Flush()
}

func printRecipesTable(recipes []lib.Recipe) {
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tURL")
	for _, r := range recipes {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.ID, r.Title, r.MealCategoryID, r.URL)
	}
	w.Flush()
}

func printMealSittingsTable(sittings []lib.MealSitting) {
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tSUMMARY\tDATE\tRECIPE ID\tCATEGORY ID")
	for _, s := range sittings {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.ID, s.Summary, s.Date, s.RecipeID, s.MealCategoryID)
	}
	w.Flush()
}

func printPhotosTable(photos []lib.Photo) {
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tTYPE\tSTATUS\tCREATED")
	for _, p := range photos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.AssetType, p.Status, p.CreatedAt)
	}
	w.Flush()
}

func printRoutinesTable(routines []lib.Routine) {
	w := newTableWriter()
	fmt.Fprintln(w, "ID\tTITLE\tASSIGNEE\tSTEPS")
	for _, r := range routines {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", r.ID, r.Title, r.AssigneeID, len(r.Steps))
	}
	w.Flush()
}
