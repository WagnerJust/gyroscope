package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/enforce"
	"github.com/WagnerJust/gyroscope/internal/standard"
	"github.com/WagnerJust/gyroscope/internal/target"
)

func newInitCmd(stdout io.Writer) *cobra.Command {
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

			if !force {
				if clashes := existingCollisions(abs, files); len(clashes) > 0 {
					return errCannotRun(fmt.Errorf("refusing to overwrite existing files (use --force): %s", strings.Join(clashes, ", ")))
				}
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

// existingCollisions returns the repo-relative destinations that init would
// write with O_EXCL (standard files plus pointer files) but which already
// exist under abs. It intentionally excludes .claude/settings.json, which is
// installed via a preserving merge rather than a clobbering write. A non-empty
// result means --apply (without --force) must refuse before writing anything.
func existingCollisions(abs string, files []standard.File) []string {
	var clashes []string
	dests := make([]string, 0, len(files)+len(target.All()))
	for _, f := range files {
		dests = append(dests, f.Dest)
	}
	for _, t := range target.All() {
		dests = append(dests, t.Path)
	}
	for _, d := range dests {
		if _, err := os.Stat(filepath.Join(abs, d)); err == nil {
			clashes = append(clashes, d)
		}
	}
	return clashes
}
