package lib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func linkedTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	t.Cleanup(func() { SkylightURL = old })
	client, _ := NewClientWithToken("u", "t")
	return client
}

const linkedResp = `{"data":[
 {"id":"1","attributes":{"summary":"Bins"},"relationships":{"category":{"data":{"id":"cat1"}},"linked_task_group":{"data":{"id":"g9","type":"linked_task_group"}}}},
 {"id":"2","attributes":{"summary":"Bins"},"relationships":{"category":{"data":{"id":"cat2"}},"linked_task_group":{"data":{"id":"g9","type":"linked_task_group"}}}}]}`

func TestCreateLinkedChore(t *testing.T) {
	var body map[string]any
	client := linkedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/frames/frame1/chores/create_multiple" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if v := r.Header.Get("skylight-api-version"); v < "2026-08-05" {
			t.Errorf("api version %q too old to link tasks", v)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(linkedResp))
	})

	chores, err := client.CreateLinkedChore(context.Background(), "frame1", ChoreData{
		Title: "Bins", DueDate: "2026-11-20", CategoryIDs: []string{"cat1", "cat2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(body["category_ids"], []any{"cat1", "cat2"}) {
		t.Errorf("category_ids: got %v", body["category_ids"])
	}
	if _, ok := body["category_id"]; ok {
		t.Error("category_id must not be sent for a linked chore")
	}
	if len(chores) != 2 || chores[0].LinkedTaskGroupID != "g9" || chores[1].AssigneeID != "cat2" {
		t.Errorf("unexpected chores: %+v", chores)
	}
}

func TestCreateChore_MultipleCategoriesIsLinked(t *testing.T) {
	client := linkedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/frames/frame1/chores/create_multiple" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(linkedResp))
	})
	chore, err := client.CreateChore(context.Background(), "frame1", ChoreData{Title: "Bins", CategoryIDs: []string{"a", "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if chore.ID != "1" || chore.LinkedTaskGroupID != "g9" {
		t.Errorf("unexpected chore: %+v", chore)
	}
}

func TestCreateLinkedChore_Validation(t *testing.T) {
	client, _ := NewClientWithToken("u", "t")
	for name, in := range map[string]ChoreData{
		"one id":        {Title: "x", CategoryIDs: []string{"a"}},
		"both assignee": {Title: "x", AssigneeID: "a", CategoryIDs: []string{"a", "b"}},
	} {
		if _, err := client.CreateLinkedChore(context.Background(), "f", in); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestUpdateChore_ApplyToProfiles(t *testing.T) {
	var body map[string]any
	client := linkedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/frames/frame1/chores/42" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"data":{"id":"42","attributes":{"summary":"Bins"},"relationships":{"linked_task_group":{"data":{"id":"g9","type":"linked_task_group"}}}}}`))
	})
	chore, err := client.UpdateChore(context.Background(), "frame1", "42-2026-11-20", ChoreData{
		Title: "Bins", ApplyToProfiles: "all", CategoryIDs: []string{"a", "b"}, After: "2026-11-20", Before: "2026-11-20",
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["apply_to_profiles"] != "all" || body["after"] != "2026-11-20" || body["before"] != "2026-11-20" {
		t.Errorf("unexpected body: %v", body)
	}
	if !reflect.DeepEqual(body["category_ids"], []any{"a", "b"}) {
		t.Errorf("category_ids: %v", body["category_ids"])
	}
	if chore.LinkedTaskGroupID != "g9" {
		t.Errorf("LinkedTaskGroupID: %q", chore.LinkedTaskGroupID)
	}
}

func TestUpdateChore_WithoutApplyToProfilesOmitsScope(t *testing.T) {
	var body map[string]any
	client := linkedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"data":{"id":"42","attributes":{}}}`))
	})
	if _, err := client.UpdateChore(context.Background(), "frame1", "42", ChoreData{Title: "x"}); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"apply_to_profiles", "after", "before", "category_ids"} {
		if _, ok := body[k]; ok {
			t.Errorf("%s should be omitted, body=%v", k, body)
		}
	}
}

func TestUpdateChore_InvalidApplyToProfiles(t *testing.T) {
	client, _ := NewClientWithToken("u", "t")
	if _, err := client.UpdateChore(context.Background(), "f", "1", ChoreData{ApplyToProfiles: "some"}); err == nil {
		t.Error("expected error")
	}
}

func TestDeleteLinkedChore(t *testing.T) {
	for _, scope := range []string{"all", "one"} {
		var got string
		client := linkedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			got = r.URL.Query().Get("apply_to_profiles")
			w.WriteHeader(http.StatusNoContent)
		})
		if err := client.DeleteLinkedChore(context.Background(), "frame1", "42", scope); err != nil {
			t.Fatal(err)
		}
		if got != scope {
			t.Errorf("apply_to_profiles: want %q got %q", scope, got)
		}
	}
	client, _ := NewClientWithToken("u", "t")
	if err := client.DeleteLinkedChore(context.Background(), "f", "1", ""); err == nil {
		t.Error("expected error for empty scope")
	}
}
