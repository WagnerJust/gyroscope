package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	gyroscope "github.com/WagnerJust/gyroscope"
)

// lookBinary resolves an executable on PATH; it is a package var so tests can stub
// the lookup. Defaults to exec.LookPath.
var lookBinary = exec.LookPath

// newInstallSkillCmd installs gyroscope's own embedded skill into a skills dir
// (default ~/.claude/skills), the way buckle installs buckle's skill.
func newInstallSkillCmd(stdout io.Writer) *cobra.Command {
	var apply bool
	cmd := &cobra.Command{
		Use:   "install-skill [skills-dir]",
		Short: "Install gyroscope's companion skill into your agent's skills dir (default ~/.claude/skills).",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var base string
			if len(args) == 1 {
				base = args[0]
			} else {
				home, err := os.UserHomeDir()
				if err != nil {
					return errCannotRun(err)
				}
				base = filepath.Join(home, ".claude", "skills")
			}
			dest := filepath.Join(base, "gyroscope", "SKILL.md")
			if !apply {
				fmt.Fprintf(stdout, "gyroscope install-skill (dry-run)\n  path: %s\n  bytes: %d\n", dest, len(gyroscope.SkillMD))
				warnIfBinaryMissing(stdout)
				fmt.Fprintln(stdout, "\nRe-run with --apply.")
				return nil
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return errCannotRun(err)
			}
			if err := os.WriteFile(dest, []byte(gyroscope.SkillMD), 0o644); err != nil {
				return errCannotRun(err)
			}
			fmt.Fprintf(stdout, "installed skill → %s\n", dest)
			// The skill's install step shells out to the `gyroscope` binary. If it
			// is not resolvable on PATH, that step fails silently later; warn now
			// with an install instruction so the failure mode is explicit (D5).
			warnIfBinaryMissing(stdout)
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&apply, "apply", false, "Actually write (default is dry-run).")
	return cmd
}

// warnIfBinaryMissing prints a resolvability note about the `gyroscope` binary.
// The skill and the binary are two halves of one tool: the skill runs the
// interview, then shells to the binary to write the standard. If the binary is not
// on PATH the skill's step 2 fails with an opaque "command not found"; surfacing it
// here — with an install instruction — makes the failure mode explicit up front.
// It never fails the command: the skill is installed regardless.
func warnIfBinaryMissing(stdout io.Writer) {
	if path, err := lookBinary("gyroscope"); err == nil {
		fmt.Fprintf(stdout, "binary: gyroscope resolves at %s\n", path)
		return
	}
	fmt.Fprintln(stdout, "\nwarning: `gyroscope` is not on your PATH.")
	fmt.Fprintln(stdout, "  The skill shells out to the binary to write the standard, so its")
	fmt.Fprintln(stdout, "  install step will fail until the binary is resolvable. Install it with:")
	fmt.Fprintln(stdout, "    go install github.com/WagnerJust/gyroscope/cmd/gyroscope@latest")
	fmt.Fprintln(stdout, "  and ensure your Go bin dir (e.g. $(go env GOPATH)/bin) is on PATH.")
}
