package enforce

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

var defaultState = []string{"gyroscope.json", "TODO.md", ".local/todo.md", ".local/local.md"}

// writeSized writes rel under dir with exactly n bytes of content: a marker line
// naming the file (so the test can find it in the hook's output), then filler.
func writeSized(t *testing.T, dir, rel string, n int) []byte {
	t.Helper()
	head := "<<" + rel + ">>\n"
	if n < len(head) {
		t.Fatalf("size %d too small for %s", n, rel)
	}
	b := append([]byte(head), bytes.Repeat([]byte("x"), n-len(head))...)
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, rel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, rel), b, 0o644); err != nil {
		t.Fatal(err)
	}
	return b
}

// runHook runs the generated SessionStart command in dir the way Claude Code
// does (a shell, cwd at the repo root) and returns its stdout.
func runHook(t *testing.T, dir string, state []string) string {
	t.Helper()
	cmd := exec.Command("sh", "-c", SessionStartCommand(state))
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("hook exited non-zero: %v (stderr: %s)", err, stderr.String())
	}
	if stderr.Len() > 0 {
		t.Fatalf("hook wrote to stderr: %s", stderr.String())
	}
	return string(out)
}

var (
	injectedHeader = regexp.MustCompile(`(?m)^### (\S+) \((\d+) bytes\)$`)
	deferredHeader = regexp.MustCompile(`(?m)^### (\S+): (\d+) bytes, over the remaining session budget; not injected\. Read it when resuming\.$`)
)

// parseHook recovers the injected and deferred sets from the hook's stdout, so
// they can be compared with PlanInjection's simulation.
func parseHook(t *testing.T, out string) (injected, deferred []FileSize) {
	t.Helper()
	for _, m := range injectedHeader.FindAllStringSubmatch(out, -1) {
		n, _ := strconv.Atoi(m[2])
		injected = append(injected, FileSize{Path: m[1], Bytes: n})
	}
	for _, m := range deferredHeader.FindAllStringSubmatch(out, -1) {
		n, _ := strconv.Atoi(m[2])
		deferred = append(deferred, FileSize{Path: m[1], Bytes: n})
	}
	return injected, deferred
}

func TestHookInjectsWithinBudgetE2E(t *testing.T) {
	cases := []struct {
		name         string
		sizes        map[string]int // missing key = file absent
		wantInjected []string
		wantDeferred []string
	}{
		{
			name:         "last file over the remaining budget is deferred",
			sizes:        map[string]int{"gyroscope.json": 300, "TODO.md": 5000, ".local/todo.md": 3000, ".local/local.md": 2000},
			wantInjected: []string{"gyroscope.json", "TODO.md", ".local/todo.md"},
			wantDeferred: []string{".local/local.md"},
		},
		{
			name:         "oversized TODO.md is deferred, everything after it still fits",
			sizes:        map[string]int{"gyroscope.json": 300, "TODO.md": 20000, ".local/todo.md": 3000, ".local/local.md": 2000},
			wantInjected: []string{"gyroscope.json", ".local/todo.md", ".local/local.md"},
			wantDeferred: []string{"TODO.md"},
		},
		{
			name:         "missing .local files are skipped silently",
			sizes:        map[string]int{"gyroscope.json": 300, "TODO.md": 5000},
			wantInjected: []string{"gyroscope.json", "TODO.md"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			contents := map[string][]byte{}
			for rel, n := range tc.sizes {
				contents[rel] = writeSized(t, dir, rel, n)
			}

			out := runHook(t, dir, defaultState)
			if n := utf8.RuneCountInString(out); n >= 10000 {
				t.Fatalf("hook stdout is %d characters, must stay under the harness's 10,000 cap", n)
			}
			if !strings.HasPrefix(out, "## Session state (gyroscope):") {
				t.Fatalf("stdout should open with the session-state header, got %.120q", out)
			}

			gotInj, gotDef := parseHook(t, out)
			names := func(fs []FileSize) []string {
				var s []string
				for _, f := range fs {
					s = append(s, f.Path)
				}
				return s
			}
			if !reflect.DeepEqual(names(gotInj), tc.wantInjected) {
				t.Fatalf("injected = %v, want %v\n%s", names(gotInj), tc.wantInjected, out)
			}
			if !reflect.DeepEqual(names(gotDef), tc.wantDeferred) {
				t.Fatalf("deferred = %v, want %v", names(gotDef), tc.wantDeferred)
			}

			for _, rel := range tc.wantInjected {
				header := "### " + rel + " (" + strconv.Itoa(tc.sizes[rel]) + " bytes)\n"
				if !strings.Contains(out, header+string(contents[rel])) {
					t.Errorf("%s: header + full contents not found in stdout", rel)
				}
			}
			for _, rel := range tc.wantDeferred {
				if strings.Contains(out, "<<"+rel+">>") {
					t.Errorf("%s: deferred but its contents were injected", rel)
				}
				if !strings.Contains(out, rel+": "+strconv.Itoa(tc.sizes[rel])+" bytes, over the remaining session budget") {
					t.Errorf("%s: deferred line should name its size", rel)
				}
			}
			for _, rel := range defaultState {
				if _, ok := tc.sizes[rel]; !ok && strings.Contains(out, rel) {
					t.Errorf("%s: missing file should be skipped silently, but stdout mentions it", rel)
				}
			}

			// The Go simulation check relies on agrees with what the shell did.
			planInj, planDef := PlanInjection(dir, defaultState, StateBudget)
			if !reflect.DeepEqual(planInj, gotInj) || !reflect.DeepEqual(planDef, gotDef) {
				t.Fatalf("PlanInjection = (%v, %v), hook did (%v, %v)", planInj, planDef, gotInj, gotDef)
			}
		})
	}
}

func TestPlanInjection(t *testing.T) {
	dir := t.TempDir()
	writeSized(t, dir, "gyroscope.json", 300)
	writeSized(t, dir, "TODO.md", 8800)
	writeSized(t, dir, ".local/local.md", 200)

	inj, def := PlanInjection(dir, defaultState, StateBudget)
	wantInj := []FileSize{{"gyroscope.json", 300}, {".local/local.md", 200}}
	wantDef := []FileSize{{"TODO.md", 8800}}
	if !reflect.DeepEqual(inj, wantInj) || !reflect.DeepEqual(def, wantDef) {
		t.Fatalf("got (%v, %v), want (%v, %v)", inj, def, wantInj, wantDef)
	}

	// A file exactly the size of the remaining budget fits (-le, not -lt).
	inj, def = PlanInjection(dir, []string{"TODO.md"}, 8800)
	if len(inj) != 1 || len(def) != 0 {
		t.Fatalf("a file equal to the budget should be injected, got (%v, %v)", inj, def)
	}
}
