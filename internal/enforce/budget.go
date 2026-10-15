package enforce

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StateBudget is the byte budget for file content the SessionStart hook injects.
// Claude Code caps hook stdout at 10,000 characters; past that the session gets a
// 2 KB preview and a path it is never told to read. 9000 leaves room for the
// headers the hook prints around the content.
const StateBudget = 9000

// hookTag marks a SessionStart command as gyroscope's own, so Install can
// replace an older version of it instead of appending beside it.
const hookTag = "# gyroscope-hook"

// SessionStartCommand builds the one-line POSIX sh hook that injects the state
// files in order within StateBudget: a readable file that fits the remaining
// budget is printed under a header and decrements it; one that doesn't is named
// with its size and left for the agent to read; a missing file is skipped. The
// trailing `true` forces exit 0 so the harness never reports a failed hook.
// PlanInjection is the Go simulation of the same loop.
func SessionStartCommand(state []string) string {
	return fmt.Sprintf(`b=%d; printf '%%s\n' '## Session state (gyroscope): working files for resuming, injected within a %d-byte budget. Repo rules live in AGENTS.md and the spokes it routes to, not here.'; `+
		`for f in %s; do [ -r "$f" ] || continue; s=$(wc -c <"$f" | tr -d ' '); `+
		`if [ "$s" -le "$b" ]; then printf '\n### %%s (%%s bytes)\n' "$f" "$s"; cat "$f"; b=$((b-s)); `+
		`else printf '\n### %%s: %%s bytes, over the remaining session budget; not injected. Read it when resuming.\n' "$f" "$s"; fi; done; true `+hookTag,
		StateBudget, StateBudget, strings.Join(state, " "))
}

// FileSize is a repo-relative path and its size in bytes.
type FileSize struct {
	Path  string
	Bytes int
}

// PlanInjection simulates SessionStartCommand's loop over repoDir: in order, a
// readable file whose size fits the remaining budget is injected and decrements
// it, a readable file that doesn't fit is deferred, and a missing or unreadable
// file is skipped. check uses it to report what a session actually receives.
func PlanInjection(repoDir string, state []string, budget int) (injected, deferred []FileSize) {
	for _, rel := range state {
		f, err := os.Open(filepath.Join(repoDir, rel))
		if err != nil {
			continue // [ -r "$f" ] || continue
		}
		info, err := f.Stat()
		f.Close()
		if err != nil {
			continue
		}
		fs := FileSize{Path: rel, Bytes: int(info.Size())}
		if fs.Bytes <= budget {
			injected = append(injected, fs)
			budget -= fs.Bytes
		} else {
			deferred = append(deferred, fs)
		}
	}
	return injected, deferred
}
