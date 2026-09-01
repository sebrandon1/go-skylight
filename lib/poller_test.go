package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
					entries[i].Relationships.Category.Data = &apiRelationshipData{ID: rw.CategoryID}
				}
			}
			if err := json.NewEncoder(w).Encode(rewardAPIResponse{Data: entries}); err != nil {
				t.Errorf("encode rewards: %v", err)
			}
		case "/api/frames/f1/categories":
			entries := make([]categoryAPIEntry, len(categories))
			for i, cat := range categories {
				e := categoryAPIEntry{ID: cat.ID}
				e.Attributes.Label = cat.Name
				e.Attributes.Color = cat.Color
				entries[i] = e
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

func TestRewardsPollerPrunesUnredeemedIDs(t *testing.T) {
	// After a reward is no longer redeemed, the in-memory seen set (and next
	// save) must drop that ID so state stays O(active redemptions).
	rewards := []Reward{{ID: "rw1", Title: "Cookie", Points: 3, Redeemed: false}}
	srv := makePollerServer(t, rewards, nil)
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	stateFile := filepath.Join(t.TempDir(), "state.json")
	// Seed state file as if rw1 was previously redeemed.
	if err := os.WriteFile(stateFile, []byte(`{"seen_reward_ids":["rw1","stale-id"]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	client, _ := NewClientWithToken("u", "t")
	p := NewRewardsPoller(client, "f1", time.Hour, stateFile)
	if len(p.seen) != 2 {
		t.Fatalf("want 2 ids loaded from state, got %d", len(p.seen))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	p.Start(ctx)
	time.Sleep(200 * time.Millisecond)
	p.Stop()

	p.mu.Lock()
	n := len(p.seen)
	_, hasStale := p.seen["stale-id"]
	_, hasRW1 := p.seen["rw1"]
	p.mu.Unlock()
	if n != 0 || hasStale || hasRW1 {
		t.Errorf("want seen pruned empty after unredeem poll, got n=%d stale=%v rw1=%v", n, hasStale, hasRW1)
	}
}

func TestSameStringSet(t *testing.T) {
	a := map[string]struct{}{"x": {}, "y": {}}
	b := map[string]struct{}{"y": {}, "x": {}}
	if !sameStringSet(a, b) {
		t.Error("expected equal sets")
	}
	if sameStringSet(a, map[string]struct{}{"x": {}}) {
		t.Error("expected unequal sets")
	}
}

func TestNewRewardsPollerDefaultStatePath(t *testing.T) {
	old := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	p := NewRewardsPoller(client, "f1", time.Hour, "")
	if p.stateFile == "" {
		t.Error("stateFile should default to non-empty path")
	}
	if !filepath.IsAbs(p.stateFile) {
		t.Errorf("stateFile should be absolute, got %q", p.stateFile)
	}
}

func TestLoadStateCorruptJSON(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")
	if err := os.WriteFile(stateFile, []byte(`{corrupt`), 0o600); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	old := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t", WithLogger(logger))
	p := NewRewardsPoller(client, "f1", time.Hour, stateFile)
	if len(p.seen) != 0 {
		t.Errorf("seen should be empty after corrupt state, got %d", len(p.seen))
	}
	if !bytes.Contains(buf.Bytes(), []byte("corrupt state file")) {
		t.Error("expected warning about corrupt state file in logs")
	}
}

func TestLoadStateReadErrorNonErrNotExist(t *testing.T) {
	// Use a directory as the state file path to trigger a non-ErrNotExist read error
	dir := t.TempDir()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	old := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t", WithLogger(logger))
	p := NewRewardsPoller(client, "f1", time.Hour, dir)
	if len(p.seen) != 0 {
		t.Errorf("seen should be empty, got %d", len(p.seen))
	}
	if !bytes.Contains(buf.Bytes(), []byte("could not read state file")) {
		t.Error("expected warning about read error in logs")
	}
}

func TestSaveStateMkdirAllError(t *testing.T) {
	// Create a file where a directory should be to cause MkdirAll failure
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(blocker, "subdir", "state.json")

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	old := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t", WithLogger(logger))
	p := NewRewardsPoller(client, "f1", time.Hour, filepath.Join(t.TempDir(), "clean-state.json"))
	p.stateFile = stateFile
	p.logger = logger
	p.seen["test"] = struct{}{}
	p.saveState()
	if !bytes.Contains(buf.Bytes(), []byte("failed to create state directory")) {
		t.Error("expected warning about MkdirAll failure")
	}
}

func TestSaveStateWriteError(t *testing.T) {
	// Create a read-only directory to prevent writing
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(stateDir, 0o500); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(stateDir, "state.json")

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	old := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t", WithLogger(logger))
	p := NewRewardsPoller(client, "f1", time.Hour, filepath.Join(t.TempDir(), "clean-state.json"))
	p.stateFile = stateFile
	p.logger = logger
	p.seen["test"] = struct{}{}
	p.saveState()
	if !bytes.Contains(buf.Bytes(), []byte("failed to write state file")) {
		t.Error("expected warning about write failure")
	}
}

func TestPollerListRewardsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/frames/f1/categories":
			if err := json.NewEncoder(w).Encode(categoryAPIResponse{}); err != nil {
				t.Errorf("encode: %v", err)
			}
		case "/api/frames/f1/rewards":
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
				t.Errorf("write: %v", err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t", WithLogger(logger))
	p := NewRewardsPoller(client, "f1", time.Hour, filepath.Join(t.TempDir(), "state.json"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	p.Start(ctx)

	time.Sleep(200 * time.Millisecond)
	p.Stop()

	if !bytes.Contains(buf.Bytes(), []byte("ListRewards failed")) {
		t.Error("expected warning about ListRewards failure")
	}
}

func TestPollerEventChannelFull(t *testing.T) {
	// Create many redeemed rewards to overflow the 64-buffer channel
	rewards := make([]Reward, 70)
	for i := range rewards {
		rewards[i] = Reward{
			ID:       fmt.Sprintf("rw%d", i),
			Title:    fmt.Sprintf("Prize %d", i),
			Points:   1,
			Redeemed: true,
		}
	}

	srv := makePollerServer(t, rewards, nil)
	defer srv.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t", WithLogger(logger))
	p := NewRewardsPoller(client, "f1", time.Hour, filepath.Join(t.TempDir(), "state.json"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	p.Start(ctx)

	time.Sleep(500 * time.Millisecond)
	p.Stop()

	if !bytes.Contains(buf.Bytes(), []byte("event channel full")) {
		t.Error("expected warning about channel full")
	}
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
func TestRewardsPollerStopIdempotent(t *testing.T) {
	old := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	dir := t.TempDir()
	p := NewRewardsPoller(client, "f1", time.Hour, filepath.Join(dir, "state.json"))
	// Start then stop twice — second Stop must not panic (#269).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Start(ctx)
	p.Stop()
	p.Stop()
}
