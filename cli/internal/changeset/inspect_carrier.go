package changeset

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// InspectCarrierSuffix marks a change-unit file that records the UP-V inspection:
// the agent's grounded attestation that the code/material deriving from a touched
// fact was inspected against the fact's post-change doc. It is the third arm of a
// change unit, alongside *.patch.md (internal/doc) and *.codemap.md (external/code
// territory). The switch (change apply) refuses without a fresh, covering one.
const InspectCarrierSuffix = ".inspect.md"

// CoveredSource records one change-material file the inspection was made against,
// with a content hash taken at inspection time. Freshness is anchored to the DOC
// material C3 owns (the patch/codemap files) — never to code, which C3 cannot own.
// If a covered file changes after the attestation, its hash no longer matches and
// the inspection is stale.
type CoveredSource struct {
	Source string `yaml:"source"`
	Hash   string `yaml:"hash"`
}

// InspectionRow is one inspected obligation: a Derived Materials / Required
// Verification row of the touched fact, the code territory it was checked in, the
// verdict, and grounded evidence of the inspection.
type InspectionRow struct {
	Obligation string
	Territory  string
	Verdict    string // matches | updated
	Evidence   string
}

// InspectCarrier is one parsed *.inspect.md: the attestation for a single touched
// fact. The gate (cmd) validates coverage-freshness, that every required obligation
// has a verdict, and that evidence is grounded inside the fact's resolved territory.
type InspectCarrier struct {
	Target string          // entity id this inspection attests
	Covers []CoveredSource // the change-material files (+ hashes) inspected against
	Rows   []InspectionRow // one row per inspected obligation
	Source string          // originating file name, for diagnostics
}

type inspectMeta struct {
	Target string          `yaml:"target"`
	Covers []CoveredSource `yaml:"covers"`
}

// MaterialHash is the content hash of a change-material file, used as the freshness
// anchor. Prefixed "sha256:" so the stored form is self-describing.
func MaterialHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ParseInspectCarrier reads one .inspect.md (YAML frontmatter + an Inspections
// table) into an InspectCarrier. A missing target, missing covers, or a table with
// no obligation rows is rejected — an empty attestation would defeat the gate.
func ParseInspectCarrier(source, raw string) (InspectCarrier, error) {
	meta, body, err := splitFrontmatter(raw)
	if err != nil {
		return InspectCarrier{}, fmt.Errorf("inspect %s: %w", source, err)
	}
	var m inspectMeta
	if err := yaml.Unmarshal([]byte(meta), &m); err != nil {
		return InspectCarrier{}, fmt.Errorf("inspect %s: frontmatter: %w", source, err)
	}
	target := strings.TrimSpace(m.Target)
	if target == "" {
		return InspectCarrier{}, fmt.Errorf("inspect %s: missing target", source)
	}
	covers := make([]CoveredSource, 0, len(m.Covers))
	for _, c := range m.Covers {
		c.Source = strings.TrimSpace(c.Source)
		c.Hash = strings.TrimSpace(c.Hash)
		if c.Source == "" || c.Hash == "" {
			return InspectCarrier{}, fmt.Errorf("inspect %s: each covers entry needs source + hash", source)
		}
		covers = append(covers, c)
	}
	if len(covers) == 0 {
		return InspectCarrier{}, fmt.Errorf("inspect %s: covers is empty — an inspection must record the material it was made against (run 'c3x change inspect %s' to stamp it)", source, target)
	}
	rows, err := parseInspectionTable(body)
	if err != nil {
		return InspectCarrier{}, fmt.Errorf("inspect %s: %w", source, err)
	}
	if len(rows) == 0 {
		return InspectCarrier{}, fmt.Errorf("inspect %s: no inspected obligations — the Inspections table must have at least one row", source)
	}
	return InspectCarrier{Target: target, Covers: covers, Rows: rows, Source: source}, nil
}

// parseInspectionTable extracts rows from the carrier's markdown pipe-table. It
// reads the first table it finds (the Inspections table) and maps columns by their
// header names so column order is not load-bearing. Required columns: Obligation,
// Verdict, Evidence; Territory is optional (an out-of-scope obligation may omit it).
func parseInspectionTable(body string) ([]InspectionRow, error) {
	var header []string
	var rows []InspectionRow
	sawSep := false
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if header != nil {
				break // table ended
			}
			continue
		}
		cells := splitTableRow(line)
		if header == nil {
			header = cells
			continue
		}
		if !sawSep && isSeparatorRow(cells) {
			sawSep = true
			continue
		}
		row := InspectionRow{
			Obligation: cell(header, cells, "Obligation"),
			Territory:  cell(header, cells, "Territory"),
			Verdict:    strings.ToLower(cell(header, cells, "Verdict")),
			Evidence:   cell(header, cells, "Evidence"),
		}
		if row.Obligation == "" && row.Verdict == "" && row.Evidence == "" {
			continue
		}
		rows = append(rows, row)
	}
	if header == nil {
		return nil, fmt.Errorf("missing the Inspections table")
	}
	for _, need := range []string{"Obligation", "Verdict", "Evidence"} {
		if !containsHeader(header, need) {
			return nil, fmt.Errorf("Inspections table missing %q column", need)
		}
	}
	return rows, nil
}

func splitTableRow(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.TrimSpace(p)
	}
	return out
}

func isSeparatorRow(cells []string) bool {
	for _, c := range cells {
		if c == "" {
			continue
		}
		if strings.Trim(c, "-: ") != "" {
			return false
		}
	}
	return true
}

func containsHeader(header []string, name string) bool {
	for _, h := range header {
		if strings.EqualFold(h, name) {
			return true
		}
	}
	return false
}

func cell(header, cells []string, name string) string {
	for i, h := range header {
		if strings.EqualFold(h, name) && i < len(cells) {
			return cells[i]
		}
	}
	return ""
}

// ReadInspectDir reads every *.inspect.md in dir, filename order. A missing folder
// yields none (not an error); a malformed carrier is an error.
func ReadInspectDir(dir string) ([]InspectCarrier, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read change folder %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), InspectCarrierSuffix) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	out := make([]InspectCarrier, 0, len(names))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read inspect %s: %w", name, err)
		}
		c, err := ParseInspectCarrier(name, string(data))
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// CoversFresh reports whether this carrier's recorded coverage matches the current
// change material that touches its target. current maps source filename → current
// MaterialHash for every patch/codemap file targeting this fact. The attestation is
// fresh iff it covers exactly those sources with matching hashes — a changed file
// (hash mismatch) or an uncovered new source means the inspection predates the
// material and must be redone. Returns a human-readable reason when not fresh.
func (c InspectCarrier) CoversFresh(current map[string]string) (bool, string) {
	covered := make(map[string]string, len(c.Covers))
	for _, cs := range c.Covers {
		covered[cs.Source] = cs.Hash
	}
	for src, h := range current {
		got, ok := covered[src]
		if !ok {
			return false, fmt.Sprintf("material %s is not covered by the inspection (re-inspect after adding it)", src)
		}
		if got != h {
			return false, fmt.Sprintf("material %s changed since inspection (covered %s, now %s) — re-inspect", src, short(got), short(h))
		}
	}
	for src := range covered {
		if _, ok := current[src]; !ok {
			return false, fmt.Sprintf("inspection covers %s which no longer targets %s — re-inspect", src, c.Target)
		}
	}
	return true, ""
}

func short(hash string) string {
	h := strings.TrimPrefix(hash, "sha256:")
	if len(h) > 10 {
		return h[:10]
	}
	return h
}
