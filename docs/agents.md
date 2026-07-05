# Agent instructions — gyroscope

Applies to all work in this repo.

## Build & test

- Build the binary with `make build` (threads version/commit/date via `-ldflags`
  into `bin/gyroscope`), or `go build ./cmd/gyroscope` for a quick local binary.
- Run `go test ./...` before claiming a change is done; use `make test-race`
  (race detector) for anything touching concurrent or filesystem writes.
- `go vet ./...` and `gofmt` must be clean — CI gates on both.

## Conventions

- Go 1.24. `github.com/spf13/cobra` is the **only** direct dependency; everything
  else is stdlib (JSON via `encoding/json`, embedding via `//go:embed`). Keep it
  that way — dependency-light is a deliberate inheritance from buckle.
- TDD: write the failing test first, then the implementation (see the per-task
  loop in the MVP plan). Every `internal/*` package has one responsibility and a
  matching `_test.go`.
- Minimal change / YAGNI: build what the current task needs, not a framework for
  imagined ones.
- All file writes go through `internal/fsutil.WriteGuarded`: `O_EXCL` refuses to
  clobber an existing file unless `force` is set. This is the "never destroy the
  user's work" guarantee — do not bypass it with a raw `os.WriteFile`.
- `.claude/settings.json` is merged, not overwritten: preserve existing keys, and
  encode with `SetEscapeHTML(false)` so a shell fragment like `2>/dev/null` is
  written literally rather than HTML-escaped. Write via temp-file + rename so an
  interrupted write can never truncate a user's settings.
- The binary guarantees **structure + hook**; the skill supplies **content**.
  The binary writes scaffolds verbatim and does NOT template — there is no
  rendering engine. Placeholder filling is the skill's job.

## Do NOT

- Do NOT add a dependency beyond cobra. Every new module weakens the
  dependency-light promise and the `go install`-clean distribution; reach for the
  stdlib first.
- Do NOT commit build artifacts — `bin/` and `dist/` are gitignored because they
  drift from source and bloat diffs. Ship source; let `make`/goreleaser produce
  binaries.
- Do NOT HTML-escape when writing `.claude/settings.json`; the default
  `json.Marshal` turns `>` into `>` and silently breaks the hook command.
- Do NOT put a templating engine in the binary. Content comes from the skill's
  interview, not from the binary substituting variables — keeping the writer
  deterministic and CI-safe is the whole point of the binary/skill split.
- Do NOT hand-roll clobber-guard logic in a new writer; call
  `fsutil.WriteGuarded` so the refuse-overwrite behavior stays in one place.
