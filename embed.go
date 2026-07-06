// Package gyroscope embeds the standard templates and the companion skill so the
// binary can write them without keeping duplicate copies in source.
package gyroscope

import "embed"

// The all: prefix keeps dotfile templates (`.gitmessage`, `.github/…`) in the
// embedded tree — a bare `//go:embed templates` silently drops names starting
// with '.' or '_'.
//
//go:embed all:templates
var Templates embed.FS

//go:embed skill/SKILL.md
var SkillMD string
