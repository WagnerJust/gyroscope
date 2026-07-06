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
		{cfg.Spokes.Contributing, "templates/CONTRIBUTING.md", "CONTRIBUTING.md"},
		{cfg.Spokes.State, "templates/TODO.md", "TODO.md"},
		{cfg.Spokes.State, "templates/local-todo.md", ".local/todo.md"},
		{cfg.Spokes.Local, "templates/local.md", ".local/local.md"},
		// Process artifacts (enforcement genre — Git/GitHub apply them; not hub-routed).
		{cfg.Spokes.PRTemplate, "templates/.github/pull_request_template.md", ".github/pull_request_template.md"},
		{cfg.Spokes.CommitConvention, "templates/.gitmessage", ".gitmessage"},
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
		if e.dest == "AGENTS.md" {
			b = renderCustomRoutes(b, cfg.Custom)
		}
		files = append(files, File{Dest: e.dest, Content: b})
	}
	// Custom spokes ride as ordinary Files so they flow through Write, the dry-run
	// listing, and the collision check. Skip forgiving on missing Name/Dest.
	for _, c := range cfg.Custom {
		if c.Name == "" || c.Dest == "" {
			continue
		}
		files = append(files, File{Dest: c.Dest, Content: customStub(c.Name)})
	}
	return files, nil
}

// customRouteMarker is the single insertion point in templates/AGENTS.md where
// per-repo custom-spoke routes are rendered. It is a comment (not a `{{...}}`
// fill-once placeholder) so it never trips the skill's fill check or embed_test.
const customRouteMarker = "<!-- gyroscope:custom-routes -->\n"

// renderCustomRoutes replaces the marker line in the hub with one route bullet
// per custom spoke (skipping entries missing a Name or Dest), or removes it
// cleanly when there are none — a single contained replacement, no templating.
func renderCustomRoutes(hub []byte, custom []config.CustomSpoke) []byte {
	var b strings.Builder
	for _, c := range custom {
		if c.Name == "" || c.Dest == "" {
			continue
		}
		fmt.Fprintf(&b, "- **%s** → `%s`\n", c.Name, c.Dest)
	}
	return []byte(strings.Replace(string(hub), customRouteMarker, b.String(), 1))
}

func customStub(name string) []byte {
	return []byte("# " + name + "\n\n" +
		"{{What this spoke covers — one or two lines. gyroscope created the file; the\n" +
		"/gyroscope skill (or you) fills this in.}}\n")
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
