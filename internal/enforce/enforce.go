package enforce

// Paths splits what an adapter delivers into two channels. Rules are the static
// repo rules (the hub and the agents spoke); State is the dynamic working state a
// session resumes from (gyroscope.json, TODO.md, the .local/ spokes). Each
// adapter routes the two through whatever its harness offers: Claude Code loads
// the rules via CLAUDE.md's @imports, so its hook carries State only.
type Paths struct {
	Rules []string
	State []string
}

// Adapter installs (and verifies) one harness's mechanism for force-feeding the
// enabled spokes into a session at startup. p is the repo-relative file split
// injectPathsFor selected; each adapter renders it into its own mechanism.
type Adapter interface {
	ID() string
	// PlanLine is a one-line dry-run description of what Apply would do.
	PlanLine(p Paths) string
	// Apply installs or updates the mechanism; changed is false when it was
	// already present and current.
	Apply(repoDir string, p Paths) (changed bool, err error)
	// Verify reports whether the mechanism is installed and current. A missing
	// mechanism is (false, nil); only an inspection failure returns a non-nil error.
	Verify(repoDir string, p Paths) (installed bool, err error)
}

// withoutHub returns paths minus AGENTS.md, which every harness but Claude Code
// reads natively.
func withoutHub(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p != "AGENTS.md" {
			out = append(out, p)
		}
	}
	return out
}

// Compile-time proof that the concrete adapters satisfy Adapter.
var (
	_ Adapter = Claude{}
	_ Adapter = PI{}
	_ Adapter = Cursor{}
	_ Adapter = OpenCode{}
)
