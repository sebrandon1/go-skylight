package lib

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

// routineLookaheadDays bounds how far ahead ListRoutines looks for
// not-yet-started routines. The chores endpoint requires a date range and
// expands each recurring chore into one row per day, so there's no way to
// ask for "all routines" directly -- a routine starting further out than
// this won't appear until it's within the window.
const routineLookaheadDays = 30

// Routine represents a Skylight routine: a recurring chore with a fixed
// time-of-day slot, assigned to one family member.
//
// NextOccurrenceDate is the date of the nearest active occurrence found
// within ListRoutines's lookahead window (in practice, almost always
// today) -- it is NOT the routine's original creation/start date. The
// chores API has no field that reports that: recur_from is never populated
// for routine (RRULE-based) chores, and each occurrence's own start/due
// date just reflects its own calendar date, not a fixed series anchor.
type Routine struct {
	ID                 string `json:"id,omitempty"`
	Title              string `json:"title,omitempty"`
	TimeOfDay          string `json:"time_of_day,omitempty"`
	AssigneeID         string `json:"assignee_id,omitempty"`
	NextOccurrenceDate string `json:"next_occurrence_date,omitempty"`
}

// RoutineData holds the fields for creating a routine. CategoryID is a
// single assignee: create_multiple fans out one chore record per category
// ID with no server-provided key correlating the resulting siblings back
// together, so multi-assignee routines aren't supported yet -- the type
// only allows one.
type RoutineData struct {
	Title      string `json:"title,omitempty"`
	TimeOfDay  string `json:"time_of_day,omitempty"`
	CategoryID string `json:"category_id,omitempty"`
	StartDate  string `json:"start_date,omitempty"`
}

// routineByHour maps the CLI/API's time-of-day vocabulary to the BYHOUR
// value the Skylight API requires in a routine's recurrence rule. The API
// rejects any other hour: {"errors":{"recurrence_set":["must have exactly
// one BYHOUR, set to 6, 14, or 20"]}}.
var routineByHour = map[string]int{
	"morning":   6,
	"afternoon": 14,
	"evening":   20,
}

var routineTimeOfDay = map[string]string{
	"6":  "morning",
	"14": "afternoon",
	"20": "evening",
}

var byHourRe = regexp.MustCompile(`BYHOUR=(\d+)`)

// timeOfDayFromRecurrence extracts the time-of-day slot from a routine's
// recurrence rules by reading its BYHOUR value.
func timeOfDayFromRecurrence(rules []string) string {
	for _, rule := range rules {
		if m := byHourRe.FindStringSubmatch(rule); m != nil {
			if tod, ok := routineTimeOfDay[m[1]]; ok {
				return tod
			}
		}
	}
	return ""
}

// CreateRoutine creates a routine as a chore via create_multiple, with
// routine:true and the time-of-day slot carried in the recurrence rule's
// BYHOUR. There is no dedicated /routines resource on the Skylight API.
func (c *Client) CreateRoutine(ctx context.Context, frameID string, data RoutineData) (*Routine, error) {
	hour, ok := routineByHour[data.TimeOfDay]
	if !ok {
		return nil, fmt.Errorf("invalid time-of-day %q: must be morning, afternoon, or evening", data.TimeOfDay)
	}
	if data.CategoryID == "" {
		return nil, fmt.Errorf("category ID is required")
	}

	chore := ChoreData{
		Title:         data.Title,
		DueDate:       data.StartDate,
		CategoryIDs:   []string{data.CategoryID},
		RecurrenceSet: []string{fmt.Sprintf("RRULE:FREQ=DAILY;INTERVAL=1;BYHOUR=%d", hour)},
		Routine:       true,
	}

	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/chores/create_multiple", c.effectiveURL(), pathSeg(frameID)), chore)
	if err != nil {
		return nil, fmt.Errorf("failed to create routine request: %w", err)
	}

	var apiResp choreAPIResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create routine: %w", err)
	}
	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("no chore returned from create_multiple")
	}

	ch := apiResp.Data[0].toChore()
	return &Routine{
		ID:                 ch.ID,
		Title:              ch.Title,
		TimeOfDay:          data.TimeOfDay,
		AssigneeID:         ch.AssigneeID,
		NextOccurrenceDate: ch.DueDate,
	}, nil
}

// ListRoutines lists routines active or starting within the next
// routineLookaheadDays. Routines are chores with routine:true; the chores
// endpoint requires a date range and returns one row per day for a
// recurring chore.
func (c *Client) ListRoutines(ctx context.Context, frameID string) ([]Routine, error) {
	now := time.Now()
	today := now.Format(DateFormat)
	until := now.AddDate(0, 0, routineLookaheadDays).Format(DateFormat)

	chores, err := c.ListChores(ctx, frameID, ChoreListOptions{After: today, Before: until})
	if err != nil {
		return nil, fmt.Errorf("failed to list routines: %w", err)
	}

	seen := make(map[string]bool)
	routines := make([]Routine, 0, len(chores))
	for _, ch := range chores {
		if !ch.Routine {
			continue
		}
		baseID, _ := parseChoreID(ch.ID)
		if seen[baseID] {
			continue
		}
		seen[baseID] = true

		routines = append(routines, Routine{
			ID:                 baseID,
			Title:              ch.Title,
			TimeOfDay:          timeOfDayFromRecurrence(ch.RecurrenceSet),
			AssigneeID:         ch.AssigneeID,
			NextOccurrenceDate: ch.DueDate,
		})
	}
	return routines, nil
}

// DeleteRoutine deletes a routine. apply_to=all is required by the API for
// any recurring chore -- without it, the request 400s with "you must have a
// valid value for apply_to".
func (c *Client) DeleteRoutine(ctx context.Context, frameID, routineID string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/chores/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(routineID)))
	if err != nil {
		return fmt.Errorf("failed to create delete routine request: %w", err)
	}
	addQueryParams(req, map[string]string{"apply_to": "all"})

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete routine: %w", err)
	}
	return nil
}
