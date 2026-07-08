package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/enforce"
	"github.com/WagnerJust/gyroscope/internal/standard"
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
			paths := hookPathsFor(cfg)
			adapters := enabledAdapters(cfg)
			items := classify(abs, files)

			if !apply {
				fmt.Fprintf(stdout, "gyroscope init (dry-run)\n  repo: %s\n", abs)
				// Per-file convergence: NEW (create) / OK (already current) /
				// MERGE (inject the managed region) / CONFLICT (differs, needs --force).
				for _, it := range items {
					fmt.Fprintf(stdout, "  %-8s %s\n", it.State, it.Dest)
				}
				// Write() appends .local/ to .gitignore whenever any planned file
				// lands under .local/ (the local-notes or local-todo spoke);
				// surface that mutation honestly by reading the plan, not one spoke.
				if plansLocalWrite(files) {
					fmt.Fprintf(stdout, "  update: .gitignore (ensure .local/ is listed)\n")
				}
				for _, a := range adapters {
					fmt.Fprintf(stdout, "  %s\n", a.PlanLine(paths))
				}
				// The merge-safe apply writes NEW + MERGE automatically; only a
				// CONFLICT needs --force. Surface that when any conflict is present.
				if c := conflicts(items); len(c) > 0 {
					fmt.Fprintf(stdout, "\n%d conflict(s) need --force: %s\n", len(c), strings.Join(c, ", "))
				}
				fmt.Fprintln(stdout, "\nNothing written. Re-run with --apply.")
				return nil
			}

			if err := applyConverge(stdout, abs, items, adapters, paths, force, false); err != nil {
				return err
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&apply, "apply", false, "Actually write (default is dry-run).")
	f.BoolVarP(&force, "force", "f", false, "Overwrite existing files.")
	return cmd
}

// hookPathsFor returns the repo-relative files the SessionStart hook should cat:
// always the hub, plus the agents-instructions spoke, the state files (so a
// fresh session resumes from current progress), and the local-notes spoke, when
// each is enabled. The context/adr/personas/contributing spokes are not catted
// by the hook — they are read on demand via the hub's routes. gyroscope.json is
// catted when personas are enabled, so the live personas state is in session
// context for the hub's personas directive.
func hookPathsFor(cfg config.Config) []string {
	paths := []string{"AGENTS.md"}
	if cfg.Spokes.Agents {
		paths = append(paths, "docs/agents.md")
	}
	if cfg.Spokes.State {
		paths = append(paths, "TODO.md", ".local/todo.md")
	}
	if cfg.Spokes.Local {
		paths = append(paths, ".local/local.md")
	}
	if cfg.Spokes.Personas.Enabled() {
		// The hub's personas directive branches on spokes.personas; cat the config
		// so the live state is in session context (the hook stays pure `cat`).
		paths = append(paths, "gyroscope.json")
	}
	return paths
}

// enabledAdapters returns the enforcement adapters cfg turns on, in a stable
// order. init installs each; check verifies each.
func enabledAdapters(cfg config.Config) []enforce.Adapter {
	var a []enforce.Adapter
	if cfg.Enforce.Claude {
		a = append(a, enforce.Claude{})
	}
	if cfg.Enforce.PI {
		a = append(a, enforce.PI{})
	}
	return a
}

// plansLocalWrite reports whether any planned file lands under .local/, which is
// what triggers standard.Write to append .local/ to .gitignore.
func plansLocalWrite(files []standard.File) bool {
	for _, f := range files {
		if strings.HasPrefix(f.Dest, ".local/") {
			return true
		}
	}
	return false
}
