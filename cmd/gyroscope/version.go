package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// versionString renders build metadata for display. When built untagged,
// `git describe --always` yields the short sha — identical to the commit — so
// collapse to a single sha rather than printing "<sha> (commit <sha>, ...)".
func versionString(version, commit, date string) string {
	if version == commit {
		return fmt.Sprintf("%s (built %s)", version, date)
	}
	return fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
}

func newVersionCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, commit, and build date.",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			fmt.Fprintf(stdout, "gyroscope %s\n", versionString(version, commit, date))
			return nil
		},
	}
}
