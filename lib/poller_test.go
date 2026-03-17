package lib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func makePollerServer(t *testing.T, rewards []Reward, categories []Category) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/frames/f1/rewards":
			entries := make([]rewardAPIEntry, len(rewards))
			for i, rw := range rewards {
				redeemed := "2024-01-01T00:00:00Z"
				var redeemedAt *string
				if rw.Redeemed {
					redeemedAt = &redeemed
				}
				entries[i] = rewardAPIEntry{
					ID: rw.ID,
					Attributes: struct {
						Name                string  `json:"name"`
						EmojiIcon           string  `json:"emoji_icon"`
						PointValue          int     `json:"point_value"`
						RespawnOnRedemption bool    `json:"respawn_on_redemption"`
						RedeemedAt          *string `json:"redeemed_at"`
					}{
						Name:       rw.Title,
						PointValue: rw.Points,
						RedeemedAt: redeemedAt,
					},
				}
				if rw.CategoryID != "" {
					entries[i].Relationships.Category.Data = &struct {
						ID string `json:"id"`
					}{ID: rw.CategoryID}
				}
			}
			if err := json.NewEncoder(w).Encode(rewardAPIResponse{Data: entries}); err != nil {
				t.Errorf("encode rewards: %v", err)
			}
		case "/api/frames/f1/categories":
			entries := make([]categoryAPIEntry, len(categories))
			for i, cat := range categories {
				entries[i] = categoryAPIEntry{ID: cat.ID, Attributes: struct {
					Label         string  `json:"label"`
					Color         string  `json:"color"`
					ProfilePicURL *string `json:"profile_pic_url"`
				}{Label: cat.Name, Color: cat.Color}}
			}
			if err := json.NewEncoder(w).Encode(categoryAPIResponse{Data: entries}); err != nil {
				t.Errorf("encode categories: %v", err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestRewardsPollerEmitsNewRedemptions(t *testing.T) {
	rewards := []Reward{
		{ID: "rw1", Title: "Ice cream", Points: 10, Redeemed: true, CategoryID: "cat1"},
		{ID: "rw2", Title: "Movie night", Points: 20, Redeemed: false},
	}
	categories := []Category{{ID: "cat1", Name: "Alice"}}

	srv := makePollerServer(t, rewards, categories)
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	stateFile := filepath.Join(t.TempDir(), "state.json")
	client, _ := NewClientWithToken("u", "t")
	p := NewRewardsPoller(client, "f1", time.Hour, stateFile)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	p.Start(ctx)

	select {
	case event := <-p.Events():
		if event.RewardID != "rw1" {
			t.Errorf("want rw1, got %s", event.RewardID)
		}
		if event.ChildName != "Alice" {
			t.Errorf("want Alice, got %s", event.ChildName)
		}
		if event.Points != 10 {
			t.Errorf("want 10 points, got %d", event.Points)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for redemption event")
	}

	p.Stop()
}

func TestRewardsPollerDeduplication(t *testing.T) {
	rewards := []Reward{
		{ID: "rw1", Title: "Prize", Points: 5, Redeemed: true},
	}

	srv := makePollerServer(t, rewards, nil)
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	stateFile := filepath.Join(t.TempDir(), "state.json")
	client, _ := NewClientWithToken("u", "t")
	p := NewRewardsPoller(client, "f1", 50*time.Millisecond, stateFile)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	p.Start(ctx)

	// Collect events for a bit; rw1 should only appear once.
	var count int
	deadline := time.After(500 * time.Millisecond)
loop:
	for {
		select {
		case <-p.Events():
			count++
		case <-deadline:
			break loop
		}
	}
	p.Stop()

	if count != 1 {
		t.Errorf("expected exactly 1 event (dedup), got %d", count)
	}
}

func TestRewardsPollerStatePersistence(t *testing.T) {
	rewards := []Reward{
		{ID: "rw1", Title: "Cookie", Points: 3, Redeemed: true},
	}

	srv := makePollerServer(t, rewards, nil)
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	stateFile := filepath.Join(t.TempDir(), "state.json")
	client, _ := NewClientWithToken("u", "t")

	// First poller: should emit rw1 and persist it.
	p1 := NewRewardsPoller(client, "f1", time.Hour, stateFile)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()
	p1.Start(ctx1)

	select {
	case <-p1.Events():
	case <-time.After(2 * time.Second):
		t.Fatal("timeout on first poller")
	}
	p1.Stop()

	// State file should now exist.
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("state file not written: %v", err)
	}

	// Second poller: loads state, should NOT emit rw1 again.
	p2 := NewRewardsPoller(client, "f1", time.Hour, stateFile)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	p2.Start(ctx2)

	select {
	case ev := <-p2.Events():
		t.Errorf("unexpected second emission for %s", ev.RewardID)
	case <-time.After(500 * time.Millisecond):
		// Good — no duplicate event.
	}
	p2.Stop()
}

func TestRewardsPollerStop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/frames/f1/rewards":
			if err := json.NewEncoder(w).Encode(rewardAPIResponse{}); err != nil {
				t.Errorf("encode: %v", err)
			}
		case "/api/frames/f1/categories":
			if err := json.NewEncoder(w).Encode([]Category{}); err != nil {
				t.Errorf("encode: %v", err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	p := NewRewardsPoller(client, "f1", 50*time.Millisecond, filepath.Join(t.TempDir(), "state.json"))
	ctx := context.Background()
	p.Start(ctx)

	// Stop should return quickly without hanging.
	done := make(chan struct{})
	go func() {
		p.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not return in time")
	}
}
