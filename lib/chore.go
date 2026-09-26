package lib

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	ChoreStatusComplete = "complete"
	ChoreStatusPending  = "pending"
	ChoreStatusSkipped  = "skipped"

	choreStatusPending = ChoreStatusPending
	paramTrue          = "true"
	applyToAll         = "all"
	queryKeyApplyTo    = "apply_to"
)

// choreIDRe matches composite chore IDs like "18731133-2026-04-28" or "70190003-2026-04-28-0600".
// The trailing HHMM group is present for routine (time-of-day) occurrences.
var choreIDRe = regexp.MustCompile(`^(\d+)-(\d{4}-\d{2}-\d{2})(?:-(\d{4}))?`)

// parseChoreID splits a composite chore ID into its base numeric ID, instance date,
// and optional instance time in HH:MM format (e.g. "06:00"). The ID encodes time
// as HHMM (e.g. "0600"); we convert to HH:MM because that is what the completions
// endpoint requires. For plain numeric IDs, all but baseID are empty.
func parseChoreID(choreID string) (baseID, instanceDate, instanceTime string) {
	if m := choreIDRe.FindStringSubmatch(choreID); m != nil {
		hhmm := m[3]
		if len(hhmm) == 4 {
			hhmm = hhmm[:2] + ":" + hhmm[2:]
		}
		return m[1], m[2], hhmm
	}
	return choreID, "", ""
}

// setCompletion calls the chore completions endpoint with the given data.
// It fills in InstanceDate and InstanceTime from the choreID before sending.
func (c *Client) setCompletion(ctx context.Context, frameID, choreID string, data ChoreCompletionData) error {
	baseID, instanceDate, instanceTime := parseChoreID(choreID)
	data.InstanceDate = instanceDate
	data.InstanceTime = instanceTime
	req, err := newRequestWithBody(ctx, "PUT",
		fmt.Sprintf("%s/frames/%s/chores/%s/completions", c.effectiveURL(), pathSeg(frameID), pathSeg(baseID)), data)
	if err != nil {
		return fmt.Errorf("failed to create completion request: %w", err)
	}
	var result choreAPISingleResponse
	return c.put(req, &result)
}

func (opts ChoreListOptions) queryParams() map[string]string {
	params := map[string]string{}
	if opts.Date != "" {
		params["date"] = opts.Date
	}
	if opts.Status != "" {
		params["status"] = opts.Status
	}
	if opts.AssigneeID != "" {
		params["assignee_id"] = opts.AssigneeID
	}
	if opts.IncludeLate {
		params["include_late"] = paramTrue
	}
	if opts.UpForGrabs {
		params["include_up_for_grabs"] = paramTrue
		params["filter"] = "linked_to_profile"
	}
	if opts.Search != "" {
		params["q"] = opts.Search
	}

	// The Skylight API rejects any chore query without an explicit date window
	// (422 "after can't be blank" / "before can't be blank"), so default one in.
	// Up-for-grabs queries always get a full week-ahead window with any caller-
	// supplied bounds winning. Otherwise explicit bounds are sent as-is — a lone
	// bound passes through untouched so the API's own validation error names the
	// missing side. With no bounds at all: Date becomes a same-day window and
	// everything else falls back to the current calendar month.
	switch {
	case opts.UpForGrabs:
		now := time.Now()
		if opts.After != "" {
			params["after"] = opts.After
		} else {
			params["after"] = now.Format(DateFormat)
		}
		if opts.Before != "" {
			params["before"] = opts.Before
		} else {
			params["before"] = now.AddDate(0, 0, 7).Format(DateFormat)
		}
	case opts.After != "" || opts.Before != "":
		if opts.After != "" {
			params["after"] = opts.After
		}
		if opts.Before != "" {
			params["before"] = opts.Before
		}
	case opts.Date != "":
		params["after"], params["before"] = opts.Date, opts.Date
	default:
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		params["after"], params["before"] = start.Format(DateFormat), start.AddDate(0, 1, -1).Format(DateFormat)
	}
	return params
}

// ListChores retrieves chores for a frame with optional filters.
func (c *Client) ListChores(ctx context.Context, frameID string, opts ChoreListOptions) ([]Chore, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/chores", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create list chores request: %w", err)
	}

	if params := opts.queryParams(); len(params) > 0 {
		addQueryParams(req, params)
	}

	var apiResp choreAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list chores: %w", err)
	}

	searchLower := strings.ToLower(opts.Search)
	chores := make([]Chore, 0, len(apiResp.Data))
	for i := range apiResp.Data {
		c := apiResp.Data[i].toChore()
		if opts.Status != "" && c.Status != opts.Status {
			continue
		}
		if opts.UpForGrabs && !c.UpForGrabs {
			continue
		}
		if searchLower != "" && !strings.Contains(strings.ToLower(c.Title), searchLower) && !strings.Contains(strings.ToLower(c.Description), searchLower) {
			continue
		}
		chores = append(chores, c)
	}

	return chores, nil
}

// CreateChore creates a new chore on a frame. POST /chores ignores frequency, interval and
// recurrence_days, so a chore with Frequency, RecurrenceDays or RecurrenceSet goes through
// create_multiple with an RRULE instead.
func (c *Client) CreateChore(ctx context.Context, frameID string, chore ChoreData) (*Chore, error) {
	if isRecurring(chore) {
		if chore.AssigneeID != "" {
			chore.CategoryIDs = []string{chore.AssigneeID}
		}
		return c.createMultiple(ctx, frameID, chore, "chore")
	}
	if chore.Interval != 0 || chore.EndDate != "" {
		return nil, errRuleRequired
	}

	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/chores", c.effectiveURL(), pathSeg(frameID)), chore)
	if err != nil {
		return nil, fmt.Errorf("failed to create chore request: %w", err)
	}

	var apiResp choreAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create chore: %w", err)
	}

	result := apiResp.Data.toChore()
	return &result, nil
}

// CreateUpForGrabsChore creates a shared chore that anyone can claim, using the
// create_multiple endpoint which accepts up_for_grabs without a category_id.
func (c *Client) CreateUpForGrabsChore(ctx context.Context, frameID string, chore ChoreData) (*Chore, error) {
	chore.UpForGrabs = true
	return c.createMultiple(ctx, frameID, chore, "up-for-grabs chore")
}

var errRuleRequired = errors.New("interval and end date require a frequency or recurrence days: the API replaces the whole rule")

func isRecurring(chore ChoreData) bool {
	return chore.Frequency != "" || len(chore.RecurrenceDays) > 0 || len(chore.RecurrenceSet) > 0
}

// withRule fills RecurrenceSet (and RecurringUntil from EndDate) for a recurring chore.
func withRule(chore ChoreData) (ChoreData, error) {
	if !isRecurring(chore) {
		if chore.Interval != 0 || chore.EndDate != "" {
			return chore, errRuleRequired
		}
		return chore, nil
	}
	if len(chore.RecurrenceSet) > 0 {
		return chore, nil
	}
	rrule, err := choreRRule(chore)
	if err != nil {
		return chore, err
	}
	chore.RecurrenceSet = []string{rrule}
	chore.RecurringUntil = chore.EndDate
	chore.EndDate = ""
	return chore, nil
}

func (c *Client) createMultiple(ctx context.Context, frameID string, chore ChoreData, kind string) (*Chore, error) {
	chore, err := withRule(chore)
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/chores/create_multiple", c.effectiveURL(), pathSeg(frameID)), chore)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s request: %w", kind, err)
	}

	var apiResp choreAPIResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", kind, err)
	}
	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("no chore returned from create_multiple")
	}

	result := apiResp.Data[0].toChore()
	return &result, nil
}

const (
	freqDaily   = "daily"
	freqWeekly  = "weekly"
	freqMonthly = "monthly"
)

var rruleDays = map[string]string{
	"sun": "SU",
	"mon": "MO",
	"tue": "TU",
	"wed": "WE",
	"thu": "TH",
	"fri": "FR",
	"sat": "SA",
}

// choreRRule builds the rule in the form the Skylight app writes, e.g.
// RRULE:FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO,WE. Days without a frequency imply weekly.
func choreRRule(chore ChoreData) (string, error) {
	freq := strings.ToLower(chore.Frequency)
	if freq == "" {
		freq = freqWeekly
	}
	if freq != freqDaily && freq != freqWeekly && freq != freqMonthly {
		return "", fmt.Errorf("invalid frequency %q: must be daily, weekly, or monthly", chore.Frequency)
	}
	if len(chore.RecurrenceDays) > 0 && freq != freqWeekly {
		return "", fmt.Errorf("recurrence days only apply to weekly chores, not %s", freq)
	}
	if chore.Interval < 0 {
		return "", fmt.Errorf("invalid interval %d: must be 1 or more", chore.Interval)
	}
	rule := fmt.Sprintf("RRULE:FREQ=%s;INTERVAL=%d;WKST=SU", strings.ToUpper(freq), max(chore.Interval, 1))
	var days []string
	for _, d := range chore.RecurrenceDays {
		day, ok := rruleDays[strings.ToLower(d)]
		if !ok {
			return "", fmt.Errorf("invalid recurrence day %q: use sun, mon, tue, wed, thu, fri, sat", d)
		}
		if !slices.Contains(days, day) {
			days = append(days, day)
		}
	}
	if len(days) > 0 {
		rule += ";BYDAY=" + strings.Join(days, ",")
	}
	return rule, nil
}

// GetChore retrieves a single chore by ID. The Skylight API has no dedicated
// single-item GET endpoint; this fetches the day's chore list and matches by
// ID. For composite instance IDs like "12345-2026-04-28" the embedded date is
// used as the query window; plain numeric IDs default to today.
func (c *Client) GetChore(ctx context.Context, frameID, choreID string) (*Chore, error) {
	baseID, instanceDate, _ := parseChoreID(choreID)
	date := instanceDate
	if date == "" {
		date = time.Now().Format(DateFormat)
	}

	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/chores", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create get chore request: %w", err)
	}
	addQueryParams(req, map[string]string{
		"after":        date,
		"before":       date,
		"include_late": paramTrue,
	})

	var apiResp choreAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get chore: %w", err)
	}

	for i := range apiResp.Data {
		id := apiResp.Data[i].ID
		entryBase, _, _ := parseChoreID(id)
		if id == choreID || entryBase == baseID {
			result := apiResp.Data[i].toChore()
			return &result, nil
		}
	}
	return nil, &NotFoundError{Resource: "chore", ID: choreID}
}

// UpdateChore updates an existing chore. Composite instance IDs are normalized
// to their base ID before the request. PUT ignores frequency, interval and recurrence_days
// too, so a recurrence change is sent as a replacement RRULE, which replaces the whole schedule.
func (c *Client) UpdateChore(ctx context.Context, frameID, choreID string, chore ChoreData) (*Chore, error) {
	chore, err := withRule(chore)
	if err != nil {
		return nil, err
	}
	baseID, _, _ := parseChoreID(choreID)
	req, err := newRequestWithBody(ctx, "PUT", fmt.Sprintf("%s/frames/%s/chores/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(baseID)), chore)
	if err != nil {
		return nil, fmt.Errorf("failed to create update chore request: %w", err)
	}

	var apiResp choreAPISingleResponse
	if err := c.put(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update chore: %w", err)
	}

	result := apiResp.Data.toChore()
	return &result, nil
}

// SkipChore skips a single instance of a recurring chore. Pass deferUntil as
// a YYYY-MM-DD date to reschedule the instance, or "" to skip without deferring.
func (c *Client) SkipChore(ctx context.Context, frameID, choreID, deferUntil string) error {
	return c.setCompletion(ctx, frameID, choreID, ChoreCompletionData{Status: ChoreStatusSkipped, DeferUntil: deferUntil})
}

// CompleteChore marks a chore instance as completed via the completions endpoint.
func (c *Client) CompleteChore(ctx context.Context, frameID, choreID string) error {
	return c.setCompletion(ctx, frameID, choreID, ChoreCompletionData{Status: ChoreStatusComplete})
}

// ClaimChore assigns an up-for-grabs chore to the given assignee.
func (c *Client) ClaimChore(ctx context.Context, frameID, choreID, assigneeID string) (*Chore, error) {
	return c.UpdateChore(ctx, frameID, choreID, ChoreData{AssigneeID: assigneeID})
}

// DeleteChore deletes a one-time (non-recurring) chore. The Skylight API
// rejects apply_to for one-time chores, so it is not sent.
func (c *Client) DeleteChore(ctx context.Context, frameID, choreID string) error {
	return c.deleteChore(ctx, frameID, choreID, "")
}

// DeleteRecurringChore deletes all instances of a recurring chore.
// apply_to=all is required by the Skylight API for recurring chores.
func (c *Client) DeleteRecurringChore(ctx context.Context, frameID, choreID string) error {
	return c.deleteChore(ctx, frameID, choreID, applyToAll)
}

func (c *Client) deleteChore(ctx context.Context, frameID, choreID, applyTo string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/chores/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(choreID)))
	if err != nil {
		return fmt.Errorf("failed to create delete chore request: %w", err)
	}
	if applyTo != "" {
		addQueryParams(req, map[string]string{queryKeyApplyTo: applyTo})
	}
	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete chore: %w", err)
	}
	return nil
}
