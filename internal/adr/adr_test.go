package adr

import (
	"reflect"
	"strings"
	"testing"
)

// rec is a tiny constructor for a well-formed ADR body carrying both required
// fields, so tests can focus on the one thing they exercise.
func rec(name, extra string) Record {
	return Record{
		Name:    name,
		Content: "# ADR\n\n- **Status:** accepted\n- **Date:** 2026-01-01\n\n" + extra,
	}
}

func TestCheckCleanSetHasNoFindings(t *testing.T) {
	records := []Record{
		{Name: "TEMPLATE.md", Content: "# ADR <NNNN>\n- **Status:** <...>\n- **Date:** <YYYY-MM-DD>\nsuperseded by ADR-XXXX\n"},
		rec("0001-first.md", ""),
		rec("0002-second.md", "builds on ADR-0001.\n"),
		rec("0003-third.md", "see ADR 0002.\n"),
	}
	if got := Check(records); len(got) != 0 {
		t.Fatalf("clean set produced findings: %#v", got)
	}
}

func TestCheckFlagsNumberingGap(t *testing.T) {
	records := []Record{rec("0001-a.md", ""), rec("0003-c.md", "")}
	got := Check(records)
	if !hasSubstr(got, "0002") {
		t.Fatalf("expected a gap finding for 0002, got %#v", got)
	}
}

func TestCheckFlagsDuplicateNumber(t *testing.T) {
	records := []Record{rec("0001-a.md", ""), rec("0002-b.md", ""), rec("0002-c.md", "")}
	got := Check(records)
	if !hasSubstr(got, "duplicat") {
		t.Fatalf("expected a duplicate finding for 0002, got %#v", got)
	}
}

func TestCheckFlagsMissingStatusAndDate(t *testing.T) {
	records := []Record{
		{Name: "0001-a.md", Content: "# ADR\n- **Date:** 2026-01-01\n"},
		{Name: "0002-b.md", Content: "# ADR\n- **Status:** accepted\n"},
	}
	got := Check(records)
	if !(hasSubstr(got, "0001-a.md") && hasSubstr(got, "Status")) {
		t.Fatalf("expected missing-Status finding for 0001, got %#v", got)
	}
	if !(hasSubstr(got, "0002-b.md") && hasSubstr(got, "Date")) {
		t.Fatalf("expected missing-Date finding for 0002, got %#v", got)
	}
}

func TestCheckFlagsDanglingReference(t *testing.T) {
	records := []Record{rec("0001-a.md", ""), rec("0002-b.md", "superseded by ADR-0010.\n")}
	got := Check(records)
	if !(hasSubstr(got, "0002-b.md") && hasSubstr(got, "0010")) {
		t.Fatalf("expected dangling-ref finding pointing at 0010, got %#v", got)
	}
}

func TestCheckIgnoresSelfReferenceAndTemplatePlaceholder(t *testing.T) {
	records := []Record{
		{Name: "TEMPLATE.md", Content: "superseded by ADR-XXXX and ADR-NNNN\n- **Status:** <..>\n- **Date:** <..>\n"},
		rec("0001-a.md", "this is ADR 0001 itself.\n"),
	}
	if got := Check(records); len(got) != 0 {
		t.Fatalf("self-ref and template placeholders must not trip: %#v", got)
	}
}

func TestCheckEmptySetIsClean(t *testing.T) {
	if got := Check(nil); len(got) != 0 {
		t.Fatalf("empty set produced findings: %#v", got)
	}
	// A set that is only the template is also clean (no decisions yet).
	only := []Record{{Name: "TEMPLATE.md", Content: "template\n"}}
	if got := Check(only); len(got) != 0 {
		t.Fatalf("template-only set produced findings: %#v", got)
	}
}

func TestCheckIsDeterministic(t *testing.T) {
	records := []Record{rec("0003-c.md", "refs ADR-0009\n"), rec("0001-a.md", "")}
	a := Check(records)
	b := Check(records)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Check not deterministic:\n%#v\n%#v", a, b)
	}
}

func hasSubstr(lines []string, sub string) bool {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}
