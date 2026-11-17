package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-skylight/lib"
)

// printAllTasksTable renders the All Tasks view grouped by kind and bucket.
func printAllTasksTable(t *lib.AllTasks) {
	w := newTableWriter()
	fmt.Fprintln(w, "KIND\tBUCKET\tID\tTITLE\tDUE DATE\tTIME\tPOINTS\tASSIGNEE")
	for _, kind := range []struct {
		name string
		b    lib.TaskBuckets
	}{{"chore", t.Chores}, {"routine", t.Routines}} {
		for _, bucket := range []struct {
			name  string
			tasks []lib.Chore
		}{
			{"late", kind.b.Late}, {"today", kind.b.Today}, {"today_timed", kind.b.TodayTimed},
			{"any_day", kind.b.AnyDay}, {"future", kind.b.Future},
		} {
			for _, c := range bucket.tasks {
				assignee := "-"
				if c.UpForGrabs {
					assignee = "up for grabs"
				} else if c.AssigneeID != "" {
					assignee = resolveCatName(c.AssigneeID)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
					kind.name, bucket.name, c.ID, c.Title, dashIfEmpty(c.DueDate), dashIfEmpty(c.StartTime), c.Points, assignee)
			}
		}
	}
	w.Flush()
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
