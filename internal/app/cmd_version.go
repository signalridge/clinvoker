package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd prints the version information.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("clinvoker %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}
