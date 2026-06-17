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

// GuardCanonicalMutation refuses direct edits to frozen facts. Codemap updates
// are exempt because they are stored outside the sealed canonical document.
func GuardCanonicalMutation(s *store.Store, c3Dir string, opts Options) error {
	switch opts.Command {
	case "write", "wire", "delete":
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
