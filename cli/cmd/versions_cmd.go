package cmd

import (
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// VersionsOptions holds parameters for the versions command.
type VersionsOptions struct {
	Store    *store.Store
	EntityID string
	JSON     bool
}

// VersionRow is the JSON representation of a version entry.
type VersionRow struct {
	Version    int    `json:"version"`
	RootMerkle string `json:"root_merkle"`
	CommitHash string `json:"commit_hash,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// RunVersions lists version history for an entity.
func RunVersions(opts VersionsOptions, w io.Writer) error {
	if opts.EntityID == "" {
		return fmt.Errorf("usage: c3x versions <entity-id>")
	}

	if _, err := opts.Store.GetEntity(opts.EntityID); err != nil {
		return fmt.Errorf("entity %q not found", opts.EntityID)
	}

	versions, err := opts.Store.ListVersions(opts.EntityID)
	if err != nil {
		return fmt.Errorf("listing versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Fprintln(w, "No versions for "+opts.EntityID)
		return nil
	}

	if opts.JSON {
		rows := make([]VersionRow, len(versions))
		for i, v := range versions {
			rows[i] = VersionRow{
				Version:    v.Version,
				RootMerkle: v.RootMerkle,
				CommitHash: v.CommitHash,
				CreatedAt:  v.CreatedAt,
			}
		}
		return writeJSON(w, rows)
	}

	fmt.Fprintf(w, "%-8s %-12s %-10s %s\n", "VERSION", "MERKLE", "COMMIT", "CREATED_AT")
	for _, v := range versions {
		merkle := v.RootMerkle
		if len(merkle) > 10 {
			merkle = merkle[:10]
		}
		commit := v.CommitHash
		if commit == "" {
			commit = "-"
		} else if len(commit) > 8 {
			commit = commit[:8]
		}
		fmt.Fprintf(w, "%-8d %-12s %-10s %s\n", v.Version, merkle, commit, v.CreatedAt)
	}
	return nil
}
