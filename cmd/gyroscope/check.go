package main

import (
	"bytes"
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

// newCheckCmd builds the read-only inverse of init: it verifies a repo still
// matches the standard gyroscope would produce for that repo's config and reports
// every nonconformance. It writes nothing. Exit 0 when conformant, exitDrift when
// drift is found, exitCannotRun for a genuine error (bad path, I/O, malformed
// config).
func newCheckCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "check [repo-path]",
		Short: "Verify a repo still conforms to the standard gyroscope would produce (read-only).",
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
			info, err := os.Stat(abs)
			if err != nil {
				return errCannotRun(err)
			}
			if !info.IsDir() {
				return errCannotRun(fmt.Errorf("not a directory: %s", abs))
			}
			cfg, err := config.Load(abs)
			if err != nil {
				return errCannotRun(err)
			}
			problems, err := checkRepo(abs, cfg)
			if err != nil {
				return errCannotRun(err)
			}
			if len(problems) == 0 {
				fmt.Fprintf(stdout, "conformant: %s\n", abs)
				return nil
			}
			for _, p := range problems {
				fmt.Fprintln(stdout, p)
			}
			return errDrift(fmt.Errorf("%d nonconformance(s) found in %s", len(problems), abs))
		},
	}
}

// checkRepo verifies that repoDir still matches the standard gyroscope would write
// for cfg, returning one human-readable line per nonconformance and writing
// nothing. A non-nil error means the repo could not be inspected (I/O, malformed
// settings) — a can't-run condition distinct from drift.
//
// It is the read-only inverse of init: it reuses standard.Plan, standard.Routes,
// target.All/target.RoutingLine, hookPathsFor and enforce.SessionStartCommand
// rather than re-deriving any of them.
func checkRepo(repoDir string, cfg config.Config) ([]string, error) {
	var problems []string
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(repoDir, rel))
		return err == nil
	}

	// 1. Hub present.
	hubPresent := exists("AGENTS.md")
	if !hubPresent {
		problems = append(problems, "AGENTS.md: hub is missing (run `gyroscope init`)")
	}

	// 2 & 6. Every planned file exists, and placeholder-bearing scaffolds have
	// been filled.
	files, err := standard.Plan(cfg)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !exists(f.Dest) {
			// .local/ spokes are gitignored and legitimately absent in a fresh
			// clone, so their absence is not drift — mirror the hub's "may not
			// exist" hedge for .local/local.md.
			if strings.HasPrefix(f.Dest, ".local/") {
				continue
			}
			problems = append(problems, fmt.Sprintf("%s: planned file missing (run `gyroscope init`)", f.Dest))
			continue
		}
		// Only scaffolds that ship WITH a placeholder are policed for a leftover
		// marker: f.Content is the bytes init would write, so a `{{` there marks a
		// fill-once scaffold the skill must fill. A ready-to-use file (e.g. TODO.md)
		// carries no marker, so it may reference the `{{...}}` syntax in prose
		// without tripping the check.
		if bytes.Contains(f.Content, []byte("{{")) {
			b, err := os.ReadFile(filepath.Join(repoDir, f.Dest))
			if err != nil {
				return nil, err
			}
			if bytes.Contains(b, []byte("{{")) {
				problems = append(problems, fmt.Sprintf("%s: unfilled `{{...}}` placeholder remains (the /gyroscope skill fills these)", f.Dest))
			}
		}
	}

	// 3. Hub routes match the enabled spokes: no dangling route to a disabled or
	// missing spoke, no missing route for an enabled one.
	if hubPresent {
		hub, err := os.ReadFile(filepath.Join(repoDir, "AGENTS.md"))
		if err != nil {
			return nil, err
		}
		if got, ok := routesSection(hub); !ok {
			problems = append(problems, "AGENTS.md: `## Routes` section not found")
		} else {
			want := standard.Routes(cfg)
			gotSet := lineSet(got)
			wantSet := lineSet(want)
			for _, l := range nonBlankLines(got) {
				if _, ok := wantSet[l]; !ok {
					problems = append(problems, "AGENTS.md: dangling route (spoke disabled or absent): "+l)
				}
			}
			for _, l := range nonBlankLines(want) {
				if _, ok := gotSet[l]; !ok {
					problems = append(problems, "AGENTS.md: missing route for an enabled spoke: "+l)
				}
			}
		}
	}

	// 4. Every pointer file exists and carries the canonical routing line.
	for _, t := range target.All() {
		if !exists(t.Path) {
			problems = append(problems, fmt.Sprintf("%s: pointer file missing (run `gyroscope init`)", t.Path))
			continue
		}
		b, err := os.ReadFile(filepath.Join(repoDir, t.Path))
		if err != nil {
			return nil, err
		}
		if !strings.Contains(string(b), target.RoutingLine) {
			problems = append(problems, fmt.Sprintf("%s: pointer file does not contain the canonical routing line", t.Path))
		}
	}

	// 5. The SessionStart enforcement hook is installed with the config-aware
	// command that init would write.
	hookCmd := enforce.SessionStartCommand(hookPathsFor(cfg)...)
	installed, err := (enforce.Claude{}).HasSessionStart(repoDir, hookCmd)
	if err != nil {
		return nil, err
	}
	if !installed {
		problems = append(problems, fmt.Sprintf(".claude/settings.json: SessionStart hook missing or altered; expected command: %s", hookCmd))
	}

	return problems, nil
}

// routesSection returns the trimmed body of the hub's `## Routes` section — the
// lines between that heading and the next `## ` heading (or EOF). ok is false when
// no `## Routes` heading is present.
func routesSection(hub []byte) (string, bool) {
	lines := strings.Split(string(hub), "\n")
	start := -1
	for i, ln := range lines {
		if strings.TrimSpace(ln) == "## Routes" {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return "", false
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines[start:end], "\n")), true
}

// nonBlankLines returns the space-trimmed, non-blank lines of s.
func nonBlankLines(s string) []string {
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(ln); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// lineSet is nonBlankLines as a membership set.
func lineSet(s string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, l := range nonBlankLines(s) {
		set[l] = struct{}{}
	}
	return set
}
