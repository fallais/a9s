package cmd

import (
	"a9s/internal/cmd/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display current version",
	Run:   version.Run,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
