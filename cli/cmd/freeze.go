package cmd

import (
	"fmt"

	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// FactIsFrozen reports whether a direct mutation of entityType must be refused.
// Facts are apply-only — they change only through a change-unit. Change-docs
// (the authoring surface) and canvases (the contract) are not facts.
func FactIsFrozen(c3Dir, entityType string) bool {
	if entityType == "canvas" {
		return false
	}
	return !schema.IsChangeDocDir(c3Dir, entityType)
}

// GuardFactMutation is the CLI-level freeze: a direct fact-mutation command
// (write/set/wire/delete on an existing fact) refuses here, naming the only legal
// path. The internal write machinery is left intact so change apply can reuse it.
// A missing id (creation, or a not-yet-existing target) is left to the command.
func GuardFactMutation(s *store.Store, c3Dir, id string) error {
	if id == "" {
		return nil
	}
	e, err := s.GetEntity(id)
	if err != nil {
		return nil
	}
	if FactIsFrozen(c3Dir, e.Type) {
		return frozenFactError(id)
	}
	return nil
}

// inCreationWindow reports whether id is a frozen-TYPE fact that has never been
// authored: zero body nodes AND never versioned (Version 0). Such a fact has no
// sealed body to protect, so its FIRST body authoring (a `write`) is allowed; the
// moment it carries a body it is frozen like any other fact. The init-seeded system
// context doc c3-0 is the canonical case — without this it is born frozen-and-empty
// with no authorable path at all (write/set refused, no citable anchor to insert).
//
// Deliberately narrow, because the freeze protects authored content and must not be
// bypassable:
//   - Empty RootMerkle ALONE is not "never authored": a legacy migration or a
//     frontmatter-only import can leave RootMerkle empty while frontmatter/edges
//     were authored. So we additionally require Version 0 and zero body nodes.
//   - Only `write` (body authoring) opens the window. `set`/`wire`/`delete` stay
//     frozen even on a bodyless fact — they touch frontmatter / edges / existence,
//     which may be authored independently of the body.
//   - Once a body exists (nodes > 0) or the fact has been versioned, the window is
//     shut; a fact re-emptied by an internal repair path (Version already > 0) does
//     not re-open it.
func inCreationWindow(s *store.Store, c3Dir, id string) bool {
	e, err := s.GetEntity(id)
	if err != nil {
		return false
	}
	if !FactIsFrozen(c3Dir, e.Type) {
		return false // not a fact — the normal create/edit path applies
	}
	if e.Version != 0 {
		return false // already versioned → authored, frozen
	}
	nodes, err := s.NodesForEntity(id)
	if err != nil || len(nodes) > 0 {
		return false // has a body → frozen
	}
	return true
}

// GuardCanonicalMutation refuses direct edits to frozen facts. Two carve-outs:
// codemap updates (stored outside the sealed canonical document) and the creation
// window — the first `write` that authors a never-authored fact's body.
func GuardCanonicalMutation(s *store.Store, c3Dir string, opts Options) error {
	switch opts.Command {
	case "write":
		if len(opts.Args) >= 1 {
			if inCreationWindow(s, c3Dir, opts.Args[0]) {
				return nil
			}
			return GuardFactMutation(s, c3Dir, opts.Args[0])
		}
	case "wire", "delete":
		if len(opts.Args) >= 1 {
			return GuardFactMutation(s, c3Dir, opts.Args[0])
		}
	case "set":
		if len(opts.Args) >= 1 {
			_, field, _ := ResolveSetArgs(opts)
			if field == "codemap" {
				return nil
			}
			return GuardFactMutation(s, c3Dir, opts.Args[0])
		}
	}
	return nil
}

// frozenFactError names the only legal path to change a fact.
func frozenFactError(id string) error {
	return fmt.Errorf("error: %s is a fact — facts are frozen and change only through a change-unit\nhint: author patches in .c3/changes/<unit-id>/ then run 'c3x change apply <unit-id>'", id)
}
