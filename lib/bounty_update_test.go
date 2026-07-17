package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateBountySuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case (r.Method == http.MethodPut || r.Method == http.MethodPatch) && r.URL.Path == "/api/frames/frame1/chores/ch1":
			if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
				Data: choreAPIEntry{ID: "ch1", Attributes: choreAPIAttributes{Summary: "Updated chore", RewardPoints: 7}},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case (r.Method == http.MethodPut || r.Method == http.MethodPatch) && r.URL.Path == "/api/frames/frame1/rewards/rw1":
			var entry rewardAPIEntry
			entry.ID = "rw1"
			entry.Attributes.Name = "Updated prize"
			entry.Attributes.EmojiIcon = "⭐"
			entry.Attributes.PointValue = 7
			if err := json.NewEncoder(w).Encode(rewardAPISingleResponse{Data: entry}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	b, err := client.UpdateBounty("frame1", "ch1", "rw1", BountyData{
		Title: "Updated chore", Points: 7, RewardTitle: "Updated prize", EmojiIcon: "⭐",
	})
	if err != nil {
		t.Fatalf("UpdateBounty: %v", err)
	}
	if b.Chore.Title != "Updated chore" {
		t.Errorf("chore title: got %q", b.Chore.Title)
	}
	if b.Reward.Title != "Updated prize" {
		t.Errorf("reward title: got %q", b.Reward.Title)
	}
	if b.Reward.Points != 7 {
		t.Errorf("reward points: got %d", b.Reward.Points)
	}
}

func TestUpdateBountyChoreError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"fail"}`))
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.UpdateBounty("frame1", "ch1", "rw1", BountyData{Title: "x", Points: 1, RewardTitle: "y"})
	if err == nil {
		t.Fatal("expected chore update error")
	}
	if !strings.Contains(err.Error(), "failed to update bounty chore") {
		t.Fatalf("want chore error, got %v", err)
	}
}

func TestUpdateBountyRewardError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case (r.Method == http.MethodPut || r.Method == http.MethodPatch) && r.URL.Path == "/api/frames/frame1/chores/ch1":
			_ = json.NewEncoder(w).Encode(choreAPISingleResponse{
				Data: choreAPIEntry{ID: "ch1", Attributes: choreAPIAttributes{Summary: "ok", RewardPoints: 1}},
			})
		case (r.Method == http.MethodPut || r.Method == http.MethodPatch) && r.URL.Path == "/api/frames/frame1/rewards/rw1":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"reward fail"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.UpdateBounty("frame1", "ch1", "rw1", BountyData{Title: "x", Points: 1, RewardTitle: "y"})
	if err == nil {
		t.Fatal("expected reward update error")
	}
	if !strings.Contains(err.Error(), "failed to update bounty reward") {
		t.Fatalf("want reward error, got %v", err)
	}
}
