package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/config"
	"github.com/lagz0ne/c3-design/cli/internal/store"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

var version = "dev"

var (
	reAddID   = regexp.MustCompile(`\(id: ([^)]+)\)`)
	reAddPath = regexp.MustCompile(`Created: (\S+)`)
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run contains all CLI logic, returning errors instead of calling os.Exit.
func run(argv []string, w io.Writer) error {
	opts := cmd.ParseArgs(argv)

	if opts.Version {
		fmt.Fprintln(w, version)
		return nil
	}

	if opts.Help || opts.Command == "" {
		cmd.ShowHelp(opts.Command, w)
		return nil
	}

	// init is special — creates .c3/ with DB, no store needed
	if opts.Command == "init" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error: cannot get working directory: %w", err)
		}
		c3Dir := filepath.Join(cwd, ".c3")
		projectName := filepath.Base(cwd)
		return cmd.RunInitDB(c3Dir, projectName, w)
	}

	// capabilities is special — describes the CLI itself, no .c3/ needed
	if opts.Command == "capabilities" {
		cmd.ShowCapabilities(w)
		return nil
	}

	// marketplace is special — uses ~/.c3/marketplace/, no .c3/ needed
	if opts.Command == "marketplace" {
		return runMarketplace(opts, w)
	}

	// All other commands need a .c3/ directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error: cannot get working directory: %w", err)
	}
	c3Dir := config.ResolveC3Dir(cwd, opts.C3Dir)
	if c3Dir == "" {
		return fmt.Errorf("error: No .c3/ directory found\nhint: run 'c3x init' to create one, or use --c3-dir <path>")
	}

	// migrate works on legacy files — special path, before DB detection
	if opts.Command == "migrate" {
		if opts.DryRun {
			return cmd.RunMigrateDryRun(c3Dir, opts.JSON, w)
		}
		return cmd.RunMigrate(c3Dir, opts.KeepOriginals, w)
	}

	// Detect format: DB vs legacy files
	dbPath := filepath.Join(c3Dir, "c3.db")
	hasDB := fileExists(dbPath)

	// Legacy format detected — block all commands except check
	if !hasDB && hasMarkdownFiles(c3Dir) {
		if opts.Command == "check" {
			return runLegacyCheck(c3Dir, opts, w)
		}
		return fmt.Errorf("error: .c3/ contains markdown files but no database (c3.db)\n\nThis version of c3x uses an embedded database.\nUse /c3 in Claude Code to run an LLM-assisted migration that\nvalidates and fixes malformed docs before importing.\n\nOr if docs are already valid: c3x migrate")
	}

	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("error: opening database: %w", err)
	}
	defer s.Close()

	return runCommand(opts, s, c3Dir, w)
}

// runMarketplace handles the marketplace subcommands.
func runMarketplace(opts cmd.Options, w io.Writer) error {
	subCmd := ""
	if len(opts.Args) >= 1 {
		subCmd = opts.Args[0]
	}
	mOpts := cmd.MarketplaceOptions{
		JSON: opts.JSON,
		Tag:  opts.Tag,
	}
	if len(opts.Args) >= 2 {
		switch subCmd {
		case "add":
			mOpts.URL = opts.Args[1]
		case "show":
			mOpts.RuleID = opts.Args[1]
		case "remove", "update":
			mOpts.SourceName = opts.Args[1]
		}
	}
	if opts.Source != "" {
		mOpts.SourceName = opts.Source
	}

	switch subCmd {
	case "add":
		return cmd.RunMarketplaceAdd(mOpts, w)
	case "list":
		return cmd.RunMarketplaceList(mOpts, w)
	case "show":
		return cmd.RunMarketplaceShow(mOpts, w)
	case "update":
		return cmd.RunMarketplaceUpdate(mOpts, w)
	case "remove":
		return cmd.RunMarketplaceRemove(mOpts, w)
	default:
		cmd.ShowHelp("marketplace", w)
		return nil
	}
}

// runCommand dispatches to the appropriate command handler.
func runCommand(opts cmd.Options, s *store.Store, c3Dir string, w io.Writer) error {
	projectDir := config.ProjectDir(c3Dir)

	switch opts.Command {
	case "list":
		return cmd.RunList(cmd.ListOptions{Store: s, JSON: opts.JSON, Flat: opts.Flat, Compact: opts.Compact, C3Dir: c3Dir, IncludeADR: opts.IncludeADR}, w)
	case "check":
		return cmd.RunCheckV2(cmd.CheckOptions{
			Store:      s,
			JSON:       opts.JSON,
			ProjectDir: projectDir,
			C3Dir:      c3Dir,
			IncludeADR: opts.IncludeADR,
			Fix:        opts.Fix,
		}, w)
	case "read":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		return cmd.RunRead(cmd.ReadOptions{Store: s, ID: entityID, JSON: opts.JSON}, w)
	case "write":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("error: no input on stdin\nhint: pipe content: echo '...' | c3x write <id>, or: c3x read <id> | c3x write <id>")
		}
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error: reading stdin: %w", err)
		}
		return cmd.RunWrite(cmd.WriteOptions{Store: s, ID: entityID, Section: opts.Section, Content: string(content)}, w)
	case "add":
		return runAdd(opts, s, c3Dir, w)
	case "set":
		return runSet(opts, s, c3Dir, w)
	case "wire", "unwire":
		return runWire(opts, s, w)
	case "lookup":
		if len(opts.Args) < 1 {
			return fmt.Errorf("error: lookup requires a <file-path> argument\nhint: run 'c3x lookup --help' for usage")
		}
		return cmd.RunLookup(cmd.LookupOptions{
			Store:      s,
			FilePath:   opts.Args[0],
			JSON:       opts.JSON,
			ProjectDir: projectDir,
			C3Dir:      c3Dir,
		}, w)
	case "codemap":
		return cmd.RunCodemap(cmd.CodemapOptions{Store: s, JSON: opts.JSON}, w)
	case "coverage":
		return cmd.RunCoverage(cmd.CoverageOptions{Store: s, C3Dir: c3Dir, ProjectDir: projectDir, JSON: opts.JSON}, w)
	case "schema":
		entityType := ""
		if len(opts.Args) >= 1 {
			entityType = opts.Args[0]
		}
		return cmd.RunSchema(entityType, opts.JSON, w)
	case "graph":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		if entityID == "" {
			return fmt.Errorf("error: graph requires an <entity-id> argument\nhint: run 'c3x graph --help' for usage")
		}
		return cmd.RunGraph(cmd.GraphOptions{
			Store: s, EntityID: entityID, Depth: opts.Depth,
			Direction: opts.Direction, Format: opts.Format,
			JSON: opts.JSON, C3Dir: c3Dir,
		}, w)
	case "delete":
		id := ""
		if len(opts.Args) >= 1 {
			id = opts.Args[0]
		}
		return cmd.RunDelete(cmd.DeleteOptions{C3Dir: c3Dir, ID: id, Store: s, DryRun: opts.DryRun}, w)
	case "query":
		queryTerm := ""
		if len(opts.Args) >= 1 {
			queryTerm = opts.Args[0]
		}
		return cmd.RunQuery(cmd.QueryOptions{Store: s, Query: queryTerm, TypeFilter: opts.TypeFilter, Limit: opts.Limit, JSON: opts.JSON}, w)
	case "diff":
		commitHash := ""
		if len(opts.Args) >= 1 {
			commitHash = opts.Args[0]
		}
		return cmd.RunDiff(s, opts.Mark, commitHash, opts.JSON, w)
	case "impact":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		return cmd.RunImpact(cmd.ImpactOptions{Store: s, EntityID: entityID, Depth: opts.Depth, JSON: opts.JSON}, w)
	case "export":
		outputDir := c3Dir
		if len(opts.Args) >= 1 {
			outputDir = opts.Args[0]
		}
		return cmd.RunExport(cmd.ExportOptions{Store: s, OutputDir: outputDir, JSON: opts.JSON}, w)
	default:
		return fmt.Errorf("error: unknown command '%s'\nhint: run 'c3x --help' to see available commands", opts.Command)
	}
}

func runAdd(opts cmd.Options, s *store.Store, c3Dir string, w io.Writer) error {
	entityType := ""
	slug := ""
	if len(opts.Args) >= 1 {
		entityType = opts.Args[0]
	}
	if len(opts.Args) >= 2 {
		slug = opts.Args[1]
	}
	var buf bytes.Buffer
	var addW io.Writer = w
	if opts.JSON {
		addW = &buf
	}
	var err error
	if opts.Goal != "" || opts.Summary != "" || opts.Boundary != "" {
		addOpts := cmd.AddOptions{
			EntityType: entityType, Slug: slug, C3Dir: c3Dir, Store: s,
			Container: opts.Container, Feature: opts.Feature,
			Goal: opts.Goal, Summary: opts.Summary, Boundary: opts.Boundary,
		}
		err = cmd.RunAddRich(addOpts, addW)
	} else {
		err = cmd.RunAdd(entityType, slug, c3Dir, s, opts.Container, opts.Feature, addW)
	}
	if err == nil && opts.JSON {
		m := reAddID.FindStringSubmatch(buf.String())
		if len(m) >= 2 {
			result := cmd.AddResult{ID: m[1], Path: buf.String()}
			if mp := reAddPath.FindStringSubmatch(buf.String()); len(mp) >= 2 {
				result.Path = mp[1]
			}
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}
		w.Write(buf.Bytes())
	}
	return err
}

func runSet(opts cmd.Options, s *store.Store, c3Dir string, w io.Writer) error {
	id := ""
	value := ""
	if len(opts.Args) >= 1 {
		id = opts.Args[0]
	}
	if len(opts.Args) >= 2 {
		value = opts.Args[1]
	}
	if opts.Field == "" && opts.Section == "" && len(opts.Args) >= 2 {
		opts.Field = opts.Args[1]
		if len(opts.Args) >= 3 {
			value = opts.Args[2]
		}
	}
	return cmd.RunSet(cmd.SetOptions{
		Store: s, C3Dir: c3Dir, ID: id,
		Field: opts.Field, Section: opts.Section,
		Value: value, Append: opts.Append,
	}, w)
}

func runWire(opts cmd.Options, s *store.Store, w io.Writer) error {
	source, relation, target := "", "", ""
	if len(opts.Args) == 2 {
		source = opts.Args[0]
		target = opts.Args[1]
	} else {
		if len(opts.Args) >= 1 {
			source = opts.Args[0]
		}
		if len(opts.Args) >= 2 {
			relation = opts.Args[1]
		}
		if len(opts.Args) >= 3 {
			target = opts.Args[2]
		}
	}
	if opts.Remove || opts.Command == "unwire" {
		return cmd.RunUnwire(s, source, relation, target, w)
	}
	return cmd.RunWire(s, source, relation, target, w)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasMarkdownFiles(c3Dir string) bool {
	entries, err := os.ReadDir(c3Dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			return true
		}
		if e.IsDir() && e.Name() != "_index" {
			subEntries, _ := os.ReadDir(filepath.Join(c3Dir, e.Name()))
			for _, se := range subEntries {
				if !se.IsDir() && filepath.Ext(se.Name()) == ".md" {
					return true
				}
			}
		}
	}
	return false
}

func runLegacyCheck(c3Dir string, opts cmd.Options, w io.Writer) error {
	walkResult, err := walker.WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		return fmt.Errorf("error: walking .c3/: %w", err)
	}
	projectDir := config.ProjectDir(c3Dir)
	checkOpts := cmd.LegacyCheckOptions{
		Docs:          walkResult.Docs,
		Graph:         walker.BuildGraph(walkResult.Docs),
		JSON:          opts.JSON,
		ProjectDir:    projectDir,
		C3Dir:         c3Dir,
		ParseWarnings: walkResult.Warnings,
		IncludeADR:    opts.IncludeADR,
		Fix:           opts.Fix,
	}
	return cmd.RunLegacyCheck(checkOpts, w)
}
