package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/WagnerJust/gyroscope/internal/config"
	"github.com/WagnerJust/gyroscope/internal/standard"
	"github.com/WagnerJust/gyroscope/internal/target"
)

// convergeState is a per-destination classification of how the on-disk file
// relates to what gyroscope would write. It replaces the old all-or-nothing
// collision check (`existingCollisions`) with a per-file verdict so init can
// apply the safe subset and refuse only genuine conflicts.
type convergeState int

const (
	// stateNew: the destination does not exist yet — a plain create.
	stateNew convergeState = iota
	// stateOK: the destination exists and already matches what gyroscope would
	// write (whole-file equal, or a hub whose managed region is current).
	stateOK
	// stateMerge: the destination exists but is missing gyroscope's managed
	// content — a hub without an up-to-date managed region. The managed region
	// can be injected in place without touching the user's surrounding content.
	stateMerge
	// stateConflict: the destination exists and its content differs from what
	// gyroscope would write, with no managed region to merge into. Overwriting
	// it needs --force.
	stateConflict
)

func (s convergeState) String() string {
	switch s {
	case stateNew:
		return "NEW"
	case stateOK:
		return "OK"
	case stateMerge:
		return "MERGE"
	case stateConflict:
		return "CONFLICT"
	default:
		return "UNKNOWN"
	}
}

// convergeItem pairs a repo-relative destination with its classification and the
// exact bytes gyroscope would write there (the whole-file content for a plain
// file; the full hub for the hub). apply uses Want for a whole-file write and
// re-derives the managed region for a merge.
type convergeItem struct {
	Dest  string
	State convergeState
	Want  []byte
}

// classify inspects every destination init would write under abs — the standard
// files plus the pointer files — and returns a per-file verdict. It reads only;
// it writes nothing. The enforcement adapters (.claude/settings.json, the PI
// extension) are excluded here: they self-classify via their own idempotent
// merge/verify and are never whole-file collisions.
func classify(abs string, files []standard.File) []convergeItem {
	items := make([]convergeItem, 0, len(files)+len(target.All()))
	for _, f := range files {
		items = append(items, classifyOne(abs, f.Dest, f.Content))
	}
	for _, t := range target.All() {
		items = append(items, classifyOne(abs, t.Path, []byte(target.RoutingLine)))
	}
	return items
}

// classifyOne classifies a single destination given the bytes gyroscope would
// write there.
func classifyOne(abs, dest string, want []byte) convergeItem {
	item := convergeItem{Dest: dest, Want: want}
	got, err := os.ReadFile(filepath.Join(abs, dest))
	if err != nil {
		// Absent (or unreadable — treat as absent, the guarded write will fail
		// loudly if it is a real I/O problem): a plain create.
		item.State = stateNew
		return item
	}
	if bytes.Equal(got, want) {
		item.State = stateOK
		return item
	}
	// The hub is the one file with a managed region: if the surrounding
	// (user-owned) content differs but the managed region can be brought current
	// in place, that is a MERGE rather than a CONFLICT.
	if dest == "AGENTS.md" {
		if merged, ok := standard.MergeManaged(got, want); ok {
			if bytes.Equal(merged, got) {
				item.State = stateOK
			} else {
				item.State = stateMerge
			}
			return item
		}
	}
	item.State = stateConflict
	return item
}

// classifyAll loads the plan for cfg and classifies every destination. It is the
// single entry point init/check use to reason about convergence.
func classifyAll(abs string, cfg config.Config) ([]convergeItem, error) {
	files, err := standard.Plan(cfg)
	if err != nil {
		return nil, err
	}
	return classify(abs, files), nil
}

// conflicts returns the destinations classified CONFLICT — the files that need
// --force before init will touch them.
func conflicts(items []convergeItem) []string {
	var out []string
	for _, it := range items {
		if it.State == stateConflict {
			out = append(out, it.Dest)
		}
	}
	return out
}

// preexisting returns every destination that already exists on disk (anything
// not NEW). It is the all-or-nothing collision set the non-force apply refuses on
// until the merge-safe apply (D3) narrows the refusal to genuine conflicts.
func preexisting(items []convergeItem) []string {
	var out []string
	for _, it := range items {
		if it.State != stateNew {
			out = append(out, it.Dest)
		}
	}
	return out
}
