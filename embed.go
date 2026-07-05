// Package gyroscope embeds the standard templates and the companion skill so the
// binary can write them without keeping duplicate copies in source.
package gyroscope

import "embed"

//go:embed templates
var Templates embed.FS

//go:embed skill/SKILL.md
var SkillMD string
