package standard

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ManagedOpen and ManagedClose bound the region of the hub gyroscope owns.
// Everything between them is gyroscope's to write and re-write; everything
// outside is the user's, and gyroscope never touches it (see ADR 0007). The
// markers are HTML comments so they are invisible in rendered Markdown.
const (
	ManagedOpen  = "<!-- gyroscope:managed -->"
	ManagedClose = "<!-- /gyroscope -->"
)

// ManagedRegion returns the exact bytes between ManagedOpen and ManagedClose in
// doc (excluding the markers themselves), and whether a well-formed managed
// region was found. A missing open marker, a missing close marker, or a close
// that precedes its open all report ok=false.
func ManagedRegion(doc []byte) (region []byte, ok bool) {
	s := string(doc)
	i := strings.Index(s, ManagedOpen)
	if i < 0 {
		return nil, false
	}
	rest := i + len(ManagedOpen)
	j := strings.Index(s[rest:], ManagedClose)
	if j < 0 {
		return nil, false
	}
	return []byte(s[rest : rest+j]), true
}

// hubManaged extracts gyroscope's managed region out of a freshly-rendered hub
// (Plan's AGENTS.md content), which always carries the markers. It is the
// authoritative "what gyroscope wants inside the region" for a merge.
func hubManaged(renderedHub []byte) ([]byte, bool) {
	return ManagedRegion(renderedHub)
}

// MergeManaged returns doc brought current with the managed region of want (the
// freshly-rendered hub), leaving all content outside the markers untouched. It has
// two paths, both of which preserve the user's prose byte-for-byte and perform no
// 3-way content merge (see ADR 0007):
//
//   - doc already has a well-formed managed region → its region is swapped for
//     want's, in place.
//   - doc has NO managed markers (a hand-written hub that predates gyroscope, or
//     one whose markers were never present) → want's full managed region, wrapped
//     in the markers, is APPENDED at EOF. This is D1's MERGE case ("present,
//     missing managed content"): the whole file is the user's, so the safest home
//     for gyroscope's region is after it, disturbing nothing above.
//
// ok is false only when want itself lacks a well-formed managed region (nothing to
// inject).
func MergeManaged(doc, want []byte) (merged []byte, ok bool) {
	wantRegion, ok := hubManaged(want)
	if !ok {
		return nil, false
	}
	s := string(doc)
	i := strings.Index(s, ManagedOpen)
	if i < 0 {
		// No managed markers anywhere: append the wrapped region at EOF, keeping
		// every byte of the user's hub above it. Ensure one blank line of
		// separation so the appended block reads as its own section.
		var buf bytes.Buffer
		buf.WriteString(s)
		if len(s) > 0 && !strings.HasSuffix(s, "\n") {
			buf.WriteByte('\n')
		}
		if !strings.HasSuffix(buf.String(), "\n\n") && len(s) > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(ManagedOpen)
		buf.Write(wantRegion)
		buf.WriteString(ManagedClose)
		buf.WriteByte('\n')
		return buf.Bytes(), true
	}
	openEnd := i + len(ManagedOpen)
	j := strings.Index(s[openEnd:], ManagedClose)
	if j < 0 {
		// An open marker with no close is malformed — not a well-formed region we
		// can safely swap. Treat as un-mergeable (CONFLICT) rather than guess where
		// the region ends.
		return nil, false
	}
	closeStart := openEnd + j
	var buf bytes.Buffer
	buf.WriteString(s[:openEnd])
	buf.Write(wantRegion)
	buf.WriteString(s[closeStart:])
	return buf.Bytes(), true
}

// InjectManaged brings the managed region of the hub at repoDir/AGENTS.md current
// with want (the freshly-rendered hub), preserving all user content outside the
// markers. It is the in-place merge path — the one deliberate exception to routing
// every write through fsutil.WriteGuarded: WriteGuarded is whole-file
// (O_EXCL-or-truncate) and cannot preserve a slice, so this reads the existing
// hub, updates only its managed region (swapping an existing one, or appending a
// wrapped region to a markerless hub — see MergeManaged), and writes the result
// atomically via a sibling temp file + rename so an interrupted merge can never
// truncate the user's hub — a reader sees the old file or the new one, never a
// partial.
//
// It returns an error only when the merge itself is impossible (want carries no
// managed region — a programmer error, since want is the freshly-rendered hub).
func InjectManaged(repoDir string, want []byte) error {
	path := filepath.Join(repoDir, "AGENTS.md")
	doc, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	merged, ok := MergeManaged(doc, want)
	if !ok {
		return fmt.Errorf("%s: rendered hub carries no managed region to inject", "AGENTS.md")
	}
	tmp := path + ".gyroscope.tmp"
	if err := os.WriteFile(tmp, merged, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
