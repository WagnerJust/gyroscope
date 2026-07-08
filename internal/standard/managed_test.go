package standard

import (
	"bytes"
	"testing"
)

func TestManagedRegionExtracts(t *testing.T) {
	doc := []byte("before\n" + ManagedOpen + "\nowned\n" + ManagedClose + "\nafter\n")
	region, ok := ManagedRegion(doc)
	if !ok {
		t.Fatal("expected a well-formed managed region")
	}
	if string(region) != "\nowned\n" {
		t.Fatalf("region = %q, want %q", region, "\nowned\n")
	}
}

func TestManagedRegionMissingMarkers(t *testing.T) {
	if _, ok := ManagedRegion([]byte("no markers here")); ok {
		t.Fatal("no markers → ok should be false")
	}
	if _, ok := ManagedRegion([]byte(ManagedOpen + " but no close")); ok {
		t.Fatal("missing close → ok should be false")
	}
}

func TestMergeManagedReplacesOnlyTheRegion(t *testing.T) {
	// doc: user prose around a stale managed region.
	doc := []byte("# User heading\n\nkeep me\n\n" +
		ManagedOpen + "\nstale\n" + ManagedClose + "\n\ntrailing user text\n")
	// want: a freshly-rendered hub with its own managed region.
	want := []byte("# AGENTS.md\n\n" + ManagedOpen + "\nfresh managed body\n" + ManagedClose + "\n")

	merged, ok := MergeManaged(doc, want)
	if !ok {
		t.Fatal("expected a successful merge")
	}
	s := string(merged)
	// User content outside the markers is preserved byte-for-byte.
	if !bytes.Contains(merged, []byte("# User heading\n\nkeep me")) {
		t.Fatalf("user preamble not preserved:\n%s", s)
	}
	if !bytes.Contains(merged, []byte("trailing user text")) {
		t.Fatalf("user trailer not preserved:\n%s", s)
	}
	// The managed body is now the rendered hub's.
	if !bytes.Contains(merged, []byte(ManagedOpen+"\nfresh managed body\n"+ManagedClose)) {
		t.Fatalf("managed region not updated:\n%s", s)
	}
	if bytes.Contains(merged, []byte("stale")) {
		t.Fatalf("stale managed content should be gone:\n%s", s)
	}
}

func TestMergeManagedFailsWithoutRegion(t *testing.T) {
	// A doc with no markers is no longer un-mergeable: MergeManaged appends the
	// managed region at EOF (see TestMergeManagedAppendsRegionToMarkerlessDoc).
	// It still fails when the *want* carries no region to inject.
	want := []byte(ManagedOpen + "\nbody\n" + ManagedClose)
	if _, ok := MergeManaged([]byte(ManagedOpen+"\nx\n"+ManagedClose), []byte("want has no region")); ok {
		t.Fatal("a want with no managed region cannot drive a merge")
	}
	_ = want
}

func TestMergeManagedAppendsRegionToMarkerlessDoc(t *testing.T) {
	// A hand-written hub that predates gyroscope has no managed markers at all.
	// That is exactly D1's MERGE case ("present, missing managed content"): the
	// region is appended at EOF, preserving every byte of the user's content.
	doc := []byte("# notwhoop hub\n\nBefore touching the band, read this.\n")
	want := []byte("# AGENTS.md\n\n" + ManagedOpen + "\nfresh managed body\n" + ManagedClose + "\n")

	merged, ok := MergeManaged(doc, want)
	if !ok {
		t.Fatal("a markerless doc should merge by appending the region")
	}
	s := string(merged)
	// All original content survives.
	if !bytes.Contains(merged, []byte("Before touching the band, read this.")) {
		t.Fatalf("user content must be preserved:\n%s", s)
	}
	// The managed region is now present (appended at EOF).
	if !bytes.Contains(merged, []byte(ManagedOpen+"\nfresh managed body\n"+ManagedClose)) {
		t.Fatalf("managed region must be appended:\n%s", s)
	}
	// The region comes after the user's content, not before it.
	if bytes.Index(merged, []byte("Before touching the band")) > bytes.Index(merged, []byte(ManagedOpen)) {
		t.Fatalf("user content must precede the appended region:\n%s", s)
	}
	// The appended region round-trips: re-merging with the same want is a no-op.
	again, ok := MergeManaged(merged, want)
	if !ok {
		t.Fatal("re-merge of an already-merged doc should succeed")
	}
	if !bytes.Equal(again, merged) {
		t.Fatalf("re-merge must be idempotent:\n%s", again)
	}
}
