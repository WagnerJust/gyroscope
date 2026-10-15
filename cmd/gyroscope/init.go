package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/enforce"
	"github.com/WagnerJust/gyroscope/internal/persona"
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
			paths := injectPathsFor(cfg)
			adapters := enabledAdapters(cfg)
			items := classify(abs, files, cfg)
			// Generated files: persona registrations (docs/agents/ → each harness's
			// subagent registry) + command re-encodings (docs/commands/ → each
			// harness's command dir). Gated per harness by enforce.*.
			mirrors, err := planMirrors(abs, cfg)
			if err != nil {
				return errCannotRun(err)
			}

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
				// Surface the persona mirrors init would write, like the other planned
				// writes. This is a byte copy for Claude registration, not a render.
				for _, m := range mirrors {
					fmt.Fprintf(stdout, "  mirror: %s (from %s)\n", m.Dest, m.Src)
				}
				// The merge-safe apply writes NEW + MERGE automatically; only a
				// CONFLICT needs --force. Surface that when any conflict is present.
				if c := conflicts(items); len(c) > 0 {
					fmt.Fprintf(stdout, "\n%d conflict(s) need --force: %s\n", len(c), strings.Join(c, ", "))
				}
				fmt.Fprintln(stdout, "\nNothing written. Re-run with --apply.")
				return nil
			}

			skipped, err := applyConverge(stdout, abs, items, adapters, paths, force)
			if err != nil {
				return err
			}
			// Register the personas: mirror each into .claude/agents/. Independent of
			// the doc convergence above (a doc CONFLICT never blocks registration), and
			// a no-op when the gate is closed (mirrors is empty).
			written, err := persona.Write(abs, mirrors)
			if err != nil {
				return errCannotRun(err)
			}
			for _, dest := range written {
				fmt.Fprintf(stdout, "generated %s\n", dest)
			}
			// Suppress AI attribution when configured (enforce.aiAttribution=false):
			// write includeCoAuthoredBy:false into .claude/settings.json.
			if err := applyAttribution(stdout, abs, cfg); err != nil {
				return err
			}
			// Merge declared MCP servers into OpenCode's shared opencode.json (Claude's
			// dedicated .mcp.json rode planMirrors above).
			if err := applyMCP(stdout, abs, cfg); err != nil {
				return err
			}
			// When sync is enabled, write the committed reconstruction script so a
			// teammate who gitignores the generated dirs can rebuild the registrations.
			if err := applySyncScript(stdout, abs, cfg); err != nil {
				return err
			}
			// The safe subset landed. If any CONFLICT was skipped, the repo is not
			// yet fully conformant — report drift (exit 1) so a caller/CI can see
			// that --force is still needed, mirroring `check`'s drift signal. A clean
			// full convergence exits 0.
			if len(skipped) > 0 {
				return errDrift(fmt.Errorf("%d conflict(s) need --force: %s", len(skipped), strings.Join(skipped, ", ")))
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&apply, "apply", false, "Actually write (default is dry-run).")
	f.BoolVarP(&force, "force", "f", false, "Overwrite existing files.")
	return cmd
}

// injectPathsFor splits the repo-relative files each enforcement adapter
// delivers into two channels. Rules are static: always the hub, plus the agents
// spoke when enabled. State is the dynamic working state a fresh session resumes
// from: gyroscope.json when personas are enabled (the hub's personas directive
// branches on its live state), the state files, and the local-notes spoke. The
// context/adr/personas/contributing spokes are in neither — they are read on
// demand via the hub's routes, and DONE.md is routed, never injected.
func injectPathsFor(cfg config.Config) enforce.Paths {
	p := enforce.Paths{Rules: []string{"AGENTS.md"}}
	if cfg.Spokes.Agents {
		p.Rules = append(p.Rules, "docs/agents.md")
	}
	if cfg.Spokes.Personas.Enabled() {
		p.State = append(p.State, "gyroscope.json")
	}
	if cfg.Spokes.State {
		p.State = append(p.State, "TODO.md", ".local/todo.md")
	}
	if cfg.Spokes.Local {
		p.State = append(p.State, ".local/local.md")
	}
	return p
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
	if cfg.Enforce.OpenCode {
		a = append(a, enforce.OpenCode{})
	}
	if cfg.Enforce.Cursor {
		a = append(a, enforce.Cursor{})
	}
	return a
}

// plansLocalWrite reports whether any planned file lands under .local/, which is
// what triggers applyConverge to append .local/ to .gitignore.
func plansLocalWrite(files []standard.File) bool {
	for _, f := range files {
		if strings.HasPrefix(f.Dest, ".local/") {
			return true
		}
	}
	return false
}
