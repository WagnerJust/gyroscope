package archive

import (
	"reflect"
	"strings"
	"testing"
)

func TestPlanMovesTopLevelDoneAndKeepsOpen(t *testing.T) {
	todo := "# TODO\n\n## Next\n- [ ] open one\n- [x] done one\n- [~] in flight\n"
	remaining, moved := Plan(todo)

	if want := []string{"- [x] done one"}; !reflect.DeepEqual(moved, want) {
		t.Fatalf("moved = %q, want %q", moved, want)
	}
	if strings.Contains(remaining, "done one") {
		t.Fatalf("remaining still contains the archived item:\n%s", remaining)
	}
	for _, keep := range []string{"open one", "in flight"} {
		if !strings.Contains(remaining, keep) {
			t.Fatalf("remaining dropped an open item %q:\n%s", keep, remaining)
		}
	}
}

func TestPlanMovesDoneItemWithIndentedSubtree(t *testing.T) {
	todo := "## Next\n" +
		"- [x] parent done\n" +
		"  - [x] sub a\n" +
		"  - [x] sub b\n" +
		"- [ ] next open\n"
	remaining, moved := Plan(todo)

	want := []string{"- [x] parent done\n  - [x] sub a\n  - [x] sub b"}
	if !reflect.DeepEqual(moved, want) {
		t.Fatalf("moved = %q, want the whole subtree as one block %q", moved, want)
	}
	if strings.Contains(remaining, "sub a") || strings.Contains(remaining, "parent done") {
		t.Fatalf("remaining still holds the moved subtree:\n%s", remaining)
	}
	if !strings.Contains(remaining, "- [ ] next open") {
		t.Fatalf("remaining dropped the following open item:\n%s", remaining)
	}
}

// A nested `[x]` under an unfinished top-level parent is NOT a completed task — the
// task as a whole is still open — so it stays put. Only a top-level `- [x]` moves.
func TestPlanLeavesNestedDoneUnderOpenParent(t *testing.T) {
	todo := "## Next\n- [~] parent in flight\n  - [x] sub done\n  - [ ] sub open\n"
	remaining, moved := Plan(todo)

	if moved != nil {
		t.Fatalf("nothing top-level is done; moved should be nil, got %q", moved)
	}
	if remaining != todo {
		t.Fatalf("remaining should be byte-identical when nothing moves:\n%q\n!=\n%q", remaining, todo)
	}
}

func TestPlanNoDoneItemsIsNoOp(t *testing.T) {
	todo := "# TODO\n- [ ] a\n- [~] b\n"
	remaining, moved := Plan(todo)
	if moved != nil {
		t.Fatalf("moved should be nil, got %q", moved)
	}
	if remaining != todo {
		t.Fatalf("remaining should equal input, got %q", remaining)
	}
}

func TestMergeInsertsAtTopOfCompletedSection(t *testing.T) {
	done := "# DONE\n\n**Legend:** `[x]` done\n\n## Completed\n- [x] old one\n"
	got := Merge(done, []string{"- [x] new one", "- [x] newer two"})

	// Newest batch sits directly under the heading, above the existing history.
	want := "# DONE\n\n**Legend:** `[x]` done\n\n## Completed\n- [x] new one\n- [x] newer two\n- [x] old one\n"
	if got != want {
		t.Fatalf("Merge =\n%q\nwant\n%q", got, want)
	}
}

func TestMergeCreatesSectionWhenAbsent(t *testing.T) {
	done := "# DONE\n\nsome preamble\n"
	got := Merge(done, []string{"- [x] moved"})
	if !strings.Contains(got, "## Completed\n- [x] moved") {
		t.Fatalf("Merge should append a `## Completed` section, got:\n%s", got)
	}
	if !strings.Contains(got, "some preamble") {
		t.Fatalf("Merge must preserve existing DONE.md content, got:\n%s", got)
	}
}
