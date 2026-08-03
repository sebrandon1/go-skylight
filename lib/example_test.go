package lib_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

func ExampleNewClientWithToken() {
	client, err := lib.NewClientWithToken("user-id", "api-token")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(client.UserID)
	fmt.Println(client.APIToken)
	// Output:
	// user-id
	// api-token
}

func ExampleClient_ListCalendarEvents() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"data":[{"id":"1","type":"calendar_event","attributes":{"summary":"Team meeting","starts_at":"","ends_at":"","all_day":false}}]}`)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	old := lib.SkylightURL
	lib.SkylightURL = srv.URL + "/api"
	defer func() { lib.SkylightURL = old }()

	client, _ := lib.NewClientWithToken("u", "t")
	events, err := client.ListCalendarEvents(context.Background(), "frame1", "", "", "")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, e := range events {
		fmt.Println(e.Title)
	}
	// Output:
	// Team meeting
}

// ExampleRewardsPoller demonstrates streaming reward redemption events.
// The poller runs on a configurable interval; here we stop it after the first
// event to keep the example deterministic.
func ExampleRewardsPoller() {
	redeemed := "2024-01-01T00:00:00Z"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/frames/f1/rewards":
			if err := json.NewEncoder(w).Encode(struct {
				Data []struct {
					ID         string `json:"id"`
					Attributes struct {
						Name       string  `json:"name"`
						PointValue int     `json:"point_value"`
						RedeemedAt *string `json:"redeemed_at"`
					} `json:"attributes"`
				} `json:"data"`
			}{Data: []struct {
				ID         string `json:"id"`
				Attributes struct {
					Name       string  `json:"name"`
					PointValue int     `json:"point_value"`
					RedeemedAt *string `json:"redeemed_at"`
				} `json:"attributes"`
			}{{ID: "rw1", Attributes: struct {
				Name       string  `json:"name"`
				PointValue int     `json:"point_value"`
				RedeemedAt *string `json:"redeemed_at"`
			}{Name: "Ice cream", PointValue: 10, RedeemedAt: &redeemed}}}}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "/api/frames/f1/categories":
			if err := json.NewEncoder(w).Encode([]lib.Category{}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Use a fresh temporary state file so the example is deterministic across runs.
	stateDir, _ := os.MkdirTemp("", "skylight-example-*")
	defer os.RemoveAll(stateDir)
	stateFile := filepath.Join(stateDir, "state.json")

	client, _ := lib.NewClientWithToken("u", "t", lib.WithBaseURL(srv.URL+"/api"))
	poller := lib.NewRewardsPoller(client, "f1", time.Hour, stateFile)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	poller.Start(ctx)

	event := <-poller.Events()
	poller.Stop()

	fmt.Println(event.RewardName)
	fmt.Println(event.Points)
	// Output:
	// Ice cream
	// 10
}
