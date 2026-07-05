package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionPrintsBuildInfo(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"version"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), version) {
		t.Fatalf("version output missing %q: %s", version, out.String())
	}
}
