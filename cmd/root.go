package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:          "cobra-to-skills",
		Short:        "Generate AI Agent skills from Cobra CLI applications",
		SilenceUsage: true,
	}
	rootCmd.AddCommand(newVersionCmd())
	return rootCmd
}

func Execute() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
