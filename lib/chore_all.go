package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// AllTasksOptions are the query options for ListAllTasks.
type AllTasksOptions struct {
	// Limit caps the total number of entries across all buckets (0 = server default).
	Limit int
	// IncludeUpForGrabs includes unassigned (up-for-grabs) tasks.
	IncludeUpForGrabs bool
}

// TaskBuckets groups tasks by when they are due.
type TaskBuckets struct {
	Late       []Chore `json:"late"`
	Today      []Chore `json:"today"`
	TodayTimed []Chore `json:"today_timed"`
	AnyDay     []Chore `json:"any_day"`
	Future     []Chore `json:"future"`
}

// Count returns the total number of tasks across all buckets.
func (b TaskBuckets) Count() int {
	return len(b.Late) + len(b.Today) + len(b.TodayTimed) + len(b.AnyDay) + len(b.Future)
}

// AllTasks is the consolidated "All Tasks" view: one entry per series, split
// into chores and routines, each bucketed by due time.
type AllTasks struct {
	Chores   TaskBuckets `json:"chores"`
	Routines TaskBuckets `json:"routines"`
}

// ListAllTasks retrieves the consolidated All Tasks view
// (GET /frames/{id}/chores/all).
func (c *Client) ListAllTasks(ctx context.Context, frameID string, opts AllTasksOptions) (*AllTasks, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/chores/all", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create list all tasks request: %w", err)
	}

	params := map[string]string{"include_up_for_grabs": strconv.FormatBool(opts.IncludeUpForGrabs)}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	addQueryParams(req, params)

	var raw map[string]json.RawMessage
	if err := c.get(req, &raw); err != nil {
		return nil, fmt.Errorf("failed to list all tasks: %w", err)
	}

	// Tolerate an optional JSON-API style {"data": {...}} wrapper.
	if inner, ok := raw["data"]; ok && raw["chores"] == nil && raw["routines"] == nil {
		var wrapped map[string]json.RawMessage
		if err := json.Unmarshal(inner, &wrapped); err == nil {
			raw = wrapped
		}
	}

	var out AllTasks
	if out.Chores, err = decodeTaskBuckets(raw["chores"]); err != nil {
		return nil, fmt.Errorf("failed to decode all tasks chores: %w", err)
	}
	if out.Routines, err = decodeTaskBuckets(raw["routines"]); err != nil {
		return nil, fmt.Errorf("failed to decode all tasks routines: %w", err)
	}
	return &out, nil
}

// decodeTaskBuckets decodes one of the "chores"/"routines" sections. Each
// bucket may be a bare array of entries or a {"data": [...]} envelope.
func decodeTaskBuckets(raw json.RawMessage) (TaskBuckets, error) {
	var b TaskBuckets
	if len(raw) == 0 || isJSONNull(raw) {
		return b, nil
	}
	var buckets map[string]json.RawMessage
	if err := json.Unmarshal(raw, &buckets); err != nil {
		return b, err
	}
	for name, dst := range map[string]*[]Chore{
		"late": &b.Late, "today": &b.Today, "today_timed": &b.TodayTimed,
		"any_day": &b.AnyDay, "future": &b.Future,
	} {
		list, err := decodeTaskList(buckets[name])
		if err != nil {
			return b, fmt.Errorf("bucket %q: %w", name, err)
		}
		*dst = list
	}
	return b, nil
}

func decodeTaskList(raw json.RawMessage) ([]Chore, error) {
	if len(raw) == 0 || isJSONNull(raw) {
		return []Chore{}, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		var env struct {
			Data []json.RawMessage `json:"data"`
		}
		if err2 := json.Unmarshal(raw, &env); err2 != nil {
			return nil, err
		}
		items = env.Data
	}
	out := make([]Chore, 0, len(items))
	for _, item := range items {
		c, err := decodeTaskEntry(item)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// decodeTaskEntry decodes one task. JSON-API entries ({id, attributes,
// relationships}) reuse the standard chore flattening; flat objects (fields
// directly on the entry) are decoded as attributes. IDs may be strings or numbers.
func decodeTaskEntry(raw json.RawMessage) (Chore, error) {
	var probe struct {
		ID            json.RawMessage `json:"id"`
		Attributes    json.RawMessage `json:"attributes"`
		Relationships json.RawMessage `json:"relationships"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return Chore{}, err
	}
	id := strings.Trim(string(probe.ID), `"`)
	if id == "null" {
		id = ""
	}

	var e choreAPIEntry
	attrs := raw // flat shape: fields live directly on the entry
	if len(probe.Attributes) > 0 {
		attrs = probe.Attributes
		if len(probe.Relationships) > 0 {
			_ = json.Unmarshal(probe.Relationships, &e.Relationships)
		}
	}
	if err := json.Unmarshal(attrs, &e.Attributes); err != nil {
		return Chore{}, err
	}
	e.ID = id
	return e.toChore(), nil
}

func isJSONNull(raw json.RawMessage) bool { return string(raw) == "null" }
