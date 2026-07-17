package cmd

import (
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

func TestOutputFlagCompletion(t *testing.T) {
	fn, ok := rootCmd.GetFlagCompletionFunc("output")
	if !ok || fn == nil {
		t.Fatal("expected completion func registered for --output")
	}
	got, _ := fn(rootCmd, nil, "")
	joined := strings.Join(got, ",")
	if !strings.Contains(joined, outputJSON) || !strings.Contains(joined, outputTable) {
		t.Errorf("output completions = %v, want json and table", got)
	}
}

func TestChoreStatusFlagCompletion(t *testing.T) {
	for _, cmd := range []*cobra.Command{choreListCmd, choreUpdateCmd} {
		fn, ok := cmd.GetFlagCompletionFunc("status")
		if !ok || fn == nil {
			t.Fatalf("expected completion func for %s --status", cmd.Name())
		}
		got, _ := fn(cmd, nil, "")
		joined := strings.Join(got, ",")
		for _, want := range []string{lib.ChoreStatusPending, lib.ChoreStatusComplete, lib.ChoreStatusSkipped} {
			if !strings.Contains(joined, want) {
				t.Errorf("%s --status completions = %v, missing %q", cmd.Name(), got, want)
			}
		}
	}
}
