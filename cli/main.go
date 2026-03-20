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

func main() {
	opts := cmd.ParseArgs(os.Args[1:])
	w := os.Stdout

	if opts.Version {
		fmt.Println(version)
		return
	}

	if opts.Help || opts.Command == "" {
		cmd.ShowHelp(opts.Command, w)
		return
	}

	// init is special — creates .c3/ with DB, no store needed
	if opts.Command == "init" {
		cwd := mustCwd()
		c3Dir := filepath.Join(cwd, ".c3")
		projectName := filepath.Base(cwd)
		if err := cmd.RunInitDB(c3Dir, projectName, w); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// capabilities is special — describes the CLI itself, no .c3/ needed
	if opts.Command == "capabilities" {
		cmd.ShowCapabilities(w)
		return
	}

	// marketplace is special — uses ~/.c3/marketplace/, no .c3/ needed
	if opts.Command == "marketplace" {
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

		var err error
		switch subCmd {
		case "add":
			err = cmd.RunMarketplaceAdd(mOpts, w)
		case "list":
			err = cmd.RunMarketplaceList(mOpts, w)
		case "show":
			err = cmd.RunMarketplaceShow(mOpts, w)
		case "update":
			err = cmd.RunMarketplaceUpdate(mOpts, w)
		case "remove":
			err = cmd.RunMarketplaceRemove(mOpts, w)
		default:
			cmd.ShowHelp("marketplace", w)
			return
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// All other commands need a .c3/ directory
	c3Dir := config.ResolveC3Dir(mustCwd(), opts.C3Dir)
	if c3Dir == "" {
		fmt.Fprintln(os.Stderr, "error: No .c3/ directory found")
		fmt.Fprintln(os.Stderr, "hint: run 'c3x init' to create one, or use --c3-dir <path>")
		os.Exit(1)
	}

	// migrate works on legacy files — special path, before DB detection
	if opts.Command == "migrate" {
		if err := cmd.RunMigrate(c3Dir, opts.KeepOriginals, w); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// Detect format: DB vs legacy files
	dbPath := filepath.Join(c3Dir, "c3.db")
	hasDB := fileExists(dbPath)

	// Legacy format detected — block all commands except check
	if !hasDB && hasMarkdownFiles(c3Dir) {
		if opts.Command == "check" {
			// check still works on file-based .c3/ via walker
			runLegacyCheck(c3Dir, opts, w)
			return
		}
		fmt.Fprintln(os.Stderr, "error: .c3/ contains markdown files but no database (c3.db)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "This version of c3x uses an embedded database.")
		fmt.Fprintln(os.Stderr, "Use /c3 in Claude Code to run an LLM-assisted migration that")
		fmt.Fprintln(os.Stderr, "validates and fixes malformed docs before importing.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or if docs are already valid: c3x migrate")
		os.Exit(1)
	}

	// Open the store (creates empty DB if none exists — fine for new projects via init)
	s, err := store.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: opening database: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	// Resolve project root (parent of .c3/)
	projectDir := config.ProjectDir(c3Dir)

	switch opts.Command {
	case "list":
		err = cmd.RunList(cmd.ListOptions{Store: s, JSON: opts.JSON, Flat: opts.Flat, Compact: opts.Compact, C3Dir: c3Dir, IncludeADR: opts.IncludeADR}, w)
	case "check":
		checkOpts := cmd.CheckOptions{
			Store:      s,
			JSON:       opts.JSON,
			ProjectDir: projectDir,
			C3Dir:      c3Dir,
			IncludeADR: opts.IncludeADR,
			Fix:        opts.Fix,
		}
		err = cmd.RunCheckV2(checkOpts, w)
	case "add":
		entityType := ""
		slug := ""
		if len(opts.Args) >= 1 {
			entityType = opts.Args[0]
		}
		if len(opts.Args) >= 2 {
			slug = opts.Args[1]
		}
		// Capture output to support --json
		var buf bytes.Buffer
		var addW io.Writer = w
		if opts.JSON {
			addW = &buf
		}
		if opts.Goal != "" || opts.Summary != "" || opts.Boundary != "" {
			addOpts := cmd.AddOptions{
				EntityType: entityType,
				Slug:       slug,
				C3Dir:      c3Dir,
				Store:      s,
				Container:  opts.Container,
				Feature:    opts.Feature,
				Goal:       opts.Goal,
				Summary:    opts.Summary,
				Boundary:   opts.Boundary,
			}
			err = cmd.RunAddRich(addOpts, addW)
		} else {
			err = cmd.RunAdd(entityType, slug, c3Dir, s, opts.Container, opts.Feature, addW)
		}
		if err == nil && opts.JSON {
			// Parse "Created: <path> (id: <id>)" from output
			re := regexp.MustCompile(`\(id: ([^)]+)\)`)
			m := re.FindStringSubmatch(buf.String())
			if len(m) >= 2 {
				result := cmd.AddResult{ID: m[1], Path: buf.String()}
				// Clean path from first Created line
				rePath := regexp.MustCompile(`Created: (\S+)`)
				if mp := rePath.FindStringSubmatch(buf.String()); len(mp) >= 2 {
					result.Path = mp[1]
				}
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				err = enc.Encode(result)
			} else {
				// Fallback: emit raw output
				w.Write(buf.Bytes())
			}
		}
	case "set":
		id := ""
		value := ""
		if len(opts.Args) >= 1 {
			id = opts.Args[0]
		}
		if len(opts.Args) >= 2 {
			value = opts.Args[1]
		}
		// If "section" is the second arg and no --section flag, treat positionally
		if opts.Field == "" && opts.Section == "" && len(opts.Args) >= 2 {
			opts.Field = opts.Args[1]
			if len(opts.Args) >= 3 {
				value = opts.Args[2]
			}
		}
		setOpts := cmd.SetOptions{
			Store:   s,
			C3Dir:   c3Dir,
			ID:      id,
			Field:   opts.Field,
			Section: opts.Section,
			Value:   value,
			Append:  opts.Append,
		}
		err = cmd.RunSet(setOpts, w)
	case "wire", "unwire":
		source, relation, target := "", "", ""
		if len(opts.Args) == 2 {
			// Short form: wire <src> <tgt> (cite is default)
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
			err = cmd.RunUnwire(s, source, relation, target, w)
		} else {
			err = cmd.RunWire(s, source, relation, target, w)
		}
	case "lookup":
		if len(opts.Args) < 1 {
			fmt.Fprintln(os.Stderr, "error: lookup requires a <file-path> argument")
			fmt.Fprintln(os.Stderr, "hint: run 'c3x lookup --help' for usage")
			os.Exit(1)
		}
		err = cmd.RunLookup(cmd.LookupOptions{
			Store:      s,
			FilePath:   opts.Args[0],
			JSON:       opts.JSON,
			ProjectDir: projectDir,
			C3Dir:      c3Dir,
		}, w)
	case "codemap":
		err = cmd.RunCodemap(cmd.CodemapOptions{
			C3Dir: c3Dir,
			Store: s,
			JSON:  opts.JSON,
		}, w)
	case "coverage":
		err = cmd.RunCoverage(cmd.CoverageOptions{
			Store:      s,
			C3Dir:      c3Dir,
			ProjectDir: projectDir,
			JSON:       opts.JSON,
		}, w)
	case "schema":
		entityType := ""
		if len(opts.Args) >= 1 {
			entityType = opts.Args[0]
		}
		err = cmd.RunSchema(entityType, opts.JSON, w)
	case "graph":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		if entityID == "" {
			fmt.Fprintln(os.Stderr, "error: graph requires an <entity-id> argument")
			fmt.Fprintln(os.Stderr, "hint: run 'c3x graph --help' for usage")
			os.Exit(1)
		}
		err = cmd.RunGraph(cmd.GraphOptions{
			Store:     s,
			EntityID:  entityID,
			Depth:     opts.Depth,
			Direction: opts.Direction,
			Format:    opts.Format,
			JSON:      opts.JSON,
			C3Dir:     c3Dir,
		}, w)
	case "delete":
		id := ""
		if len(opts.Args) >= 1 {
			id = opts.Args[0]
		}
		err = cmd.RunDelete(cmd.DeleteOptions{
			C3Dir:  c3Dir,
			ID:     id,
			Store:  s,
			DryRun: opts.DryRun,
		}, w)
	case "query":
		queryTerm := ""
		if len(opts.Args) >= 1 {
			queryTerm = opts.Args[0]
		}
		err = cmd.RunQuery(cmd.QueryOptions{
			Store:      s,
			Query:      queryTerm,
			TypeFilter: opts.TypeFilter,
			Limit:      opts.Limit,
			JSON:       opts.JSON,
		}, w)
	case "diff":
		commitHash := ""
		if len(opts.Args) >= 1 {
			commitHash = opts.Args[0]
		}
		err = cmd.RunDiff(s, opts.Mark, commitHash, opts.JSON, w)
	case "impact":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		err = cmd.RunImpact(cmd.ImpactOptions{
			Store:    s,
			EntityID: entityID,
			Depth:    opts.Depth,
			JSON:     opts.JSON,
		}, w)
	case "export":
		outputDir := c3Dir
		if len(opts.Args) >= 1 {
			outputDir = opts.Args[0]
		}
		err = cmd.RunExport(cmd.ExportOptions{
			Store:     s,
			OutputDir: outputDir,
			JSON:      opts.JSON,
		}, w)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n", opts.Command)
		fmt.Fprintln(os.Stderr, "hint: run 'c3x --help' to see available commands")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mustCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot get working directory: %v\n", err)
		os.Exit(1)
	}
	return dir
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
		// Check subdirectories (containers have .md files inside)
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

func runLegacyCheck(c3Dir string, opts cmd.Options, w io.Writer) {
	// Use the old walker-based check for pre-migration validation
	walkResult, err := walker.WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: walking .c3/: %v\n", err)
		os.Exit(1)
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
	if err := cmd.RunLegacyCheck(checkOpts, w); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
