package enforce

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/WagnerJust/gyroscope/internal/fsutil"
)

// codexConfigPath is Codex CLI's project-committed config, merged root→cwd.
const codexConfigPath = ".codex/config.toml"

// codexAttributionKey is Codex's kill-switch for the git-commit co-author trailer:
// commit_attribution = "" suppresses the "Co-authored-by: Codex …" line Codex adds
// when it makes commits. gyroscope writes it only to suppress attribution.
const codexAttributionKey = "commit_attribution"
const codexAttributionLine = codexAttributionKey + ` = ""`

// SetCodexCommitAttributionEmpty ensures .codex/config.toml carries a top-level
// `commit_attribution = ""`, preserving all other content. It edits by LINE rather
// than parsing TOML — gyroscope pulls in no TOML dependency — which is safe here
// because it manages exactly one flat top-level key: it replaces an existing
// top-level commit_attribution line, prepends the key when absent (valid before any
// table), or creates the file. Returns changed=false when already empty. Writes are
// atomic (temp + rename), so an interrupted edit can never truncate the user's file.
func SetCodexCommitAttributionEmpty(repoDir string) (changed bool, err error) {
	path := filepath.Join(repoDir, codexConfigPath)
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return true, fsutil.WriteAtomic(repoDir, codexConfigPath, []byte(codexAttributionLine+"\n"))
	}
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(b), "\n")
	idx, val := topLevelKeyLine(lines, codexAttributionKey)
	if idx >= 0 {
		if val == "" {
			return false, nil
		}
		lines[idx] = codexAttributionLine
		return true, fsutil.WriteAtomic(repoDir, codexConfigPath, []byte(strings.Join(lines, "\n")))
	}
	return true, fsutil.WriteAtomic(repoDir, codexConfigPath, []byte(codexAttributionLine+"\n"+string(b)))
}

// CodexCommitAttribution returns the top-level commit_attribution string and whether
// the key is present in .codex/config.toml. A missing file reports present=false.
// Read-only inverse of SetCodexCommitAttributionEmpty.
func CodexCommitAttribution(repoDir string) (value string, present bool, err error) {
	b, err := os.ReadFile(filepath.Join(repoDir, codexConfigPath))
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	idx, val := topLevelKeyLine(strings.Split(string(b), "\n"), codexAttributionKey)
	if idx < 0 {
		return "", false, nil
	}
	return val, true, nil
}

// topLevelKeyLine finds a flat `key = value` line above the first `[table]` header
// (where TOML top-level keys must live) and returns its index and unquoted string
// value, or (-1, "") when absent. Comments and blank lines are skipped; a key inside
// a table is deliberately not matched.
func topLevelKeyLine(lines []string, key string) (int, string) {
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "[") {
			break // first table header: top-level region ends here
		}
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		k, v, ok := strings.Cut(t, "=")
		if !ok {
			continue
		}
		if strings.TrimSpace(k) == key {
			return i, unquoteTOML(strings.TrimSpace(v))
		}
	}
	return -1, ""
}

// unquoteTOML strips one pair of surrounding double quotes from a basic-string
// value; other forms are returned as-is (gyroscope only writes the quoted form).
func unquoteTOML(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
