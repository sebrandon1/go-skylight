package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestParseWatchResources_All(t *testing.T) {
	cases := []string{"", "all"}
	for _, c := range cases {
		got, err := parseWatchResources(c)
		if err != nil {
			t.Fatalf("parseWatchResources(%q): unexpected error: %v", c, err)
		}
		if len(got) != len(allWatchResources) {
			t.Errorf("parseWatchResources(%q): expected %d resources, got %d: %v", c, len(allWatchResources), len(got), got)
		}
	}
}

func TestParseWatchResources_Specific(t *testing.T) {
	got, err := parseWatchResources("rewards,chores")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 resources, got %d: %v", len(got), got)
	}
	if got[0] != "rewards" || got[1] != "chores" {
		t.Errorf("unexpected resources: %v", got)
	}
}

func TestParseWatchResources_Single(t *testing.T) {
	got, err := parseWatchResources("calendar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "calendar" {
		t.Errorf("expected [calendar], got %v", got)
	}
}

func TestParseWatchResources_TrimsSpaces(t *testing.T) {
	got, err := parseWatchResources(" rewards , chores ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 resources, got %d: %v", len(got), got)
	}
	if got[0] != "rewards" || got[1] != "chores" {
		t.Errorf("expected trimmed resources, got %v", got)
	}
}

func TestWatchState_TracksSeenRewards(t *testing.T) {
	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	if _, seen := state.seenRewardIDs["r1"]; seen {
		t.Error("expected r1 not yet seen")
	}
	state.seenRewardIDs["r1"] = struct{}{}
	if _, seen := state.seenRewardIDs["r1"]; !seen {
		t.Error("expected r1 to be seen after marking")
	}
}

func TestWatchState_TracksSeenChores(t *testing.T) {
	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	if _, seen := state.seenChoreIDs["c1"]; seen {
		t.Error("expected c1 not yet seen")
	}
	state.seenChoreIDs["c1"] = struct{}{}
	if _, seen := state.seenChoreIDs["c1"]; !seen {
		t.Error("expected c1 to be seen after marking")
	}
}

func TestWatchState_TracksSeenEvents(t *testing.T) {
	state := &watchState{
		seenRewardIDs: make(map[string]struct{}),
		seenChoreIDs:  make(map[string]struct{}),
		seenEventIDs:  make(map[string]struct{}),
	}

	if _, seen := state.seenEventIDs["e1"]; seen {
		t.Error("expected e1 not yet seen")
	}
	state.seenEventIDs["e1"] = struct{}{}
	if _, seen := state.seenEventIDs["e1"]; !seen {
		t.Error("expected e1 to be seen after marking")
	}
}

func TestParseWatchResources_IgnoresInvalid(t *testing.T) {
	got, err := parseWatchResources("rewards,bogus,chores")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 valid resources, got %d: %v", len(got), got)
	}
	if got[0] != "rewards" || got[1] != "chores" {
		t.Errorf("expected [rewards chores], got %v", got)
	}
}

func TestWatchCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "watch" {
			found = true
			break
		}
	}
	if !found {
		t.Error("watch command not registered on root")
	}
}

func TestWatchCmdHasFlags(t *testing.T) {
	if f := watchCmd.Flags().Lookup("interval"); f == nil {
		t.Error("expected --interval flag on watch command")
	}
	if f := watchCmd.Flags().Lookup("resources"); f == nil {
		t.Error("expected --resources flag on watch command")
	}
}

func TestWatchCmdLong_ContainsResources(t *testing.T) {
	long := watchCmd.Long
	for _, r := range []string{"rewards", "chores", "calendar", "lists", "routines", "meals", "photos"} {
		if !strings.Contains(long, r) {
			t.Errorf("expected %q mentioned in watch command long description", r)
		}
	}
}

func TestWatchState_TracksSeenPhotos(t *testing.T) {
	state := &watchState{seenPhotoIDs: make(map[string]struct{})}

	if _, seen := state.seenPhotoIDs["p1"]; seen {
		t.Error("expected p1 not yet seen")
	}
	state.seenPhotoIDs["p1"] = struct{}{}
	if _, seen := state.seenPhotoIDs["p1"]; !seen {
		t.Error("expected p1 to be seen after marking")
	}
}

func TestWatchCmd_ExitsWhenTokenRefreshIsRejected(t *testing.T) {
	oauthCalls := 0
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oauthCalls++
		w.Header().Set("Content-Type", "application/json")
		if oauthCalls == 1 {
			fmt.Fprint(w, `{"access_token":"a1","refresh_token":"r1"}`)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer oauth.Close()
	origOAuth := lib.OAuthURL
	lib.OAuthURL = oauth.URL
	t.Cleanup(func() { lib.OAuthURL = origOAuth })

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/chores") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"timezone":"UTC"}}}`)
	}))
	defer api.Close()

	client, err := lib.NewClientWithRefreshToken("r0", "fp", lib.WithBaseURL(api.URL))
	if err != nil {
		t.Fatal(err)
	}
	pointClientAt(t, client)
	origInterval, origResources, origPersist := watchInterval, watchResources, watchPersist
	watchInterval, watchResources, watchPersist = 1, "chores", false
	t.Cleanup(func() { watchInterval, watchResources, watchPersist = origInterval, origResources, origPersist })

	var runErr error
	captureStderr(func() {
		captureStdout(func() { runErr = watchCmd.RunE(watchCmd, nil) })
	})
	if !lib.IsAuthError(runErr) {
		t.Errorf("error = %v, want an AuthError", runErr)
	}
}

func TestWatchCmd_ExitsWhenRefreshIsRejectedAfterStartup(t *testing.T) {
	oauthCalls := 0
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oauthCalls++
		w.Header().Set("Content-Type", "application/json")
		if oauthCalls == 1 {
			fmt.Fprint(w, `{"access_token":"a1","refresh_token":"r1"}`)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer oauth.Close()
	origOAuth := lib.OAuthURL
	lib.OAuthURL = oauth.URL
	t.Cleanup(func() { lib.OAuthURL = origOAuth })

	polls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.HasSuffix(r.URL.Path, "/chores") {
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"timezone":"UTC"}}}`)
			return
		}
		polls++
		if polls > 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer api.Close()

	client, err := lib.NewClientWithRefreshToken("r0", "fp", lib.WithBaseURL(api.URL))
	if err != nil {
		t.Fatal(err)
	}
	pointClientAt(t, client)
	origInterval, origResources, origPersist := watchInterval, watchResources, watchPersist
	watchInterval, watchResources, watchPersist = 1, "chores", false
	t.Cleanup(func() { watchInterval, watchResources, watchPersist = origInterval, origResources, origPersist })

	var runErr error
	captureStderr(func() {
		captureStdout(func() { runErr = watchCmd.RunE(watchCmd, nil) })
	})
	if !lib.IsAuthError(runErr) {
		t.Errorf("error = %v, want an AuthError", runErr)
	}
}
