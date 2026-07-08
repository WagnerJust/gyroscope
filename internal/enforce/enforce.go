package enforce

// Adapter installs (and verifies) one harness's mechanism for force-feeding the
// enabled spokes into a session at startup. paths are the repo-relative files
// hookPathsFor selected; each adapter renders them into its own mechanism.
type Adapter interface {
	ID() string
	// PlanLine is a one-line dry-run description of what Apply would do.
	PlanLine(paths []string) string
	// Apply installs or updates the mechanism; changed is false when it was
	// already present and current.
	Apply(repoDir string, paths []string) (changed bool, err error)
	// Verify reports whether the mechanism is installed and current. A missing
	// mechanism is (false, nil); only an inspection failure returns a non-nil error.
	Verify(repoDir string, paths []string) (installed bool, err error)
}

// Compile-time proof that the concrete adapters satisfy Adapter.
var (
	_ Adapter = Claude{}
	_ Adapter = PI{}
)
