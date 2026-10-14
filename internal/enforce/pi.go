// The PI adapter enforces the hub for the Pi Agent Harness. PI reads AGENTS.md
// natively once the project is trusted, so gyroscope does not re-inject the hub;
// instead it installs a project-local extension that, on session_start, injects
// the *non-hub* spokes (state/resume files) the way Claude's SessionStart hook does. The
// extension is a managed file — gyroscope owns it and overwrites it on apply.
package enforce

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WagnerJust/gyroscope/internal/fsutil"
)

// piExtensionPath is the auto-discovered, project-local extension PI loads.
const piExtensionPath = ".pi/extensions/gyroscope-context.ts"

// piExtensionTemplate is the managed extension. %s is the JSON array of injected
// paths. It reads each file live on session_start and injects the concatenation as
// a non-triggering next-turn message, so the state spokes ride the user's first
// prompt — the analog of Claude's SessionStart hook.
const piExtensionTemplate = `// gyroscope-context.ts — managed by gyroscope; do not edit by hand.
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { readFileSync } from "node:fs";

const PATHS: string[] = %s;

export default function (pi: ExtensionAPI) {
	pi.on("session_start", async () => {
		const parts: string[] = [];
		for (const p of PATHS) {
			try {
				parts.push("# " + p + "\n\n" + readFileSync(p, "utf8"));
			} catch {
				// missing spoke: skip, mirroring cat ... 2>/dev/null
			}
		}
		if (parts.length === 0) return;
		pi.sendMessage(
			{ customType: "gyroscope-context", content: parts.join("\n\n"), display: false },
			{ deliverAs: "nextTurn" },
		);
	});
}
`

// PI is the Pi Agent Harness enforcement adapter.
type PI struct{}

func (PI) ID() string { return "pi" }

// piInjectPaths is the rules minus the hub (PI reads AGENTS.md natively), then
// the state, in that order.
func piInjectPaths(p Paths) []string {
	return append(withoutHub(p.Rules), p.State...)
}

// renderPIExtension produces the extension source for p.
func renderPIExtension(p Paths) string {
	arr, _ := json.Marshal(piInjectPaths(p)) // []string always marshals
	return fmt.Sprintf(piExtensionTemplate, string(arr))
}

func (PI) PlanLine(p Paths) string {
	return "write: " + piExtensionPath + " (PI session_start context injection)"
}

func (PI) Apply(repoDir string, p Paths) (bool, error) {
	want := renderPIExtension(p)
	dest := filepath.Join(repoDir, piExtensionPath)
	if b, err := os.ReadFile(dest); err == nil {
		if string(b) == want {
			return false, nil
		}
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := fsutil.WriteGuarded(repoDir, piExtensionPath, []byte(want), true); err != nil {
		return false, err
	}
	return true, nil
}

func (PI) Verify(repoDir string, p Paths) (bool, error) {
	want := renderPIExtension(p)
	b, err := os.ReadFile(filepath.Join(repoDir, piExtensionPath))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return string(b) == want, nil
}
