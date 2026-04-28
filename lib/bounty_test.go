package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCreateBounty(t *testing.T) {
	tests := []struct {
		name        string
		input       BountyData
		handler     http.HandlerFunc
		wantChore   string
		wantReward  string
		wantPoints  int
		wantDeleteN int32
		wantErr     bool
	}{
		{
			name:  "derives category_ids from AssigneeID",
			input: BountyData{Title: "Task", Points: 5, RewardTitle: "Prize", AssigneeID: "6750947"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/chores":
					w.WriteHeader(http.StatusCreated)
					if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
						Data: choreAPIEntry{ID: "ch1", Attributes: struct {
							Summary      string `json:"summary"`
							Status       string `json:"status"`
							Start        string `json:"start"`
							RewardPoints int    `json:"reward_points"`
							Recurring    bool   `json:"recurring"`
							UpForGrabs   bool   `json:"up_for_grabs"`
						}{Summary: "Task", RewardPoints: 5}},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/rewards":
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					catIDs, ok := body["category_ids"]
					if !ok {
						t.Error("reward request body missing category_ids")
					} else {
						ids, _ := catIDs.([]any)
						if len(ids) != 1 || ids[0] != float64(6750947) {
							t.Errorf("category_ids: want [6750947], got %v", catIDs)
						}
					}
					w.WriteHeader(http.StatusCreated)
					if err := json.NewEncoder(w).Encode(rewardAPIResponse{
						Data: []rewardAPIEntry{{ID: "rw1", Attributes: struct {
							Name                string  `json:"name"`
							EmojiIcon           string  `json:"emoji_icon"`
							PointValue          int     `json:"point_value"`
							RespawnOnRedemption bool    `json:"respawn_on_redemption"`
							RedeemedAt          *string `json:"redeemed_at"`
						}{Name: "Prize", PointValue: 5}}},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
			},
			wantChore:  "Task",
			wantReward: "Prize",
			wantPoints: 5,
		},
		{
			name:  "no category_ids when AssigneeID empty",
			input: BountyData{Title: "Task", Points: 5, RewardTitle: "Prize"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/chores":
					w.WriteHeader(http.StatusCreated)
					if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
						Data: choreAPIEntry{ID: "ch1", Attributes: struct {
							Summary      string `json:"summary"`
							Status       string `json:"status"`
							Start        string `json:"start"`
							RewardPoints int    `json:"reward_points"`
							Recurring    bool   `json:"recurring"`
							UpForGrabs   bool   `json:"up_for_grabs"`
						}{Summary: "Task", RewardPoints: 5}},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/rewards":
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					if catIDs, ok := body["category_ids"]; ok {
						ids, _ := catIDs.([]any)
						if len(ids) > 0 {
							t.Errorf("expected no category_ids, got %v", catIDs)
						}
					}
					w.WriteHeader(http.StatusCreated)
					if err := json.NewEncoder(w).Encode(rewardAPIResponse{
						Data: []rewardAPIEntry{{ID: "rw1", Attributes: struct {
							Name                string  `json:"name"`
							EmojiIcon           string  `json:"emoji_icon"`
							PointValue          int     `json:"point_value"`
							RespawnOnRedemption bool    `json:"respawn_on_redemption"`
							RedeemedAt          *string `json:"redeemed_at"`
						}{Name: "Prize", PointValue: 5}}},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
			},
			wantChore:  "Task",
			wantReward: "Prize",
			wantPoints: 5,
		},
		{
			name:  "creates chore and reward",
			input: BountyData{Title: "Do dishes", Points: 10, RewardTitle: "Ice cream", EmojiIcon: "🍦"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/chores":
					w.WriteHeader(http.StatusCreated)
					if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
						Data: choreAPIEntry{
							ID: "ch1",
							Attributes: struct {
								Summary      string `json:"summary"`
								Status       string `json:"status"`
								Start        string `json:"start"`
								RewardPoints int    `json:"reward_points"`
								Recurring    bool   `json:"recurring"`
								UpForGrabs   bool   `json:"up_for_grabs"`
							}{Summary: "Do dishes", RewardPoints: 10},
						},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/rewards":
					w.WriteHeader(http.StatusCreated)
					if err := json.NewEncoder(w).Encode(rewardAPIResponse{
						Data: []rewardAPIEntry{{
							ID: "rw1",
							Attributes: struct {
								Name                string  `json:"name"`
								EmojiIcon           string  `json:"emoji_icon"`
								PointValue          int     `json:"point_value"`
								RespawnOnRedemption bool    `json:"respawn_on_redemption"`
								RedeemedAt          *string `json:"redeemed_at"`
							}{Name: "Ice cream", PointValue: 10, EmojiIcon: "🍦"},
						}},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
			},
			wantChore:  "Do dishes",
			wantReward: "Ice cream",
			wantPoints: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			bounty, err := client.CreateBounty("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if bounty.Chore.Title != tc.wantChore {
				t.Errorf("Chore.Title: want %q got %q", tc.wantChore, bounty.Chore.Title)
			}
			if bounty.Reward.Title != tc.wantReward {
				t.Errorf("Reward.Title: want %q got %q", tc.wantReward, bounty.Reward.Title)
			}
			if bounty.Chore.Points != tc.wantPoints || bounty.Reward.Points != tc.wantPoints {
				t.Errorf("Points: want %d, chore=%d reward=%d", tc.wantPoints, bounty.Chore.Points, bounty.Reward.Points)
			}
		})
	}
}

func TestCreateBountyCleanupOnRewardFailure(t *testing.T) {
	var deleteChoreCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/chores":
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
				Data: choreAPIEntry{
					ID: "ch1",
					Attributes: struct {
						Summary      string `json:"summary"`
						Status       string `json:"status"`
						Start        string `json:"start"`
						RewardPoints int    `json:"reward_points"`
						Recurring    bool   `json:"recurring"`
						UpForGrabs   bool   `json:"up_for_grabs"`
					}{Summary: "Test"},
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/rewards":
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
				t.Errorf("write: %v", err)
			}
		case r.Method == http.MethodDelete && r.URL.Path == "/api/frames/frame1/chores/ch1":
			deleteChoreCount.Add(1)
			w.WriteHeader(http.StatusNoContent)
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
	_, err := client.CreateBounty("frame1", BountyData{Title: "Test", Points: 5, RewardTitle: "Prize"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if deleteChoreCount.Load() != 1 {
		t.Errorf("expected 1 cleanup DELETE, got %d", deleteChoreCount.Load())
	}
}

func TestCreateBountyChoreFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.CreateBounty("frame1", BountyData{Title: "Test", Points: 5, RewardTitle: "Prize"})
	if err == nil {
		t.Fatal("expected error when CreateChore fails")
	}
}

func TestCreateBountyInvalidAssigneeID(t *testing.T) {
	var deleteCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/chores":
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
				Data: choreAPIEntry{ID: "ch1"},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case r.Method == http.MethodDelete:
			deleteCount.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.CreateBounty("frame1", BountyData{
		Title:       "Test",
		Points:      5,
		RewardTitle: "Prize",
		AssigneeID:  "not-a-number",
	})
	if err == nil {
		t.Fatal("expected error for invalid AssigneeID")
	}
	if deleteCount.Load() != 1 {
		t.Errorf("expected 1 cleanup DELETE, got %d", deleteCount.Load())
	}
}

func TestListBountiesChoreError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/frames/frame1/chores":
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
				t.Errorf("write: %v", err)
			}
		case "/api/frames/frame1/rewards":
			if err := json.NewEncoder(w).Encode(rewardAPIResponse{}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.ListBounties("frame1")
	if err == nil {
		t.Fatal("expected error when ListChores fails")
	}
}

func TestListBountiesRewardError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/frames/frame1/chores":
			if err := json.NewEncoder(w).Encode(choreAPIResponse{}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "/api/frames/frame1/rewards":
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
				t.Errorf("write: %v", err)
			}
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.ListBounties("frame1")
	if err == nil {
		t.Fatal("expected error when ListRewards fails")
	}
}

func TestListBounties(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantLen    int
		wantChore  string
		wantReward string
		wantErr    bool
	}{
		{
			name: "matches chores to rewards by points",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				switch r.URL.Path {
				case "/api/frames/frame1/chores":
					if err := json.NewEncoder(w).Encode(choreAPIResponse{
						Data: []choreAPIEntry{
							{ID: "ch1", Attributes: struct {
								Summary      string `json:"summary"`
								Status       string `json:"status"`
								Start        string `json:"start"`
								RewardPoints int    `json:"reward_points"`
								Recurring    bool   `json:"recurring"`
								UpForGrabs   bool   `json:"up_for_grabs"`
							}{Summary: "Task A", Status: "pending", RewardPoints: 10}},
							{ID: "ch2", Attributes: struct {
								Summary      string `json:"summary"`
								Status       string `json:"status"`
								Start        string `json:"start"`
								RewardPoints int    `json:"reward_points"`
								Recurring    bool   `json:"recurring"`
								UpForGrabs   bool   `json:"up_for_grabs"`
							}{Summary: "Task B", Status: "pending", RewardPoints: 0}},
						},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				case "/api/frames/frame1/rewards":
					if err := json.NewEncoder(w).Encode(rewardAPIResponse{
						Data: []rewardAPIEntry{{ID: "rw1", Attributes: struct {
							Name                string  `json:"name"`
							EmojiIcon           string  `json:"emoji_icon"`
							PointValue          int     `json:"point_value"`
							RespawnOnRedemption bool    `json:"respawn_on_redemption"`
							RedeemedAt          *string `json:"redeemed_at"`
						}{Name: "Prize", PointValue: 10}}},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
			},
			wantLen:    1,
			wantChore:  "Task A",
			wantReward: "Prize",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			bounties, err := client.ListBounties("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(bounties) != tc.wantLen {
				t.Fatalf("wantLen=%d got %d", tc.wantLen, len(bounties))
			}
			if tc.wantLen > 0 {
				if bounties[0].Chore.Title != tc.wantChore {
					t.Errorf("Chore.Title: want %q got %q", tc.wantChore, bounties[0].Chore.Title)
				}
				if bounties[0].Reward.Title != tc.wantReward {
					t.Errorf("Reward.Title: want %q got %q", tc.wantReward, bounties[0].Reward.Title)
				}
			}
		})
	}
}

func TestListBountiesDateRange(t *testing.T) {
	var gotAfter, gotBefore string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/frames/frame1/chores":
			gotAfter = r.URL.Query().Get("after")
			gotBefore = r.URL.Query().Get("before")
			if err := json.NewEncoder(w).Encode(choreAPIResponse{}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "/api/frames/frame1/rewards":
			if err := json.NewEncoder(w).Encode(rewardAPIResponse{}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
	if _, err := client.ListBounties("frame1"); err != nil {
		t.Fatalf("ListBounties: %v", err)
	}

	if gotAfter == "" {
		t.Error("expected non-empty after query param")
	}
	if gotBefore == "" {
		t.Error("expected non-empty before query param")
	}
	afterT, err := time.Parse(DateFormat, gotAfter)
	if err != nil {
		t.Errorf("after=%q is not a valid date: %v", gotAfter, err)
	}
	beforeT, err := time.Parse(DateFormat, gotBefore)
	if err != nil {
		t.Errorf("before=%q is not a valid date: %v", gotBefore, err)
	}
	if !beforeT.After(afterT) {
		t.Errorf("expected before (%s) > after (%s)", gotBefore, gotAfter)
	}
}
