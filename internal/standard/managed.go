package standard

import (
	"bytes"
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

// MergeManaged returns doc with its managed region replaced by the managed region
// of want (the freshly-rendered hub), leaving all content outside the markers
// untouched. ok is false when either doc or want lacks a well-formed managed
// region — the caller then treats the file as an un-mergeable whole (CONFLICT).
//
// This is the in-place merge path: it re-writes only the bytes gyroscope owns and
// preserves the user's surrounding prose byte-for-byte. It performs no 3-way
// content merge — it swaps one managed region for another (see ADR 0007).
func MergeManaged(doc, want []byte) (merged []byte, ok bool) {
	wantRegion, ok := hubManaged(want)
	if !ok {
		return nil, false
	}
	s := string(doc)
	i := strings.Index(s, ManagedOpen)
	if i < 0 {
		return nil, false
	}
	openEnd := i + len(ManagedOpen)
	j := strings.Index(s[openEnd:], ManagedClose)
	if j < 0 {
		return nil, false
	}
	closeStart := openEnd + j
	var buf bytes.Buffer
	buf.WriteString(s[:openEnd])
	buf.Write(wantRegion)
	buf.WriteString(s[closeStart:])
	return buf.Bytes(), true
}
