// Package adr polices the structural hygiene of a docs/adr/ set: contiguous
// numbering from 0001, a Status and Date field on every decision record, and no
// reference to an ADR number that does not exist. It checks structure only —
// never "did you write enough ADRs", which is a judgment call left to humans (see
// docs/adr/TEMPLATE.md's own guard). Pure and disk-free so it is unit-testable;
// the caller reads docs/adr/ and hands the filenames + contents in.
package adr

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Record is one file under docs/adr/: its base name and full content.
type Record struct {
	Name    string
	Content string
}

// numbered matches an ADR filename's leading zero-padded number ("0007-...md").
// TEMPLATE.md and any other non-numbered file are simply not decisions.
var numbered = regexp.MustCompile(`^(\d{3,4})-`)

// refPattern finds a reference to another ADR by number in prose — "ADR-0003",
// "ADR 0003", "ADR0003". The `\d` guard means the template's "ADR-XXXX" /
// "ADR-NNNN" placeholders never match, so they are not mistaken for dead links.
var refPattern = regexp.MustCompile(`ADR[\s-]?0*(\d+)`)

// Check returns one human-readable finding per structural problem, in
// deterministic order. TEMPLATE.md is recognized and excluded from the numbered
// set (it is the blank form, not a decision) and from the dangling-reference scan
// (its placeholders are not real references). An empty or template-only set is
// clean.
func Check(records []Record) []string {
	// Collect the numbered decisions: number -> filenames carrying it.
	byNum := map[int][]string{}
	type dec struct {
		name    string
		num     int
		content string
	}
	var decs []dec
	for _, r := range records {
		m := numbered.FindStringSubmatch(r.Name)
		if m == nil {
			continue // TEMPLATE.md or a stray non-ADR file — not a decision.
		}
		n, _ := strconv.Atoi(m[1])
		byNum[n] = append(byNum[n], r.Name)
		decs = append(decs, dec{r.Name, n, r.Content})
	}
	if len(decs) == 0 {
		return nil
	}

	var findings []string

	// Numbering: contiguous from 0001, no duplicates. Report each missing and each
	// duplicated number once, in numeric order.
	max := 0
	nums := make([]int, 0, len(byNum))
	for n := range byNum {
		nums = append(nums, n)
		if n > max {
			max = n
		}
	}
	sort.Ints(nums)
	for n := 1; n <= max; n++ {
		if _, ok := byNum[n]; !ok {
			findings = append(findings, fmt.Sprintf(
				"docs/adr: ADR %04d is missing (numbering must be contiguous from 0001)", n))
		}
	}
	for _, n := range nums {
		if names := byNum[n]; len(names) > 1 {
			sort.Strings(names)
			findings = append(findings, fmt.Sprintf(
				"docs/adr: ADR %04d is duplicated by %v", n, names))
		}
	}

	// Each decision carries a Status and a Date field, and references no ADR number
	// that has no file. Walk decisions in filename order for stable output.
	sort.Slice(decs, func(i, j int) bool { return decs[i].name < decs[j].name })
	for _, d := range decs {
		if !strings.Contains(d.content, "**Status:**") {
			findings = append(findings, fmt.Sprintf("docs/adr/%s: missing **Status:** field", d.name))
		}
		if !strings.Contains(d.content, "**Date:**") {
			findings = append(findings, fmt.Sprintf("docs/adr/%s: missing **Date:** field", d.name))
		}
		seen := map[int]bool{}
		for _, m := range refPattern.FindAllStringSubmatch(d.content, -1) {
			ref, _ := strconv.Atoi(m[1])
			if ref == d.num || seen[ref] {
				continue // a record referring to itself is fine; report each target once.
			}
			seen[ref] = true
			if _, ok := byNum[ref]; !ok {
				findings = append(findings, fmt.Sprintf(
					"docs/adr/%s: references ADR %04d which does not exist", d.name, ref))
			}
		}
	}

	return findings
}
