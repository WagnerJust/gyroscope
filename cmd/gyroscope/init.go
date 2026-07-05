package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/enforce"
	"github.com/WagnerJust/gyroscope/internal/standard"
	"github.com/WagnerJust/gyroscope/internal/target"
)

func newInitCmd(stdout, stderr io.Writer) *cobra.Command {
	var apply, force bool
	cmd := &cobra.Command{
		Use:   "init [repo-path]",
		Short: "Write the standard, pointer files, and enforcement hook into a repo.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			repo := "."
			if len(args) == 1 {
				repo = args[0]
			}
			abs, err := filepath.Abs(repo)
			if err != nil {
				return errCannotRun(err)
			}
			cfg, err := config.Load(abs)
			if err != nil {
				return errCannotRun(err)
			}
			files, err := standard.Plan(cfg)
			if err != nil {
				return errCannotRun(err)
			}

			if !apply {
				fmt.Fprintf(stdout, "gyroscope init (dry-run)\n  repo: %s\n", abs)
				for _, f := range files {
					fmt.Fprintf(stdout, "  write: %s\n", f.Dest)
				}
				for _, t := range target.All() {
					fmt.Fprintf(stdout, "  write: %s (pointer)\n", t.Path)
				}
				fmt.Fprintf(stdout, "  merge: .claude/settings.json (SessionStart hook)\n")
				fmt.Fprintln(stdout, "\nNothing written. Re-run with --apply.")
				return nil
			}

			written, err := standard.Write(abs, files, force)
			if err != nil {
				return errCannotRun(err)
			}
			for _, w := range written {
				fmt.Fprintf(stdout, "wrote %s\n", w)
			}
			for _, t := range target.All() {
				if err := target.WritePointer(abs, t, force); err != nil {
					return errCannotRun(err)
				}
				fmt.Fprintf(stdout, "wrote %s (pointer)\n", t.Path)
			}
			changed, err := (enforce.Claude{}).Install(abs)
			if err != nil {
				return errCannotRun(err)
			}
			if changed {
				fmt.Fprintln(stdout, "installed SessionStart hook → .claude/settings.json")
			} else {
				fmt.Fprintln(stdout, "SessionStart hook already present")
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&apply, "apply", false, "Actually write (default is dry-run).")
	f.BoolVarP(&force, "force", "f", false, "Overwrite existing files.")
	return cmd
}
