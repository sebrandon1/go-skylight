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

	out := captureStdout(func() { watchCmd.Run(watchCmd, nil) })
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

	out := captureStdout(func() { watchCmd.Run(watchCmd, nil) })
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

	stderr := captureStderr(func() { watchCmd.Run(watchCmd, nil) })
	if !strings.Contains(stderr, "--persist has no effect") {
		t.Errorf("expected persist warning on stderr, got: %s", stderr)
	}
}

// TestWatchCmd_InvalidInterval_Crasher is invoked as a subprocess by
// TestWatchCmd_InvalidInterval to exercise watchCmd's os.Exit(1) path
// without terminating the real test binary.
func TestWatchCmd_InvalidInterval_Crasher(t *testing.T) {
	if os.Getenv("WANT_WATCH_INTERVAL_CRASH") != "1" {
		t.Skip("only runs as a subprocess of TestWatchCmd_InvalidInterval")
	}
	watchInterval = 0
	watchCmd.Run(watchCmd, nil)
}

func TestWatchCmd_InvalidInterval(t *testing.T) {
	runCrasherTest(t, "TestWatchCmd_InvalidInterval_Crasher", "WANT_WATCH_INTERVAL_CRASH", "--interval must be at least 1")
}
