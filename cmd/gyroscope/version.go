package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newVersionCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, commit, and build date.",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			fmt.Fprintf(stdout, "gyroscope %s (commit %s, built %s)\n", version, commit, date)
			return nil
		},
	}
}
