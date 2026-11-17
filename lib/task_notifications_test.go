package lib

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const notifSettingsJSON = `{"data":{"task_due":{"enabled":true,"filters":{"category_ids":["a","b"],"up_for_grabs":false}},"task_completed":{"enabled":false,"filters":{"category_ids":[],"up_for_grabs":true}}}}`

func TestGetTaskNotificationSettings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/frames/f1/task_notification_settings" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(notifSettingsJSON))
	}))
	defer srv.Close()

	client, _ := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"))
	s, err := client.GetTaskNotificationSettings(context.Background(), "f1")
	if err != nil {
		t.Fatalf("GetTaskNotificationSettings: %v", err)
	}
	if !s.TaskDue.Enabled || !reflect.DeepEqual(s.TaskDue.Filters.CategoryIDs, []string{"a", "b"}) {
		t.Errorf("task_due decoded wrong: %+v", s.TaskDue)
	}
	if s.TaskCompleted.Enabled || !s.TaskCompleted.Filters.UpForGrabs {
		t.Errorf("task_completed decoded wrong: %+v", s.TaskCompleted)
	}
}

func TestUpdateTaskNotificationSettings(t *testing.T) {
	on, off := true, false
	cats := []string{"c"}
	empty := []string{}
	tests := []struct {
		name   string
		update TaskNotificationUpdate
		want   string
	}{
		{
			name:   "disabling task_due sends no filters and omits task_completed",
			update: TaskNotificationUpdate{TaskDue: &TaskNotificationChange{Enabled: &off}},
			want:   `{"task_due":{"enabled":false}}`,
		},
		{
			name:   "replace categories and up_for_grabs on task_completed",
			update: TaskNotificationUpdate{TaskCompleted: &TaskNotificationChange{Enabled: &on, CategoryIDs: &cats, UpForGrabs: &off}},
			want:   `{"task_completed":{"enabled":true,"filters":{"category_ids":["c"],"up_for_grabs":false}}}`,
		},
		{
			name:   "clear categories sends an empty array",
			update: TaskNotificationUpdate{TaskDue: &TaskNotificationChange{CategoryIDs: &empty, UpForGrabs: &on}},
			want:   `{"task_due":{"enabled":true,"filters":{"category_ids":[],"up_for_grabs":true}}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.Method {
				case http.MethodGet:
					_, _ = w.Write([]byte(notifSettingsJSON))
				case http.MethodPatch:
					b, _ := io.ReadAll(r.Body)
					got = strings.TrimSpace(string(b))
					_, _ = w.Write([]byte(notifSettingsJSON))
				default:
					t.Errorf("unexpected method %s", r.Method)
				}
			}))
			defer srv.Close()

			client, _ := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"))
			if _, err := client.UpdateTaskNotificationSettings(context.Background(), "f1", tc.update); err != nil {
				t.Fatalf("UpdateTaskNotificationSettings: %v", err)
			}
			var gotV, wantV any
			if err := json.Unmarshal([]byte(got), &gotV); err != nil {
				t.Fatalf("PATCH body not JSON: %q", got)
			}
			_ = json.Unmarshal([]byte(tc.want), &wantV)
			if !reflect.DeepEqual(gotV, wantV) {
				t.Errorf("PATCH body:\n got %s\nwant %s", got, tc.want)
			}
		})
	}
}

func TestUpdateTaskNotificationSettings_NoChange(t *testing.T) {
	client, _ := NewClientWithToken("u", "t", WithBaseURL("http://127.0.0.1:1/api"))
	if _, err := client.UpdateTaskNotificationSettings(context.Background(), "f1", TaskNotificationUpdate{}); err == nil {
		t.Error("expected error for empty update")
	}
}

func TestValidateStartTime(t *testing.T) {
	for _, ok := range []string{"", "00:00", "07:30", "23:59"} {
		if err := ValidateStartTime(ok); err != nil {
			t.Errorf("ValidateStartTime(%q) unexpected error: %v", ok, err)
		}
	}
	for _, bad := range []string{"5pm", "25:00", "24:00", "12:60", "7:30", "07:30:00", "0730", " 07:30"} {
		if err := ValidateStartTime(bad); err == nil {
			t.Errorf("ValidateStartTime(%q) expected error", bad)
		}
	}
}

func TestChoreWrites_RejectBadStartTime(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hit = true }))
	defer srv.Close()
	client, _ := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"))
	ctx := context.Background()
	bad := ChoreData{Title: "x", StartTime: "25:00"}

	if _, err := client.CreateChore(ctx, "f1", bad); err == nil {
		t.Error("CreateChore: expected error")
	}
	if _, err := client.CreateUpForGrabsChore(ctx, "f1", bad); err == nil {
		t.Error("CreateUpForGrabsChore: expected error")
	}
	if _, err := client.UpdateChore(ctx, "f1", "c1", bad); err == nil {
		t.Error("UpdateChore: expected error")
	}
	if hit {
		t.Error("no request should reach the server for an invalid due time")
	}
}

func TestUpdateTaskNotificationSettings_RejectsEnableWithoutRecipients(t *testing.T) {
	patched := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			patched = true
		}
		_, _ = w.Write([]byte(notifSettingsJSON))
	}))
	defer server.Close()
	client, _ := NewClientWithToken("u", "t", WithBaseURL(server.URL+"/api"))

	on := true
	empty := []string{}
	// task_completed starts with up_for_grabs true; clearing it and the
	// (already empty) categories while enabling leaves no recipients.
	off := false
	_, err := client.UpdateTaskNotificationSettings(context.Background(), "f1", TaskNotificationUpdate{
		TaskCompleted: &TaskNotificationChange{Enabled: &on, CategoryIDs: &empty, UpForGrabs: &off},
	})
	if !errors.Is(err, ErrNoNotificationRecipients) {
		t.Fatalf("want ErrNoNotificationRecipients, got %v", err)
	}
	if patched {
		t.Error("no PATCH should be sent when the rule has no recipients")
	}
}

func TestUpdateTaskNotificationSettings_RejectsFiltersOnDisabledRule(t *testing.T) {
	patched := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			patched = true
		}
		_, _ = w.Write([]byte(notifSettingsJSON))
	}))
	defer server.Close()
	client, _ := NewClientWithToken("u", "t", WithBaseURL(server.URL+"/api"))

	// task_completed is disabled in the fixture, so changing filters alone must fail.
	cats := []string{"c"}
	_, err := client.UpdateTaskNotificationSettings(context.Background(), "f1", TaskNotificationUpdate{
		TaskCompleted: &TaskNotificationChange{CategoryIDs: &cats},
	})
	if !errors.Is(err, ErrFiltersOnDisabledRule) {
		t.Fatalf("want ErrFiltersOnDisabledRule, got %v", err)
	}
	if patched {
		t.Error("no PATCH should be sent for filters on a disabled rule")
	}
}

func TestUpdateTaskNotificationSettings_MergesPartialResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			// The API echoes only the rule that was sent.
			_, _ = w.Write([]byte(`{"data":{"task_due":{"enabled":false,"filters":{"category_ids":[],"up_for_grabs":false}}}}`))
			return
		}
		_, _ = w.Write([]byte(notifSettingsJSON))
	}))
	defer server.Close()
	client, _ := NewClientWithToken("u", "t", WithBaseURL(server.URL+"/api"))

	off := false
	s, err := client.UpdateTaskNotificationSettings(context.Background(), "f1", TaskNotificationUpdate{
		TaskDue: &TaskNotificationChange{Enabled: &off},
	})
	if err != nil {
		t.Fatalf("UpdateTaskNotificationSettings: %v", err)
	}
	if s.TaskDue.Enabled {
		t.Errorf("task_due should reflect the PATCH response: %+v", s.TaskDue)
	}
	if !s.TaskCompleted.Filters.UpForGrabs {
		t.Errorf("task_completed should be kept from the earlier GET: %+v", s.TaskCompleted)
	}
}

func TestTaskNotificationSettings_APIErrors(t *testing.T) {
	for _, failOn := range []string{http.MethodGet, http.MethodPatch} {
		t.Run(failOn, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == failOn {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				_, _ = w.Write([]byte(notifSettingsJSON))
			}))
			defer server.Close()
			client, _ := NewClientWithToken("u", "t", WithBaseURL(server.URL+"/api"), WithRetry(1, 0, 0))

			off := false
			_, err := client.UpdateTaskNotificationSettings(context.Background(), "f1", TaskNotificationUpdate{
				TaskDue: &TaskNotificationChange{Enabled: &off},
			})
			if err == nil {
				t.Errorf("expected error when %s fails", failOn)
			}
			if failOn == http.MethodGet {
				if _, err := client.GetTaskNotificationSettings(context.Background(), "f1"); err == nil {
					t.Error("expected GetTaskNotificationSettings error")
				}
			}
		})
	}
}
