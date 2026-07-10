// Package archive moves completed top-level tasks out of TODO.md (which the
// SessionStart hook injects into every session) and into DONE.md (history, never
// injected), so the per-session context stays lean. The move is deterministic and
// lossless: every archived line reappears in DONE.md verbatim and nothing else
// changes. This is the *mechanism* half of the TODO/DONE convention (ADR 0009) —
// the `check` archive nudge detects the backlog, `check --fix` converges it by
// calling here. Pure string transforms; the caller does the file I/O.
package archive

import "strings"

// Plan splits a TODO.md body into that body with completed top-level tasks removed
// (remaining) and the removed blocks (moved). A completed top-level task is a line
// beginning `- [x]` at column zero, together with its indented continuation lines
// (sub-items, wrapped detail) up to the next non-indented line. A nested `[x]` under
// an unfinished top-level parent is left in place: only a top-level `[x]` marks a
// whole task done. moved is nil when nothing qualifies, so the caller can no-op
// without rewriting files, and remaining is then byte-identical to todo.
func Plan(todo string) (remaining string, moved []string) {
	lines := strings.Split(todo, "\n")
	keep := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		if !isTopLevelDone(lines[i]) {
			keep = append(keep, lines[i])
			continue
		}
		block := []string{lines[i]}
		for i+1 < len(lines) && isIndented(lines[i+1]) {
			i++
			block = append(block, lines[i])
		}
		moved = append(moved, strings.Join(block, "\n"))
	}
	if moved == nil {
		return todo, nil
	}
	return strings.Join(keep, "\n"), moved
}

// Merge inserts the moved blocks at the top of done's `## Completed` section
// (newest on top, per the DONE.md convention). When done has no `## Completed`
// heading the blocks are appended under a fresh one. Precondition: len(moved) > 0.
func Merge(done string, moved []string) string {
	block := strings.Join(moved, "\n")
	lines := strings.Split(done, "\n")
	for i, ln := range lines {
		if strings.TrimSpace(ln) == "## Completed" {
			out := make([]string, 0, len(lines)+1)
			out = append(out, lines[:i+1]...)
			out = append(out, block)
			out = append(out, lines[i+1:]...)
			return strings.Join(out, "\n")
		}
	}
	return strings.TrimRight(done, "\n") + "\n\n## Completed\n" + block + "\n"
}

// isTopLevelDone reports whether line is a completed task at column zero.
func isTopLevelDone(line string) bool {
	return strings.HasPrefix(line, "- [x]")
}

// isIndented reports whether line is a continuation of the task above it: any
// non-empty line starting with a space or tab. A blank line ends the block.
func isIndented(line string) bool {
	return line != "" && (line[0] == ' ' || line[0] == '\t')
}
