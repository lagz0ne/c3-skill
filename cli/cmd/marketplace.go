package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/marketplace"
)

// MarketplaceOptions holds parameters for marketplace subcommands.
type MarketplaceOptions struct {
	BaseDir    string // override for ~/.c3/marketplace/
	URL        string // git URL for add
	SourceName string // filter for list, target for remove/update
	Tag        string // filter for list
	RuleID     string // target for show
	JSON       bool
}

// RunMarketplaceAdd clones a marketplace repo and registers it.
func RunMarketplaceAdd(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	if opts.URL == "" {
		return fmt.Errorf("error: usage: c3x marketplace add <github-url>")
	}

	// Clone to temp, read manifest, then move to final location
	tmpDir := filepath.Join(baseDir, ".tmp-clone")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	if err := marketplace.Clone(opts.URL, tmpDir); err != nil {
		return fmt.Errorf("error: cloning %s: %w", opts.URL, err)
	}

	manifestPath := filepath.Join(tmpDir, "marketplace.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("error: no marketplace.yaml found in repo\nhint: create a marketplace.yaml with name, description, and rules")
	}

	manifest, err := marketplace.ParseManifest(data)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	// Validate rule files exist
	for _, r := range manifest.Rules {
		ruleFile := filepath.Join(tmpDir, r.ID+".md")
		if _, err := os.Stat(ruleFile); os.IsNotExist(err) {
			return fmt.Errorf("error: manifest lists %s but %s.md not found in repo", r.ID, r.ID)
		}
	}

	cacheDir := reg.CacheDir(manifest.Name)
	os.RemoveAll(cacheDir)
	if err := os.Rename(tmpDir, cacheDir); err != nil {
		return fmt.Errorf("error: moving clone to cache: %w", err)
	}

	if err := reg.Add(marketplace.Source{Name: manifest.Name, URL: opts.URL}); err != nil {
		os.RemoveAll(cacheDir) // rollback on registration failure
		return fmt.Errorf("error: %w", err)
	}

	fmt.Fprintf(w, "Added: %s (%d rules from %s)\n", manifest.Name, len(manifest.Rules), opts.URL)
	return nil
}

// MarketplaceListResult is JSON output from list.
type MarketplaceListResult struct {
	Sources []MarketplaceSourceResult `json:"sources"`
}

// MarketplaceSourceResult is one source in list output.
type MarketplaceSourceResult struct {
	Name        string                  `json:"name"`
	URL         string                  `json:"url"`
	Description string                  `json:"description"`
	Tags        []string                `json:"tags,omitempty"`
	Rules       []marketplace.RuleEntry `json:"rules"`
}

// RunMarketplaceList lists available rules across all sources.
func RunMarketplaceList(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	sources, err := reg.List()
	if err != nil {
		return err
	}

	var result MarketplaceListResult

	for _, src := range sources {
		if opts.SourceName != "" && src.Name != opts.SourceName {
			continue
		}

		manifestPath := filepath.Join(reg.CacheDir(src.Name), "marketplace.yaml")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		manifest, err := marketplace.ParseManifest(data)
		if err != nil {
			continue
		}

		entry := MarketplaceSourceResult{
			Name:        src.Name,
			URL:         src.URL,
			Description: manifest.Description,
			Tags:        manifest.Tags,
		}

		for _, r := range manifest.Rules {
			if opts.Tag != "" && !containsTag(r.Tags, opts.Tag) && !containsTag(manifest.Tags, opts.Tag) {
				continue
			}
			entry.Rules = append(entry.Rules, r)
		}

		if len(entry.Rules) > 0 || opts.Tag == "" {
			result.Sources = append(result.Sources, entry)
		}
	}

	if opts.JSON {
		return writeJSON(w, result)
	}

	if len(result.Sources) == 0 {
		fmt.Fprintln(w, "No marketplace sources registered.")
		fmt.Fprintln(w, "hint: c3x marketplace add <github-url>")
		return nil
	}

	for _, src := range result.Sources {
		fmt.Fprintf(w, "\n%s (%s)\n", src.Name, src.URL)
		if src.Description != "" {
			fmt.Fprintf(w, "  %s\n", src.Description)
		}
		for _, r := range src.Rules {
			fmt.Fprintf(w, "  - %s: %s\n", r.ID, r.Summary)
		}
	}
	return nil
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// RunMarketplaceShow prints the full content of a marketplace rule.
func RunMarketplaceShow(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	if opts.RuleID == "" {
		return fmt.Errorf("error: usage: c3x marketplace show <rule-id>")
	}

	sources, err := reg.List()
	if err != nil {
		return err
	}

	for _, src := range sources {
		if opts.SourceName != "" && src.Name != opts.SourceName {
			continue
		}
		rulePath := filepath.Join(reg.CacheDir(src.Name), opts.RuleID+".md")
		data, err := os.ReadFile(rulePath)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "# Source: %s\n\n", src.Name)
		w.Write(data)
		return nil
	}

	return fmt.Errorf("error: %s not found in any registered source\nhint: c3x marketplace list", opts.RuleID)
}

// RunMarketplaceUpdate pulls latest from one or all sources.
func RunMarketplaceUpdate(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	sources, err := reg.List()
	if err != nil {
		return err
	}

	if len(sources) == 0 {
		fmt.Fprintln(w, "No marketplace sources registered.")
		return nil
	}

	var errs []string
	for _, src := range sources {
		if opts.SourceName != "" && src.Name != opts.SourceName {
			continue
		}
		cacheDir := reg.CacheDir(src.Name)
		if err := marketplace.Pull(cacheDir); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", src.Name, err))
			continue
		}
		reg.UpdateFetched(src.Name)
		fmt.Fprintf(w, "Updated: %s\n", src.Name)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// RunMarketplaceRemove unregisters a source and deletes its cache.
func RunMarketplaceRemove(opts MarketplaceOptions, w io.Writer) error {
	baseDir := resolveBaseDir(opts.BaseDir)
	reg := marketplace.NewRegistry(baseDir)

	if opts.SourceName == "" {
		return fmt.Errorf("error: usage: c3x marketplace remove <source-name>")
	}

	cacheDir := reg.CacheDir(opts.SourceName)
	os.RemoveAll(cacheDir)

	if err := reg.Remove(opts.SourceName); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	fmt.Fprintf(w, "Removed: %s\n", opts.SourceName)
	return nil
}

func resolveBaseDir(override string) string {
	if override != "" {
		return override
	}
	return marketplace.DefaultBaseDir()
}
