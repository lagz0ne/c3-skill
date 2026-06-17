package changeset

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PatchSuffix is the file suffix that marks a change-material patch file inside
// a change-unit's folder.
const PatchSuffix = ".patch.md"

// ReadPatchDir reads every *.patch.md file in dir, in filename order, parsing
// each into a Patch. A missing folder yields no patches (not an error) — a
// change-unit may carry no material yet. A malformed patch file is an error.
func ReadPatchDir(dir string) ([]Patch, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read patch folder %s: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), PatchSuffix) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	patches := make([]Patch, 0, len(names))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read patch %s: %w", name, err)
		}
		p, err := ParsePatch(name, string(data))
		if err != nil {
			return nil, err
		}
		patches = append(patches, p)
	}
	return patches, nil
}
