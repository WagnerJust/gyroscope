package gyroscope

import (
	"bytes"
	"io/fs"
	"testing"
)

// TestADRTemplateHasNoFillOncePlaceholders locks the marker convention: the ADR
// TEMPLATE.md is a blank form whose fields use `<...>` and are filled per-ADR, so it
// must contain zero `{{` (which is reserved for fill-once scaffold placeholders the
// skill replaces). If the collision ever silently returns, this test fails.
func TestADRTemplateHasNoFillOncePlaceholders(t *testing.T) {
	const path = "templates/docs/adr/TEMPLATE.md"

	data, err := fs.ReadFile(Templates, path)
	if err != nil {
		t.Fatalf("read %s from embedded Templates: %v", path, err)
	}

	if bytes.Contains(data, []byte("{{")) {
		t.Errorf("%s must not contain `{{` — it is a blank form using `<...>` form fields, "+
			"not fill-once `{{...}}` scaffold placeholders", path)
	}
}
