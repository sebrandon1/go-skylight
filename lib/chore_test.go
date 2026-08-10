package lib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListChores(t *testing.T) {
	tests := []struct {
		name       string
		opts       ChoreListOptions
		status     int
		response   string
		wantLen    int
		wantTitle  string
		wantAssign string
		wantDesc   string
		wantErr    bool
	}{
		{
			name:      "returns chores",
			status:    http.StatusOK,
			response:  `{"data":[{"id":"1","attributes":{"summary":"Clean room","status":"pending"}},{"id":"2","attributes":{"summary":"Do homework","status":"completed"}}]}`,
			wantLen:   2,
			wantTitle: "Clean room",
		},
		{
			name:      "returns chore with description",
			status:    http.StatusOK,
			response:  `{"data":[{"id":"1","attributes":{"summary":"Clean room","description":"Vacuum and mop","status":"pending"}}]}`,
			wantLen:   1,
			wantTitle: "Clean room",
			wantDesc:  "Vacuum and mop",
		},
		{
			name:     "passes date filter",
			opts:     ChoreListOptions{Date: "2024-01-15"},
			status:   http.StatusOK,
			response: `{"data":[]}`,
		},
		{
			name:     "passes status filter",
			opts:     ChoreListOptions{Status: "completed"},
			status:   http.StatusOK,
			response: `{"data":[]}`,
		},
		{
			name:     "passes all filters",
			opts:     ChoreListOptions{Date: "2024-01-15", Status: "pending", AssigneeID: "cat1"},
			status:   http.StatusOK,
			response: `{"data":[]}`,
		},
		{
			name:       "flattens category relationship",
			status:     http.StatusOK,
			response:   `{"data":[{"id":"1","attributes":{"summary":"Clean room","status":"pending","start":"2024-01-15","reward_points":5},"relationships":{"category":{"data":{"id":"cat123","type":"category"}}}}]}`,
			wantLen:    1,
			wantTitle:  "Clean room",
			wantAssign: "cat123",
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			status:   http.StatusOK,
			response: `not valid json`,
			wantErr:  true,
		},
		{
			name:     "sets include_late param",
			opts:     ChoreListOptions{IncludeLate: true},
			status:   http.StatusOK,
			response: `{"data":[]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/chores" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				q := r.URL.Query()
				if tc.opts.Date != "" && q.Get("date") != tc.opts.Date {
					t.Errorf("date: want %q got %q", tc.opts.Date, q.Get("date"))
				}
				if tc.opts.Status != "" && q.Get("status") != tc.opts.Status {
					t.Errorf("status: want %q got %q", tc.opts.Status, q.Get("status"))
				}
				if tc.opts.AssigneeID != "" && q.Get("assignee_id") != tc.opts.AssigneeID {
					t.Errorf("assignee_id: want %q got %q", tc.opts.AssigneeID, q.Get("assignee_id"))
				}
				if tc.opts.IncludeLate && q.Get("include_late") != "true" {
					t.Errorf("include_late: want true got %q", q.Get("include_late"))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			chores, err := client.ListChores(context.Background(), "frame1", tc.opts)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(chores) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(chores))
			}
			if tc.wantTitle != "" && len(chores) > 0 && chores[0].Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, chores[0].Title)
			}
			if tc.wantAssign != "" && len(chores) > 0 && chores[0].AssigneeID != tc.wantAssign {
				t.Errorf("AssigneeID: want %q got %q", tc.wantAssign, chores[0].AssigneeID)
			}
			if tc.wantDesc != "" && len(chores) > 0 && chores[0].Description != tc.wantDesc {
				t.Errorf("Description: want %q got %q", tc.wantDesc, chores[0].Description)
			}
		})
	}
}

func TestCreateChore(t *testing.T) {
	tests := []struct {
		name      string
		input     ChoreData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "creates chore",
			input:     ChoreData{Title: "Walk dog", Points: 5},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"3","attributes":{"summary":"Walk dog","reward_points":5}}}`,
			wantTitle: "Walk dog",
		},
		{
			name:      "sends all fields in request body",
			input:     ChoreData{Title: "Walk the dog", DueDate: "2024-01-15", Points: 10, AssigneeID: "cat1"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"c1","attributes":{"summary":"Walk the dog"}}}`,
			wantTitle: "Walk the dog",
		},
		{
			name:      "sends description in request body",
			input:     ChoreData{Title: "Clean room", Description: "Vacuum and mop the floor"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"c2","attributes":{"summary":"Clean room","description":"Vacuum and mop the floor"}}}`,
			wantTitle: "Clean room",
		},
		{
			name:    "server error returns error",
			input:   ChoreData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/chores" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var raw map[string]any
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if raw["summary"] != tc.input.Title {
					t.Errorf("summary: want %q got %v", tc.input.Title, raw["summary"])
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			chore, err := client.CreateChore(context.Background(), "frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && chore.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, chore.Title)
			}
		})
	}
}

func TestUpdateChore(t *testing.T) {
	tests := []struct {
		name       string
		choreID    string
		input      ChoreData
		status     int
		response   string
		wantStatus string
		wantErr    bool
	}{
		{
			name:       "updates chore",
			choreID:    "chore1",
			input:      ChoreData{Title: "Updated chore", Status: "completed"},
			status:     http.StatusOK,
			response:   `{"data":{"id":"1","attributes":{"summary":"Updated chore","status":"completed"}}}`,
			wantStatus: "completed",
		},
		{
			name:    "server error returns error",
			choreID: "chore1",
			input:   ChoreData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			chore, err := client.UpdateChore(context.Background(), "frame1", tc.choreID, tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && chore.Status != tc.wantStatus {
				t.Errorf("Status: want %q got %q", tc.wantStatus, chore.Status)
			}
		})
	}
}

func TestSkipChore(t *testing.T) {
	tests := []struct {
		name       string
		choreID    string
		deferUntil string
		status     int
		wantErr    bool
	}{
		{
			name:    "skips recurring chore via completions endpoint",
			choreID: "18731133-2026-04-28",
			status:  http.StatusOK,
		},
		{
			name:    "skips non-recurring chore (plain ID)",
			choreID: "12345",
			status:  http.StatusOK,
		},
		{
			name:    "skips routine occurrence with time suffix",
			choreID: "97871502-2026-08-10-0600",
			status:  http.StatusOK,
		},
		{
			name:       "sends defer_until when provided",
			choreID:    "18731133-2026-04-28",
			deferUntil: "2026-05-05",
			status:     http.StatusOK,
		},
		{
			name:    "server error returns error",
			choreID: "18731133-2026-04-28",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseID, instanceDate, instanceTime := parseChoreID(tc.choreID)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				wantPath := "/api/frames/frame1/chores/" + baseID + "/completions"
				if r.URL.Path != wantPath {
					t.Errorf("path: want %q got %q", wantPath, r.URL.Path)
				}
				var raw map[string]any
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if raw["status"] != "skipped" {
					t.Errorf("status: want %q got %v", "skipped", raw["status"])
				}
				if instanceDate != "" && raw["instance_date"] != instanceDate {
					t.Errorf("instance_date: want %q got %v", instanceDate, raw["instance_date"])
				}
				if instanceTime != "" && raw["instance_time"] != instanceTime {
					t.Errorf("instance_time: want %q got %v", instanceTime, raw["instance_time"])
				}
				if tc.deferUntil != "" {
					if raw["defer_until"] != tc.deferUntil {
						t.Errorf("defer_until: want %q got %v", tc.deferUntil, raw["defer_until"])
					}
				} else {
					if _, ok := raw["defer_until"]; ok {
						t.Errorf("defer_until should be omitted when empty, got %v", raw["defer_until"])
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.status == http.StatusOK {
					if _, err := w.Write([]byte(`{"data":{"id":"` + baseID + `","attributes":{"status":"skipped"}}}`)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.SkipChore(context.Background(), "frame1", tc.choreID, tc.deferUntil)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestClaimChore(t *testing.T) {
	tests := []struct {
		name       string
		assigneeID string
		status     int
		wantErr    bool
	}{
		{
			name:       "claims chore by setting assignee",
			assigneeID: "member1",
			status:     http.StatusOK,
		},
		{
			name:       "server error returns error",
			assigneeID: "member1",
			status:     http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/chores/chore1" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var raw map[string]any
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if raw["category_id"] != tc.assigneeID {
					t.Errorf("category_id: want %q got %v", tc.assigneeID, raw["category_id"])
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.status == http.StatusOK {
					resp := `{"data":{"id":"chore1","attributes":{"summary":"Clean room"},"relationships":{"category":{"data":{"id":"member1"}}}}}`
					if _, err := w.Write([]byte(resp)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			chore, err := client.ClaimChore(context.Background(), "frame1", "chore1", tc.assigneeID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && chore.AssigneeID != tc.assigneeID {
				t.Errorf("AssigneeID: want %q got %q", tc.assigneeID, chore.AssigneeID)
			}
		})
	}
}

func TestListChoresUpForGrabs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("include_up_for_grabs") != "true" {
			t.Errorf("expected include_up_for_grabs=true query param")
		}
		if r.URL.Query().Get("filter") != "linked_to_profile" {
			t.Errorf("expected filter=linked_to_profile query param")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := `{"data":[{"id":"1","attributes":{"summary":"Wash car","up_for_grabs":true}}]}`
		if _, err := w.Write([]byte(resp)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	chores, err := client.ListChores(context.Background(), "frame1", ChoreListOptions{UpForGrabs: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chores) != 1 {
		t.Fatalf("want 1 chore, got %d", len(chores))
	}
	if !chores[0].UpForGrabs {
		t.Errorf("expected UpForGrabs=true on returned chore")
	}
}

func TestCreateUpForGrabsChore(t *testing.T) {
	tests := []struct {
		name      string
		input     ChoreData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "creates up-for-grabs chore via create_multiple",
			input:     ChoreData{Title: "Pet the cats", DueDate: "2026-04-28"},
			status:    http.StatusOK,
			response:  `{"data":[{"id":"99","attributes":{"summary":"Pet the cats","up_for_grabs":true}}]}`,
			wantTitle: "Pet the cats",
		},
		{
			name:    "server error returns error",
			input:   ChoreData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "empty data returns error",
			input:    ChoreData{Title: "Test"},
			status:   http.StatusOK,
			response: `{"data":[]}`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/chores/create_multiple" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var raw map[string]any
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if raw["up_for_grabs"] != true {
					t.Errorf("up_for_grabs: want true got %v", raw["up_for_grabs"])
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			chore, err := client.CreateUpForGrabsChore(context.Background(), "frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && chore.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, chore.Title)
			}
		})
	}
}

func TestParseChoreID(t *testing.T) {
	tests := []struct {
		input        string
		wantBase     string
		wantInstance string
		wantTime     string
	}{
		{"18731133-2026-04-28", "18731133", "2026-04-28", ""},
		{"70190003-2026-04-28-0600", "70190003", "2026-04-28", "06:00"},
		{"55819571-2026-08-10-2000", "55819571", "2026-08-10", "20:00"},
		{"12345", "12345", "", ""},
		{"abc", "abc", "", ""},
	}
	for _, tc := range tests {
		base, inst, tod := parseChoreID(tc.input)
		if base != tc.wantBase {
			t.Errorf("parseChoreID(%q) base=%q, want %q", tc.input, base, tc.wantBase)
		}
		if inst != tc.wantInstance {
			t.Errorf("parseChoreID(%q) instance=%q, want %q", tc.input, inst, tc.wantInstance)
		}
		if tod != tc.wantTime {
			t.Errorf("parseChoreID(%q) time=%q, want %q", tc.input, tod, tc.wantTime)
		}
	}
}

func TestCompleteChore(t *testing.T) {
	tests := []struct {
		name    string
		choreID string
		status  int
		wantErr bool
	}{
		{
			name:    "completes recurring chore",
			choreID: "18731133-2026-04-28",
			status:  http.StatusOK,
		},
		{
			name:    "completes non-recurring chore",
			choreID: "12345",
			status:  http.StatusOK,
		},
		{
			name:    "completes routine occurrence with time suffix",
			choreID: "55819571-2026-08-10-2000",
			status:  http.StatusOK,
		},
		{
			name:    "server error returns error",
			choreID: "18731133-2026-04-28",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseID, instanceDate, instanceTime := parseChoreID(tc.choreID)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				wantPath := "/api/frames/frame1/chores/" + baseID + "/completions"
				if r.URL.Path != wantPath {
					t.Errorf("path: want %q got %q", wantPath, r.URL.Path)
				}
				var raw map[string]any
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if raw["status"] != ChoreStatusComplete {
					t.Errorf("status: want %q got %v", ChoreStatusComplete, raw["status"])
				}
				if instanceDate != "" && raw["instance_date"] != instanceDate {
					t.Errorf("instance_date: want %q got %v", instanceDate, raw["instance_date"])
				}
				if instanceTime != "" && raw["instance_time"] != instanceTime {
					t.Errorf("instance_time: want %q got %v", instanceTime, raw["instance_time"])
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.status == http.StatusOK {
					if _, err := w.Write([]byte(`{"data":{"id":"` + baseID + `","attributes":{"status":"` + ChoreStatusComplete + `"}}}`)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.CompleteChore(context.Background(), "frame1", tc.choreID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestDeleteChore(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"deletes with 204", http.StatusNoContent, false},
		{"server error returns error", http.StatusInternalServerError, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if got := r.URL.Query().Get("apply_to"); got != "all" {
					t.Errorf("apply_to: want %q got %q", "all", got)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.DeleteChore(context.Background(), "frame1", "chore1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestListChores_RecurrenceFields(t *testing.T) {
	response := `{"data":[{"id":"1","attributes":{"summary":"Daily walk","status":"pending","frequency":"weekly","interval":2,"recurrence_days":["mon","wed"],"end_date":"2026-12-31","recur_from":"completed"}}]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	chores, err := client.ListChores(context.Background(), "frame1", ChoreListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chores) != 1 {
		t.Fatalf("expected 1 chore got %d", len(chores))
	}
	c := chores[0]
	if c.Frequency != "weekly" {
		t.Errorf("Frequency: want %q got %q", "weekly", c.Frequency)
	}
	if c.Interval != 2 {
		t.Errorf("Interval: want 2 got %d", c.Interval)
	}
	if len(c.RecurrenceDays) != 2 || c.RecurrenceDays[0] != "mon" || c.RecurrenceDays[1] != "wed" {
		t.Errorf("RecurrenceDays: got %v", c.RecurrenceDays)
	}
	if c.EndDate != "2026-12-31" {
		t.Errorf("EndDate: want %q got %q", "2026-12-31", c.EndDate)
	}
	if c.RecurFrom != "completed" {
		t.Errorf("RecurFrom: want %q got %q", "completed", c.RecurFrom)
	}
}

func TestListChores_SearchParam(t *testing.T) {
	tests := []struct {
		name      string
		search    string
		response  string
		wantQ     string
		wantLen   int
		wantTitle string
	}{
		{
			name:      "forwards search as q param",
			search:    "dishes",
			response:  `{"data":[{"id":"1","attributes":{"summary":"Wash dishes","status":"pending"}}]}`,
			wantQ:     "dishes",
			wantLen:   1,
			wantTitle: "Wash dishes",
		},
		{
			name:     "client-side filter excludes non-matching chores",
			search:   "vacuum",
			response: `{"data":[{"id":"1","attributes":{"summary":"Wash dishes","status":"pending"}},{"id":"2","attributes":{"summary":"Vacuum floors","status":"pending"}}]}`,
			wantLen:  1,
		},
		{
			name:      "client-side filter is case-insensitive",
			search:    "VACUUM",
			response:  `{"data":[{"id":"1","attributes":{"summary":"Vacuum floors","status":"pending"}}]}`,
			wantLen:   1,
			wantTitle: "Vacuum floors",
		},
		{
			name:      "client-side filter matches description",
			search:    "thoroughly",
			response:  `{"data":[{"id":"1","attributes":{"summary":"Clean room","description":"Clean thoroughly","status":"pending"}}]}`,
			wantLen:   1,
			wantTitle: "Clean room",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.wantQ != "" && r.URL.Query().Get("q") != tc.wantQ {
					t.Errorf("q param: want %q got %q", tc.wantQ, r.URL.Query().Get("q"))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			chores, err := client.ListChores(context.Background(), "frame1", ChoreListOptions{Search: tc.search})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(chores) != tc.wantLen {
				t.Errorf("len: want %d got %d", tc.wantLen, len(chores))
			}
			if tc.wantTitle != "" && len(chores) > 0 && chores[0].Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, chores[0].Title)
			}
		})
	}
}
