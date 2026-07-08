package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/fsutil"
)

// newAgentsCmd builds `gyroscope agents`, whose only subcommand `set` mutates the
// personas lifecycle state in gyroscope.json. The skill calls this rather than
// hand-editing JSON, so config integrity stays in the binary.
func newAgentsCmd(stdout io.Writer) *cobra.Command {
	agents := &cobra.Command{
		Use:   "agents",
		Short: "Manage the agent-personas spoke.",
	}
	set := &cobra.Command{
		Use:   "set <unknown|on|skipped|off> [repo-path]",
		Short: "Set the personas lifecycle state in gyroscope.json.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(_ *cobra.Command, args []string) error {
			state := config.PersonaState(args[0])
			if !state.Valid() {
				return errCannotRun(fmt.Errorf("unknown state %q (want unknown|on|skipped|off)", args[0]))
			}
			repo := "."
			if len(args) == 2 {
				repo = args[1]
			}
			abs, err := filepath.Abs(repo)
			if err != nil {
				return errCannotRun(err)
			}
			cfg, err := config.Load(abs)
			if err != nil {
				return errCannotRun(err)
			}
			cfg.Spokes.Personas = state
			b, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return errCannotRun(err)
			}
			b = append(b, '\n')
			// force: this command deliberately mutates an existing gyroscope.json.
			if err := fsutil.WriteGuarded(abs, "gyroscope.json", b, true); err != nil {
				return errCannotRun(err)
			}
			fmt.Fprintf(stdout, "personas: %s\n", state)
			return nil
		},
	}
	agents.AddCommand(set)
	return agents
}
