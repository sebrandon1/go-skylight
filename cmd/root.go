package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sebrandon1/go-skylight/lib"
)

var ( 
	outputFormat string
	quiet        bool
)

var rootCmd = &cobra.Command{
	Use:   "skylight",
	Short: "Skylight CLI",
}

func Execute() error {
	if err := rootCmd.Execute(); err!= nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP("output", "o", "o", "json", "Output format (json|table)")
	rootCmd.PersistentFlags().BoolVarP("quiet", "q", "q", false, "Quiet mode")
}

func printOutput(data interface{}) error {
	if quiet {
		return nil
	}

	if outputFormat == "table" {
		switch v := data.(type) {
		case *lib.Bounty:
			return printBountiesTable([]*lib.Bounty{v})
		case []*lib.Bounty:
			return printBountiesTable(v)
		default:
			return fmt.Errorf("unsupported type for table output: %T", data)
		}
	}

	return printJSON(data)
}

func printJSON(data interface{}) error {
	// implementation
	return nil
}

func printBountiesTable(bounties []*lib.Bounty) error {
	// implementation
	return nil
}