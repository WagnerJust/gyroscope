package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	gyroscope "github.com/WagnerJust/gyroscope"
)

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
				fmt.Fprintf(stdout, "gyroscope install-skill (dry-run)\n  path: %s\n  bytes: %d\n\nRe-run with --apply.\n", dest, len(gyroscope.SkillMD))
				return nil
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return errCannotRun(err)
			}
			if err := os.WriteFile(dest, []byte(gyroscope.SkillMD), 0o644); err != nil {
				return errCannotRun(err)
			}
			fmt.Fprintf(stdout, "installed skill → %s\n", dest)
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&apply, "apply", false, "Actually write (default is dry-run).")
	return cmd
}
