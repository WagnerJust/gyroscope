package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionStringCollapsesWhenUntagged(t *testing.T) {
	// Untagged build: git describe --always == git rev-parse --short HEAD, so the
	// version and commit are the identical sha — collapse to avoid printing it twice.
	if got := versionString("abc123", "abc123", "d"); got != "abc123 (built d)" {
		t.Fatalf("collapsed form wrong: %q", got)
	}
	// Tagged build: distinct version and commit, keep both.
	if got := versionString("1.0.0", "abc", "d"); got != "1.0.0 (commit abc, built d)" {
		t.Fatalf("full form wrong: %q", got)
	}
}

func TestVersionPrintsBuildInfo(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"version"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), version) {
		t.Fatalf("version output missing %q: %s", version, out.String())
	}
}
