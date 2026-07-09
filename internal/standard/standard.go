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
		{cfg.Spokes.Personas.Enabled(), "templates/docs/agents/README.md", "docs/agents/README.md"},
		{cfg.Spokes.Contributing, "templates/CONTRIBUTING.md", "CONTRIBUTING.md"},
		{cfg.Spokes.State, "templates/TODO.md", "TODO.md"},
		// DONE.md is the completed-work archive — enforced + hub-routed under the
		// same State toggle as TODO.md, but never catted by the hook (see ADR 0009).
		{cfg.Spokes.State, "templates/DONE.md", "DONE.md"},
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
			b = renderRoutes(b, cfg)
			b = renderPersonasDirective(b, cfg)
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

// routesMarker is the single insertion point in templates/AGENTS.md where the
// hub's route list is rendered. The binary emits a route only for enabled spokes
// (plus any custom spokes), so the hub never points at a file that wasn't
// written — an agent following a route never hits a missing spoke. This is a
// contained assembly of fixed route strings gated by config, not a templating
// engine (see ADR 0003).
const routesMarker = "<!-- gyroscope:routes -->"

// renderRoutes replaces the routes marker in the hub with the route block for
// cfg: one bullet per enabled built-in spoke, in canonical order, followed by one
// per custom spoke. Disabled spokes are simply absent.
func renderRoutes(hub []byte, cfg config.Config) []byte {
	return []byte(strings.Replace(string(hub), routesMarker, Routes(cfg), 1))
}

// personasMarker is where the hub's standing personas directive is spliced in.
const personasMarker = "<!-- gyroscope:personas-directive -->"

// PersonasDirective returns the hub's standing instruction about the personas
// lifecycle when the spoke is enabled, else "". It is the single source of truth
// for the directive text, shared by renderPersonasDirective and by
// `gyroscope check`, which compares it against the on-disk hub.
func PersonasDirective(cfg config.Config) string {
	if !cfg.Spokes.Personas.Enabled() {
		return ""
	}
	return "## Personas\n\n" +
		"If `gyroscope.json` `spokes.personas` is `unknown`, then before starting other work, " +
		"ask the user whether to wire agent personas for this repo (run `/gyroscope`) or skip for now. " +
		"When the state is `on`, `skipped`, or `off`, do nothing about this."
}

// renderPersonasDirective replaces personasMarker in the hub with the directive
// for cfg. When the spoke is off the marker (and its preceding blank line) is
// removed cleanly.
func renderPersonasDirective(hub []byte, cfg config.Config) []byte {
	d := PersonasDirective(cfg)
	s := string(hub)
	if d == "" {
		s = strings.Replace(s, "\n\n"+personasMarker, "", 1)
		s = strings.Replace(s, personasMarker, "", 1)
		return []byte(s)
	}
	return []byte(strings.Replace(s, personasMarker, d, 1))
}

// Routes returns the hub's route block for cfg — the newline-joined bullets
// renderRoutes splices into the hub. It is the single source of truth for the
// route strings, shared by init (via renderRoutes) and by `gyroscope check`,
// which compares it against the on-disk hub rather than re-deriving the bullets.
func Routes(cfg config.Config) string {
	return strings.Join(routeLines(cfg), "\n")
}

// routeLines builds one hub route bullet per enabled built-in spoke, in canonical
// order, followed by one per custom spoke (skipping entries missing a Name or
// Dest).
func routeLines(cfg config.Config) []string {
	builtins := []struct {
		on   bool
		line string
	}{
		{cfg.Spokes.Context, "- **Naming things / writing prose** → read `CONTEXT.md` first for the canonical vocabulary."},
		{cfg.Spokes.Agents, "- **Build, test, conventions** → `docs/agents.md`."},
		{cfg.Spokes.State, "- **Where work stands — in flight / next (resume here)** → `TODO.md` (repo-wide, open work only); `.local/todo.md` holds your personal, gitignored state."},
		{cfg.Spokes.State, "- **Completed work / history** → `DONE.md` (archive; not injected — move a task's line here from `TODO.md` when it's done)."},
		{cfg.Spokes.Contributing, "- **How changes get proposed & reviewed here** → `CONTRIBUTING.md`."},
		{cfg.Spokes.Local, "- **Your** personal setup / stack (may differ from repo defaults) → `.local/local.md` (gitignored; may not exist)."},
		{cfg.Spokes.ADR, "- **Why the code is shaped this way** → `docs/adr/` (architecture decisions)."},
		{cfg.Spokes.Personas.Enabled(), "- **Specialized agent personas for this repo** → `docs/agents/`."},
	}
	var lines []string
	for _, b := range builtins {
		if b.on {
			lines = append(lines, b.line)
		}
	}
	for _, c := range cfg.Custom {
		if c.Name == "" || c.Dest == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- **%s** → `%s`", c.Name, c.Dest))
	}
	return lines
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

// EnsureLocalGitignore appends ".local/" to repoDir/.gitignore when it is not
// already listed, the same mutation Write performs when a .local/ spoke is
// written. The merge-safe apply calls it directly, since it drives writes
// per-file rather than through Write.
func EnsureLocalGitignore(repoDir string) error {
	return ensureGitignored(repoDir, ".local/")
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
