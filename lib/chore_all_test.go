package lib

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newAllTasksClient(t *testing.T, status int, body string, check func(*http.Request)) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	t.Cleanup(func() { SkylightURL = old })
	client, _ := NewClientWithToken("u", "t")
	return client
}

func TestListAllTasks_JSONAPI(t *testing.T) {
	body := `{
	  "chores": {
	    "late": [{"id":"1","attributes":{"summary":"Late one","status":"pending","reward_points":3},"relationships":{"category":{"data":{"id":"cat1","type":"category"}}}}],
	    "today": [{"id":2,"attributes":{"summary":"Today one","up_for_grabs":true}}],
	    "today_timed": [], "any_day": null,
	    "future": [{"id":"4","attributes":{"summary":"Later","start":"2026-12-01"}}]
	  },
	  "routines": {"today_timed": [{"id":"9-2026-11-17-0600","attributes":{"summary":"Morning","routine":true,"start_time":"06:00"}}]}
	}`
	var gotPath, gotQuery string
	c := newAllTasksClient(t, http.StatusOK, body, func(r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
	})

	got, err := c.ListAllTasks(context.Background(), "f1", AllTasksOptions{Limit: 25, IncludeUpForGrabs: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/frames/f1/chores/all" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "include_up_for_grabs=true&limit=25" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(got.Chores.Late) != 1 || got.Chores.Late[0].Title != "Late one" || got.Chores.Late[0].AssigneeID != "cat1" || got.Chores.Late[0].Points != 3 {
		t.Errorf("late = %+v", got.Chores.Late)
	}
	if len(got.Chores.Today) != 1 || got.Chores.Today[0].ID != "2" || !got.Chores.Today[0].UpForGrabs {
		t.Errorf("today = %+v", got.Chores.Today)
	}
	if got.Chores.AnyDay == nil || len(got.Chores.Future) != 1 || got.Chores.Future[0].DueDate != "2026-12-01" {
		t.Errorf("any_day/future = %+v / %+v", got.Chores.AnyDay, got.Chores.Future)
	}
	if len(got.Routines.TodayTimed) != 1 || !got.Routines.TodayTimed[0].Routine || got.Routines.TodayTimed[0].StartTime != "06:00" {
		t.Errorf("routines = %+v", got.Routines)
	}
	if got.Chores.Count() != 3 || got.Routines.Count() != 1 {
		t.Errorf("counts = %d, %d", got.Chores.Count(), got.Routines.Count())
	}
}

func TestListAllTasks_FlatAndEnvelopeShapes(t *testing.T) {
	body := `{"data":{"chores":{"today":{"data":[{"id":"7","summary":"Flat","reward_points":2}]}}}}`
	var gotQuery string
	c := newAllTasksClient(t, http.StatusOK, body, func(r *http.Request) { gotQuery = r.URL.RawQuery })

	got, err := c.ListAllTasks(context.Background(), "f1", AllTasksOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "include_up_for_grabs=false" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(got.Chores.Today) != 1 || got.Chores.Today[0].Title != "Flat" || got.Chores.Today[0].Points != 2 {
		t.Errorf("today = %+v", got.Chores.Today)
	}
}

func TestListAllTasks_Empty(t *testing.T) {
	c := newAllTasksClient(t, http.StatusOK, `{}`, nil)
	got, err := c.ListAllTasks(context.Background(), "f1", AllTasksOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Chores.Count() != 0 || got.Routines.Count() != 0 {
		t.Errorf("expected empty, got %+v", got)
	}
}

func TestListAllTasks_Errors(t *testing.T) {
	c := newAllTasksClient(t, http.StatusInternalServerError, `{}`, nil)
	if _, err := c.ListAllTasks(context.Background(), "f1", AllTasksOptions{}); err == nil {
		t.Error("expected error on 500")
	}
	c = newAllTasksClient(t, http.StatusOK, `{"chores":{"today":"oops"}}`, nil)
	if _, err := c.ListAllTasks(context.Background(), "f1", AllTasksOptions{}); err == nil {
		t.Error("expected decode error")
	}
}

func TestListAllTasks_IncludedEnvelopeAndBadShapes(t *testing.T) {
	c := newAllTasksClient(t, http.StatusOK,
		`{"chores":{"late":{"data":[{"id":"5","attributes":{"summary":"Env"}}],"included":[{"id":"x"}]}},"routines":null}`, nil)
	got, err := c.ListAllTasks(context.Background(), "f1", AllTasksOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Chores.Late) != 1 || got.Chores.Late[0].Title != "Env" || got.Routines.Count() != 0 {
		t.Errorf("got %+v", got)
	}

	for name, body := range map[string]string{
		"chores section not object": `{"chores":[1]}`,
		"routines bad entry":        `{"routines":{"late":[{"id":"1","attributes":"x"}]}}`,
		"entry not object":          `{"chores":{"late":[5]}}`,
	} {
		c = newAllTasksClient(t, http.StatusOK, body, nil)
		if _, err := c.ListAllTasks(context.Background(), "f1", AllTasksOptions{}); err == nil {
			t.Errorf("%s: expected decode error", name)
		}
	}
}
