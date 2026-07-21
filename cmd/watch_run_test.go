package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func watchRunMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/chores"):
			fmt.Fprint(w, `{"data":[]}`)
		case strings.HasSuffix(r.URL.Path, "/calendar_events"):
			fmt.Fprint(w, `{"data":[]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	}
}

func TestWatchCmd_StopsOnContextCancel(t *testing.T) {
	newCmdTestClient(t, watchRunMockHandler())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	watchCmd.SetContext(ctx)
	t.Cleanup(func() { watchCmd.SetContext(context.Background()) })

	origInterval, origResources, origPersist := watchInterval, watchResources, watchPersist
	watchInterval, watchResources, watchPersist = 1, "chores", false
	t.Cleanup(func() { watchInterval, watchResources, watchPersist = origInterval, origResources, origPersist })

	out := captureStdout(func() {
		if err := watchCmd.RunE(watchCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Watching") || !strings.Contains(out, "Stopped.") {
		t.Errorf("expected watch start/stop messages, got: %s", out)
	}
}

func TestWatchCmd_PersistRewards(t *testing.T) {
	newCmdTestClient(t, watchRunMockHandler())

	// lib.NewRewardsPoller defaults to ~/.skylight/poller-state.json when
	// given an empty path; point HOME at a temp dir so this test can't
	// touch the real user's state file.
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	watchCmd.SetContext(ctx)
	t.Cleanup(func() { watchCmd.SetContext(context.Background()) })

	origInterval, origResources, origPersist := watchInterval, watchResources, watchPersist
	watchInterval, watchResources, watchPersist = 1, "rewards", true
	t.Cleanup(func() { watchInterval, watchResources, watchPersist = origInterval, origResources, origPersist })

	out := captureStdout(func() {
		if err := watchCmd.RunE(watchCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "rewards persisted") {
		t.Errorf("expected persist label in output, got: %s", out)
	}
}

func TestWatchCmd_PersistWithoutRewardsWarns(t *testing.T) {
	newCmdTestClient(t, watchRunMockHandler())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	watchCmd.SetContext(ctx)
	t.Cleanup(func() { watchCmd.SetContext(context.Background()) })

	origInterval, origResources, origPersist := watchInterval, watchResources, watchPersist
	watchInterval, watchResources, watchPersist = 1, "chores", true
	t.Cleanup(func() { watchInterval, watchResources, watchPersist = origInterval, origResources, origPersist })

	stderr := captureStderr(func() {
		if err := watchCmd.RunE(watchCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(stderr, "--persist has no effect") {
		t.Errorf("expected persist warning on stderr, got: %s", stderr)
	}
}

func TestWatchCmd_InvalidInterval(t *testing.T) {
	origInterval := watchInterval
	watchInterval = 0
	t.Cleanup(func() { watchInterval = origInterval })

	err := watchCmd.RunE(watchCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}
	if !strings.Contains(err.Error(), "--interval must be at least 1") {
		t.Errorf("expected '--interval must be at least 1' in error, got: %v", err)
	}
}

func TestWatchCmd_MissingFrameID(t *testing.T) {
	origFrameID, origInterval := frameID, watchInterval
	frameID = ""
	watchInterval = 1
	t.Cleanup(func() { frameID, watchInterval = origFrameID, origInterval })

	err := watchCmd.RunE(watchCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing frame ID, got nil")
	}
	if !strings.Contains(err.Error(), "--frame-id is required") {
		t.Errorf("expected '--frame-id is required' in error, got: %v", err)
	}
}

func TestWatchCmd_AllInvalidResources(t *testing.T) {
	newCmdTestClient(t, watchRunMockHandler())
	origResources, origInterval := watchResources, watchInterval
	watchResources = "bogus"
	watchInterval = 1
	t.Cleanup(func() { watchResources, watchInterval = origResources, origInterval })

	err := watchCmd.RunE(watchCmd, nil)
	if err == nil {
		t.Fatal("expected error for all-invalid resources, got nil")
	}
	if !strings.Contains(err.Error(), "no valid resources specified") {
		t.Errorf("expected 'no valid resources specified' in error, got: %v", err)
	}
}

func TestWatchCmd_FrameAPIError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	origResources, origInterval := watchResources, watchInterval
	watchResources = "chores"
	watchInterval = 1
	t.Cleanup(func() { watchResources, watchInterval = origResources, origInterval })

	err := watchCmd.RunE(watchCmd, nil)
	if err == nil {
		t.Fatal("expected error when frame API returns 500, got nil")
	}
	if !strings.Contains(err.Error(), "getting frame info") {
		t.Errorf("expected 'getting frame info' in error, got: %v", err)
	}
}
