package persona

import (
	"encoding/base64"
	"strings"
)

// syncScriptHeader is the fixed preamble of the generated sync script. The `w`
// helper writes a file only if absent (skip-if-exists), so the script is idempotent
// — running it on a fresh clone reconstructs the missing registrations, and running
// it again is a no-op. Dependency-free POSIX sh; the only external command is
// base64 (coreutils, universally present), which sidesteps shell-escaping the
// mirror bytes and lets a single generator handle every harness's format.
const syncScriptHeader = `#!/bin/sh
# gyroscope-sync.sh — managed by gyroscope; do not edit by hand.
# Reconstructs gyroscope's generated agent registrations for a teammate who cloned a
# repo that gitignores the generated dirs (.claude/, .opencode/, .codex/, .pi/) and
# never installed gyroscope — so they still get dispatchable subagents. Idempotent:
# each file is written only if absent (skip-if-exists), so re-running is safe.
set -eu

w() { # w <dest> <base64>
	[ -f "$1" ] && return 0
	mkdir -p "$(dirname "$1")"
	printf '%s' "$2" | base64 -d > "$1"
}

`

// SyncScript bakes an idempotent POSIX-sh script that reconstructs each mirror's
// bytes on a fresh clone — the "reconstruct per-dev" alternative to committing the
// generated dirs. It bakes the fully-rendered bytes (base64) rather than a copy
// list, because the OpenCode/Codex/PI mirrors are format re-encodes bash cannot
// reproduce without re-implementing gyroscope's encoders (dual-implementation
// drift). Deterministic: mirrors arrive in a stable order and base64 is fixed, so
// `check` can verify the script byte-for-byte.
func SyncScript(mirrors []Mirror) []byte {
	var b strings.Builder
	b.WriteString(syncScriptHeader)
	for _, m := range mirrors {
		b.WriteString("w ")
		b.WriteString(shSingleQuote(m.Dest))
		b.WriteByte(' ')
		b.WriteString(shSingleQuote(base64.StdEncoding.EncodeToString(m.Bytes)))
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

// shSingleQuote wraps s in POSIX single quotes, escaping any embedded quote. The
// base64 alphabet and gyroscope-controlled dest paths contain no single quotes, so
// this is defensive; it keeps the two arguments intact regardless of their bytes.
func shSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
