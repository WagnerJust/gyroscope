package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/WagnerJust/gyroscope/internal/archive"
	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/enforce"
	"github.com/WagnerJust/gyroscope/internal/fsutil"
	"github.com/WagnerJust/gyroscope/internal/persona"
	"github.com/WagnerJust/gyroscope/internal/standard"
	"github.com/WagnerJust/gyroscope/internal/target"
)

// newCheckCmd builds the read-only inverse of init: it verifies a repo still
// matches the standard gyroscope would produce for that repo's config and reports
// every nonconformance. It writes nothing. Exit 0 when conformant, exitDrift when
// drift is found, exitCannotRun for a genuine error (bad path, I/O, malformed
// config).
func newCheckCmd(stdout io.Writer) *cobra.Command {
	var fix bool
	cmd := &cobra.Command{
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

			// --fix makes check symmetric with init: it auto-applies the safe
			// convergence (create NEW files, inject the hub's managed region) and
			// leaves genuine conflicts untouched, so `check` (detect) and fix
			// (converge) share one code path. Conflicts still surface as drift
			// below — --fix never clobbers user content.
			if fix {
				files, err := standard.Plan(cfg)
				if err != nil {
					return errCannotRun(err)
				}
				items := classify(abs, files)
				adapters := enabledAdapters(cfg)
				paths := hookPathsFor(cfg)
				// force=false: create NEW files, merge the hub's managed region, and
				// leave conflicts untouched (they surface as drift in the re-check
				// below). --fix never clobbers user content.
				if _, err := applyConverge(stdout, abs, items, adapters, paths, false); err != nil {
					return err
				}
				// Re-mirror personas so registration converges too. gyroscope owns the
				// .claude/agents/ mirror files, so this safely overwrites a drifted
				// mirror (unlike the user-content-preserving convergence above). A
				// closed gate mirrors nothing.
				mirrors, err := planPersonaMirrors(abs, cfg)
				if err != nil {
					return errCannotRun(err)
				}
				written, err := persona.Write(abs, mirrors)
				if err != nil {
					return errCannotRun(err)
				}
				for _, dest := range written {
					fmt.Fprintf(stdout, "mirrored persona %s\n", dest)
				}

				// Archive completed top-level tasks: move `- [x]` items from the
				// injected TODO.md into the non-injected DONE.md so the per-session
				// context stays lean. This is the convergence half of the archive
				// nudge (ADR 0009) — until now the only check finding --fix could not
				// fix. Gated on the state spoke; a no-op when nothing is done to move.
				if cfg.Spokes.State {
					moved, err := archiveDone(abs)
					if err != nil {
						return errCannotRun(err)
					}
					if moved > 0 {
						fmt.Fprintf(stdout, "archived %d completed item(s) to DONE.md\n", moved)
					}
				}

				// Suppress AI attribution when configured (enforce.aiAttribution=false):
				// converge includeCoAuthoredBy:false into .claude/settings.json.
				if err := applyAttribution(stdout, abs, cfg); err != nil {
					return err
				}
			}

			problems, notes, err := checkRepo(abs, cfg)
			if err != nil {
				return errCannotRun(err)
			}
			// Soft notes (e.g. the archive nudge) print regardless of drift but never
			// affect the exit code — they are housekeeping advice, not nonconformance.
			for _, n := range notes {
				fmt.Fprintln(stdout, n)
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
	f := cmd.Flags()
	f.BoolVar(&fix, "fix", false, "Auto-apply the safe convergence (create NEW files, merge the hub's managed region); conflicts still report as drift.")
	return cmd
}

// archiveDone moves completed top-level tasks from TODO.md into DONE.md's
// `## Completed` section and returns how many blocks moved. It rewrites both files
// atomically. A missing TODO.md, or no completed top-level `[x]` item, is a no-op
// returning 0 (nothing written). DONE.md is created from the moved content when
// absent, though `check --fix`'s convergence normally scaffolds it first.
//
// DONE.md is written before TODO.md on purpose: if the second write fails, the
// items are duplicated (still in TODO, now also in DONE) rather than lost — a safe
// re-run converges, and no completed work vanishes.
func archiveDone(repoDir string) (int, error) {
	todoBytes, err := os.ReadFile(filepath.Join(repoDir, "TODO.md"))
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	remaining, moved := archive.Plan(string(todoBytes))
	if len(moved) == 0 {
		return 0, nil
	}
	doneBytes, err := os.ReadFile(filepath.Join(repoDir, "DONE.md"))
	if err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	newDone := archive.Merge(string(doneBytes), moved)
	if err := fsutil.WriteAtomic(repoDir, "DONE.md", []byte(newDone)); err != nil {
		return 0, err
	}
	if err := fsutil.WriteAtomic(repoDir, "TODO.md", []byte(remaining)); err != nil {
		return 0, err
	}
	return len(moved), nil
}

// archiveNudgeThreshold is the count of completed `[x]` items in TODO.md above
// which check emits the archive nudge. TODO.md is catted into every session, so a
// pile of done items is dead weight in the per-session context; past this many the
// heuristic tells the user to move them to DONE.md. A handful is tolerated (the
// legend line and an in-flight item's just-finished sub-tasks), so the bar is
// deliberately generous rather than zero.
const archiveNudgeThreshold = 5

// checkRepo verifies that repoDir still matches the standard gyroscope would write
// for cfg. It returns one human-readable line per nonconformance (problems), plus
// soft notes — advice that prints but never affects the exit code (the archive
// nudge). Writing nothing. A non-nil error means the repo could not be inspected
// (I/O, malformed settings) — a can't-run condition distinct from drift.
//
// It is the read-only inverse of init: it reuses standard.Plan, standard.Routes,
// target.All/target.RoutingLine, hookPathsFor and enforce.SessionStartCommand
// rather than re-deriving any of them.
func checkRepo(repoDir string, cfg config.Config) (problems, notes []string, err error) {
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
		return nil, nil, err
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
				return nil, nil, err
			}
			if bytes.Contains(b, []byte("{{")) {
				problems = append(problems, fmt.Sprintf("%s: unfilled `{{...}}` placeholder remains (the /gyroscope skill fills these)", f.Dest))
			}
		}
	}

	// Checks 3 and 7 inspect only gyroscope's managed region of the hub — the
	// content between the managed markers. Everything outside is the user's and is
	// invisible to check (ADR 0007), so user-authored sections and their own
	// route-like bullets never count as drift.
	if hubPresent {
		hub, err := os.ReadFile(filepath.Join(repoDir, "AGENTS.md"))
		if err != nil {
			return nil, nil, err
		}
		region, ok := standard.ManagedRegion(hub)
		if !ok {
			// A hub with no well-formed managed region is not in the managed-block
			// form gyroscope now writes — the whole route/directive comparison is
			// moot, so report it once and skip the region-scoped checks.
			problems = append(problems, "AGENTS.md: managed region not found (run `gyroscope init --apply` to migrate to the managed-block form)")
		} else {
			// 3. Hub routes match the enabled spokes: no dangling route to a disabled
			// or missing spoke, no missing route for an enabled one.
			if got, ok := routesSection(region); !ok {
				problems = append(problems, "AGENTS.md: `## Routes` section not found in the managed region")
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

			// 7. When personas are enabled, the managed region carries the standing
			// personas directive.
			if cfg.Spokes.Personas.Enabled() {
				if !strings.Contains(string(region), standard.PersonasDirective(cfg)) {
					problems = append(problems, "AGENTS.md: personas directive missing or altered (run `gyroscope init`)")
				}
			}

			// 7c. When AI attribution is suppressed, the managed region carries the
			// standing no-attribution directive — the cross-harness (prompt-injection)
			// half of the toggle, for harnesses without a native co-author setting.
			if d := standard.AttributionDirective(cfg); d != "" && !strings.Contains(string(region), d) {
				problems = append(problems, "AGENTS.md: AI-attribution directive missing or altered (run `gyroscope init`)")
			}
		}
	}

	// 7b. Managed spokes: any planned file other than the hub whose standard content
	// carries a gyroscope-managed region (e.g. CONTRIBUTING.md's contributor block)
	// must, on disk, carry that region byte-for-byte. The hub is excluded — its
	// region is config-rendered and verified semantically above (routes + personas
	// directive); a spoke's region is static, so byte-equality is the right test. A
	// missing or drifted region is nonconformance; `check --fix` re-injects it.
	for _, f := range files {
		if f.Dest == "AGENTS.md" {
			continue
		}
		want, ok := standard.ManagedRegion(f.Content)
		if !ok {
			continue // not a managed spoke
		}
		if !exists(f.Dest) {
			continue // absence is already reported by the planned-file check above
		}
		b, err := os.ReadFile(filepath.Join(repoDir, f.Dest))
		if err != nil {
			return nil, nil, err
		}
		got, ok := standard.ManagedRegion(b)
		if !ok {
			problems = append(problems, fmt.Sprintf("%s: gyroscope-managed region missing (run `gyroscope check --fix`)", f.Dest))
			continue
		}
		if !bytes.Equal(got, want) {
			problems = append(problems, fmt.Sprintf("%s: gyroscope-managed region differs from the standard (run `gyroscope check --fix`)", f.Dest))
		}
	}

	// 8. State `on` means personas are wired: at least one non-README file under
	// docs/agents/.
	if cfg.Spokes.Personas == config.PersonaOn {
		entries, err := os.ReadDir(filepath.Join(repoDir, "docs", "agents"))
		if err != nil && !os.IsNotExist(err) {
			return nil, nil, err
		}
		wired := false
		for _, e := range entries {
			if !e.IsDir() && e.Name() != "README.md" {
				wired = true
				break
			}
		}
		if !wired {
			problems = append(problems, "docs/agents/: personas state is `on` but no persona files present (only README)")
		}

		// 8b. Registration: when the persona mirror is gated (personas on AND the
		// Claude adapter enabled), each canonical persona must be mirrored into
		// .claude/agents/<name>.md, byte-equal to its docs/agents/ source — that is
		// what makes Claude register it as a dispatchable subagent (ADR 0010). A
		// missing or drifted mirror is nonconformance; `check --fix` re-mirrors above.
		if personaMirrorGated(cfg) {
			mirrors, err := persona.Plan(repoDir)
			if err != nil {
				return nil, nil, err
			}
			for _, m := range mirrors {
				got, err := os.ReadFile(filepath.Join(repoDir, m.Dest))
				if os.IsNotExist(err) {
					problems = append(problems, fmt.Sprintf("%s: persona not registered — mirror missing (run `gyroscope init --apply` or `check --fix`)", m.Dest))
					continue
				}
				if err != nil {
					return nil, nil, err
				}
				if !bytes.Equal(got, m.Bytes) {
					problems = append(problems, fmt.Sprintf("%s: persona mirror differs from its docs/agents/ source (run `check --fix` to re-mirror)", m.Dest))
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
			return nil, nil, err
		}
		if !strings.Contains(string(b), target.RoutingLine) {
			problems = append(problems, fmt.Sprintf("%s: pointer file does not contain the canonical routing line", t.Path))
		}
	}

	// 5. Every enabled enforcement adapter's mechanism is installed and current.
	paths := hookPathsFor(cfg)
	for _, a := range enabledAdapters(cfg) {
		installed, err := a.Verify(repoDir, paths)
		if err != nil {
			return nil, nil, err
		}
		if !installed {
			problems = append(problems, fmt.Sprintf("enforcement adapter %q: not installed or altered (run `gyroscope init`) — %s", a.ID(), a.PlanLine(paths)))
		}
	}

	// Archive nudge (soft, never drift): TODO.md is injected every session, so a
	// pile of completed `[x]` items is dead per-session weight. When the state
	// spoke is on and TODO.md carries more than the threshold, advise moving them
	// to DONE.md — the enforcement half of the TODO/DONE move convention (ADR 0009).
	// It is a note, not a problem, so an otherwise-conformant repo still exits 0;
	// housekeeping shouldn't break CI.
	if cfg.Spokes.State {
		if b, err := os.ReadFile(filepath.Join(repoDir, "TODO.md")); err == nil {
			done := 0
			for _, ln := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(strings.TrimSpace(ln), "- [x]") {
					done++
				}
			}
			if done > archiveNudgeThreshold {
				notes = append(notes, fmt.Sprintf(
					"TODO.md: %d completed `[x]` items (> %d) — archive them to DONE.md to keep the injected file small",
					done, archiveNudgeThreshold))
			}
		}
	}

	// AI attribution: when suppression is configured (enforce.aiAttribution=false)
	// and the Claude adapter is on, .claude/settings.json must carry
	// includeCoAuthoredBy:false — the native lever that drops the co-author trailer /
	// generated-by line. Absent or true is drift; `check --fix` sets it. gyroscope
	// manages the key only under suppression, so an attribution-on repo is never
	// checked for it.
	if cfg.Enforce.Claude && !cfg.Enforce.AIAttribution {
		v, present, err := enforce.CoAuthoredBy(repoDir)
		if err != nil {
			return nil, nil, err
		}
		if !present || v {
			problems = append(problems, ".claude/settings.json: AI attribution suppressed (enforce.aiAttribution=false) but includeCoAuthoredBy is not false (run `gyroscope check --fix`)")
		}
	}

	// Ship-to-team note (soft, never drift): when the Claude adapter is on but the
	// repo gitignores `.claude/`, the SessionStart hook (and, with personas on, the
	// `.claude/agents/` mirror) are version-control-invisible — a teammate who clones
	// gets the tracked docs but no auto-injection and no dispatchable subagents, so
	// the "self-enforcing, ships to the team" promise silently breaks. Advise a
	// targeted `!`-negation. Advisory only: gitignoring `.claude/` is a legitimate
	// choice, so this is a note, not nonconformance.
	if cfg.Enforce.Claude && claudeShipBlocked(repoDir) {
		msg := ".gitignore: `.claude/` is ignored, so the SessionStart hook won't ship to teammates"
		neg := "add `!.claude/settings.json`"
		if cfg.Spokes.Personas == config.PersonaOn {
			msg += " (nor the `.claude/agents/` persona mirror)"
			neg += " and `!.claude/agents/`"
		}
		notes = append(notes, msg+" — "+neg+" after the `.claude/` line to commit the enforcement (keep `.claude/settings.local.json` ignored)")
	}

	return problems, notes, nil
}

// claudeShipBlocked reports whether repoDir's root .gitignore would keep gyroscope's
// Claude enforcement out of version control — it ignores `.claude/` (blanket) with no
// `!`-negation re-including the hook. It is a deliberately small heuristic parse (no
// git dependency): it recognizes a blanket `.claude` / `.claude/` ignore and honors a
// negation for the hook path or for all of `.claude/`. Good enough for a soft note —
// the negation it advises makes the note go away. Nested/edge .gitignore forms may
// slip past; that only costs a missed advisory, never a wrong hard failure.
func claudeShipBlocked(repoDir string) bool {
	b, err := os.ReadFile(filepath.Join(repoDir, ".gitignore"))
	if err != nil {
		return false
	}
	ignores, negated := false, false
	for _, ln := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(ln)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		switch strings.TrimPrefix(t, "/") {
		case ".claude", ".claude/", ".claude/*", ".claude/**":
			ignores = true
		}
		switch t {
		case "!.claude", "!.claude/", "!/.claude", "!/.claude/",
			"!.claude/settings.json", "!/.claude/settings.json":
			negated = true
		}
	}
	return ignores && !negated
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
