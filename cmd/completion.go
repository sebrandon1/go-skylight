package cmd

import (
	"github.com/spf13/cobra"
)

// registerEnumFlagCompletion wires shell completion for a flag with fixed values.
func registerEnumFlagCompletion(cmd *cobra.Command, flag string, values ...string) {
	_ = cmd.RegisterFlagCompletionFunc(flag, cobra.FixedCompletions(values, cobra.ShellCompDirectiveNoFileComp))
}

// registerCommonFlagCompletions registers completion for global enum flags.
func registerCommonFlagCompletions() {
	registerEnumFlagCompletion(rootCmd, "output", outputJSON, outputTable)
}
