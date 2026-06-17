package changeset

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// CodemapCarrierSuffix marks a change-unit file that declares a fact's external
// code bindings (its code-map globs) as part of the unit's footprint.
const CodemapCarrierSuffix = ".codemap.md"

// CodemapChange is one declared external-binding change: the full glob set a fact
// should map to after the work. It is the EXTERNAL (right-V) counterpart to a Patch.
// Unlike a patch it is never sealed and never drift-frozen — code C3 cannot own —
// so it carries no result hash: it is applied to the live code_map table and the
// check-time introspection (not a freeze gate) verifies the globs actually resolve.
// Base is optional and informational, recording the pre-change globs for the
// two-arm view; it does not gate apply.
type CodemapChange struct {
	Target string   // entity id whose codemap is (re)declared
	Base   []string // optional: the globs before this change, for the derive→match view
	Globs  []string // the declared post-state glob set (full replace)
	Source string   // originating file name, for diagnostics
}

type codemapMeta struct {
	Target string   `yaml:"target"`
	Base   []string `yaml:"base"`
}

// ParseCodemapCarrier reads one .codemap.md file (YAML frontmatter + body) into a
// CodemapChange. The body lists the declared post-state globs, one per line; blank
// lines and '#'-comments are ignored. Globs are de-duplicated (the code_map table
// keys on (entity, pattern), so a repeat would fail the insert). An EMPTY body is
// rejected: a carrier applies a full-replace, so a blank one would silently clear
// the fact's bindings — almost always a mistake, and the introspection would not
// flag it. To intentionally drop bindings, edit the code-map directly.
func ParseCodemapCarrier(source, raw string) (CodemapChange, error) {
	meta, body, err := splitFrontmatter(raw)
	if err != nil {
		return CodemapChange{}, fmt.Errorf("codemap %s: %w", source, err)
	}
	var m codemapMeta
	if err := yaml.Unmarshal([]byte(meta), &m); err != nil {
		return CodemapChange{}, fmt.Errorf("codemap %s: frontmatter: %w", source, err)
	}
	target := strings.TrimSpace(m.Target)
	if target == "" {
		return CodemapChange{}, fmt.Errorf("codemap %s: missing target", source)
	}
	var globs []string
	seen := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if seen[line] {
			continue
		}
		seen[line] = true
		globs = append(globs, line)
	}
	if len(globs) == 0 {
		return CodemapChange{}, fmt.Errorf("codemap %s: no globs declared — a carrier must list at least one pattern (it full-replaces the target's code-map; an empty one would silently clear it)", source)
	}
	base := make([]string, 0, len(m.Base))
	for _, b := range m.Base {
		if b = strings.TrimSpace(b); b != "" {
			base = append(base, b)
		}
	}
	return CodemapChange{Target: target, Base: base, Globs: globs, Source: source}, nil
}

// ReadCodemapDir reads every *.codemap.md file in dir, in filename order, parsing
// each into a CodemapChange. A missing folder yields none (not an error). A
// malformed carrier is an error. This is the discovery half that keeps a
// carrier-only change-unit from silently doing nothing at apply.
func ReadCodemapDir(dir string) ([]CodemapChange, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read change folder %s: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), CodemapCarrierSuffix) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	changes := make([]CodemapChange, 0, len(names))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read codemap %s: %w", name, err)
		}
		c, err := ParseCodemapCarrier(name, string(data))
		if err != nil {
			return nil, err
		}
		changes = append(changes, c)
	}
	return changes, nil
}
