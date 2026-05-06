package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRoutines(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns routines",
			status:   http.StatusOK,
			response: `{"data":[{"id":"1","attributes":{"title":"Morning","assignee_id":"a1","steps":[]}},{"id":"2","attributes":{"title":"Bedtime","assignee_id":"a2","steps":[]}}]}`,
			wantLen:  2,
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/routines" {
					t.Errorf("unexpected path: %s", r.URL.Path)
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
			routines, err := client.ListRoutines("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(routines) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(routines))
			}
		})
	}
}

func TestCreateRoutine(t *testing.T) {
	tests := []struct {
		name      string
		input     RoutineData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "creates routine",
			input:     RoutineData{Title: "Morning", AssigneeID: "a1", Steps: []string{"Brush teeth", "Get dressed"}},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"1","attributes":{"title":"Morning","assignee_id":"a1","steps":[]}}}`,
			wantTitle: "Morning",
		},
		{
			name:    "server error returns error",
			input:   RoutineData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			input:    RoutineData{Title: "Test"},
			status:   http.StatusOK,
			response: `{invalid`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
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
			routine, err := client.CreateRoutine("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && routine.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, routine.Title)
			}
		})
	}
}

func TestUpdateRoutine(t *testing.T) {
	tests := []struct {
		name      string
		routineID string
		input     RoutineData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "updates routine",
			routineID: "r1",
			input:     RoutineData{Title: "Updated"},
			status:    http.StatusOK,
			response:  `{"data":{"id":"r1","attributes":{"title":"Updated","assignee_id":"","steps":[]}}}`,
			wantTitle: "Updated",
		},
		{
			name:      "server error returns error",
			routineID: "r1",
			input:     RoutineData{Title: "Test"},
			status:    http.StatusInternalServerError,
			wantErr:   true,
		},
		{
			name:      "invalid JSON returns error",
			routineID: "r1",
			input:     RoutineData{Title: "Test"},
			status:    http.StatusOK,
			response:  `{invalid`,
			wantErr:   true,
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
			routine, err := client.UpdateRoutine("frame1", tc.routineID, tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && routine != nil && routine.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, routine.Title)
			}
		})
	}
}

func TestDeleteRoutine(t *testing.T) {
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
				if r.URL.Path != "/api/frames/frame1/routines/r1" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.DeleteRoutine("frame1", "r1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestReorderRoutines(t *testing.T) {
	tests := []struct {
		name       string
		routineIDs []string
		status     int
		response   string
		wantErr    bool
	}{
		{
			name:       "reorders routines",
			routineIDs: []string{"r2", "r1", "r3"},
			status:     http.StatusOK,
			response:   `{}`,
		},
		{
			name:       "sends ids in request body",
			routineIDs: []string{"a", "b"},
			status:     http.StatusNoContent,
		},
		{
			name:       "server error returns error",
			routineIDs: []string{"r1"},
			status:     http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/routines/reorder" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var body struct {
					IDs []string `json:"ids"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if len(body.IDs) != len(tc.routineIDs) {
					t.Errorf("ids: want %v got %v", tc.routineIDs, body.IDs)
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
			err := client.ReorderRoutines("frame1", tc.routineIDs)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}
