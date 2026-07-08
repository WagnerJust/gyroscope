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
	want := []byte(ManagedOpen + "\nbody\n" + ManagedClose)
	if _, ok := MergeManaged([]byte("a plain file, no markers"), want); ok {
		t.Fatal("a doc with no managed region cannot be merged")
	}
	if _, ok := MergeManaged([]byte(ManagedOpen+"\nx\n"+ManagedClose), []byte("want has no region")); ok {
		t.Fatal("a want with no managed region cannot drive a merge")
	}
}
