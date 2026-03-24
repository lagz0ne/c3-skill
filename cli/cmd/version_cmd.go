package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// VersionOptions holds parameters for the version command.
type VersionOptions struct {
	Store    *store.Store
	EntityID string
	Version  int
}

// RunVersion outputs the content at a specific version.
func RunVersion(opts VersionOptions, w io.Writer) error {
	if opts.EntityID == "" {
		return fmt.Errorf("usage: c3x version <entity-id> <n>")
	}

	v, err := opts.Store.GetVersion(opts.EntityID, opts.Version)
	if err != nil {
		return fmt.Errorf("version %d of %q not found", opts.Version, opts.EntityID)
	}

	fmt.Fprint(w, v.Content)
	return nil
}
