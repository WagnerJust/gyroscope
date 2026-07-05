// Package standard plans and writes gyroscope's opinionated doc standard into a
// repo. The hub (AGENTS.md) is always written; other spokes are gated by config.
// Content comes verbatim from the embedded templates (scaffolds with placeholders
// the companion skill fills after). Writing never clobbers an existing file
// unless force is set — the O_EXCL open is the guarantee, mirroring buckle.
package standard

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	gyroscope "github.com/WagnerJust/gyroscope"
	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/fsutil"
)

type File struct {
	Dest    string
	Content []byte
}

func Plan(cfg config.Config) ([]File, error) {
	type entry struct {
		on         bool
		tmpl, dest string
	}
	entries := []entry{
		{true, "templates/AGENTS.md", "AGENTS.md"},
		{cfg.Spokes.Context, "templates/CONTEXT.md", "CONTEXT.md"},
		{cfg.Spokes.Agents, "templates/docs/agents.md", "docs/agents.md"},
		{cfg.Spokes.ADR, "templates/docs/adr/TEMPLATE.md", "docs/adr/TEMPLATE.md"},
		{cfg.Spokes.ADR, "templates/docs/adr/0001-record-architecture-decisions.md", "docs/adr/0001-record-architecture-decisions.md"},
		{cfg.Spokes.Personas, "templates/docs/agents/README.md", "docs/agents/README.md"},
		{cfg.Spokes.Local, "templates/local.md", ".local/local.md"},
	}
	var files []File
	for _, e := range entries {
		if !e.on {
			continue
		}
		b, err := fs.ReadFile(gyroscope.Templates, e.tmpl)
		if err != nil {
			return nil, fmt.Errorf("read template %s: %w", e.tmpl, err)
		}
		files = append(files, File{Dest: e.dest, Content: b})
	}
	return files, nil
}

// Write creates each planned file under repoDir, refusing to clobber an existing
// file unless force is set, and appends ".local/" to .gitignore when a local-spoke
// file is written. On error it returns early: written lists the files already
// created — there is no rollback of partial progress.
func Write(repoDir string, files []File, force bool) (written []string, err error) {
	wroteLocal := false
	for _, f := range files {
		if err := fsutil.WriteGuarded(repoDir, f.Dest, f.Content, force); err != nil {
			return written, err
		}
		written = append(written, f.Dest)
		if strings.HasPrefix(f.Dest, ".local/") {
			wroteLocal = true
		}
	}
	if wroteLocal {
		if err := ensureGitignored(repoDir, ".local/"); err != nil {
			return written, err
		}
	}
	return written, nil
}

func ensureGitignored(repoDir, pattern string) error {
	path := filepath.Join(repoDir, ".gitignore")
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == pattern {
			return nil
		}
	}
	content := string(b)
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += pattern + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}
