// Command gyroscope installs an opinionated, self-enforcing agent-doc standard
// into a repo. The binary writes the standard's scaffolds, pointer files, and a
// Claude SessionStart hook; the companion skill runs the interview and fills the
// scaffolds. Default is dry-run; --apply writes.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const (
	exitCannotRun = 2
	exitInternal  = 4
)

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }
func errCannotRun(err error) error { return &exitError{exitCannotRun, err} }

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "gyroscope:", err)
		var ee *exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		os.Exit(exitInternal)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	root := &cobra.Command{
		Use:           "gyroscope",
		Short:         "Install an opinionated, self-enforcing agent-doc standard.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
	root.AddCommand(newInitCmd(stdout, stderr))
	root.AddCommand(newVersionCmd(stdout))
	root.AddCommand(newInstallSkillCmd(stdout, stderr))
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		// Cobra usage errors (bad flags/args, unknown command) come back unwrapped;
		// classify them as cannot-run (exit 2) rather than internal (exit 4).
		var ee *exitError
		if !errors.As(err, &ee) {
			return errCannotRun(err)
		}
		return err
	}
	return nil
}
