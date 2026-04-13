package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/config"
	"github.com/lagz0ne/c3-design/cli/internal/schema"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

var version = "dev"

var reAddID = regexp.MustCompile(`\(id: ([^)]+)\)`)

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

	if opts.Help {
		cmd.ShowHelp(opts.Command, w)
		return nil
	}
	if opts.Command == "" {
		return runNoArgs(opts, w)
	}

	// init is special — creates .c3/ with DB, no store needed
	if opts.Command == "init" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error: cannot get working directory: %w", err)
		}
		c3Dir := filepath.Join(cwd, ".c3")
		projectName := filepath.Base(cwd)
		if err := cmd.RunInitDB(c3Dir, projectName, w); err != nil {
			return err
		}
		s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
		if err != nil {
			return fmt.Errorf("error: opening database: %w", err)
		}
		defer s.Close()
		return cmd.AutoExportCanonical(s, c3Dir)
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

	if opts.Command == "migrate-legacy" {
		if opts.DryRun {
			return cmd.RunMigrateDryRun(c3Dir, opts.JSON, w)
		}
		if err := cmd.RunMigrate(c3Dir, opts.KeepOriginals, w); err != nil {
			return err
		}
		s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
		if err != nil {
			return fmt.Errorf("error: opening database: %w", err)
		}
		defer s.Close()
		return cmd.AutoExportCanonical(s, c3Dir)
	}
	if opts.Command == "import" {
		return cmd.RunImport(cmd.ImportOptions{C3Dir: c3Dir, Force: opts.Force}, w)
	}
	if opts.Command == "git" {
		return runGit(opts, config.ProjectDir(c3Dir), c3Dir, w)
	}
	if opts.Command == "verify" {
		return cmd.RunVerify(cmd.VerifyOptions{C3Dir: c3Dir, JSON: opts.JSON}, w)
	}
	if opts.Command == "repair" {
		return cmd.RunRepair(cmd.RepairOptions{C3Dir: c3Dir, JSON: opts.JSON}, w)
	}

	// Detect format: DB
	dbPath := filepath.Join(c3Dir, "c3.db")
	hasDB := fileExists(dbPath)

	if !hasDB {
		return fmt.Errorf("error: no database found at %s\nhint: run 'c3x init' to create one, or 'c3x migrate-legacy' if you have legacy .c3/ markdown files", dbPath)
	}

	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("error: opening database: %w", err)
	}
	defer s.Close()

	return runCommand(opts, s, c3Dir, w)
}

func runGit(opts cmd.Options, projectDir, c3Dir string, w io.Writer) error {
	subCmd := ""
	if len(opts.Args) >= 1 {
		subCmd = opts.Args[0]
	}
	switch subCmd {
	case "install":
		return cmd.RunGitInstall(projectDir, c3Dir, w)
	default:
		cmd.ShowHelp("git", w)
		return nil
	}
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
	mutating := commandMutatesCanonical(opts)

	var err error
	switch opts.Command {
	case "status":
		return cmd.RunStatus(cmd.StatusOptions{
			Store: s, C3Dir: c3Dir, ProjectDir: projectDir,
			JSONExplicit: opts.JSONExplicit,
		}, w)
	case "list":
		err = cmd.RunList(cmd.ListOptions{Store: s, JSON: opts.JSON, Flat: opts.Flat, Compact: opts.Compact, C3Dir: c3Dir, IncludeADR: opts.IncludeADR, JSONExplicit: opts.JSONExplicit}, w)
	case "check":
		err = cmd.RunCheckV2(cmd.CheckOptions{
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
		err = cmd.RunRead(cmd.ReadOptions{Store: s, ID: entityID, JSON: opts.JSON, Section: opts.Section, Full: opts.Full}, w)
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
		err = cmd.RunWrite(cmd.WriteOptions{Store: s, ID: entityID, Section: opts.Section, Content: string(content)}, w)
	case "add":
		err = runAdd(opts, s, w)
	case "set":
		err = runSet(opts, s, c3Dir, w)
	case "wire", "unwire":
		err = runWire(opts, s, w)
	case "lookup":
		if len(opts.Args) < 1 {
			return fmt.Errorf("error: lookup requires a <file-path> argument\nhint: run 'c3x lookup --help' for usage")
		}
		err = cmd.RunLookup(cmd.LookupOptions{
			Store:      s,
			FilePath:   opts.Args[0],
			JSON:       opts.JSON,
			ProjectDir: projectDir,
			C3Dir:      c3Dir,
		}, w)
	case "codemap":
		err = cmd.RunCodemap(cmd.CodemapOptions{Store: s, JSON: opts.JSON}, w)
	case "coverage":
		err = cmd.RunCoverage(cmd.CoverageOptions{Store: s, C3Dir: c3Dir, ProjectDir: projectDir, JSON: opts.JSON}, w)
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
			return fmt.Errorf("error: graph requires an <entity-id> argument\nhint: run 'c3x graph --help' for usage")
		}
		err = cmd.RunGraph(cmd.GraphOptions{
			Store: s, EntityID: entityID, Depth: opts.Depth,
			Direction: opts.Direction, Format: opts.Format,
			JSON: opts.JSON, C3Dir: c3Dir,
		}, w)
	case "delete":
		id := ""
		if len(opts.Args) >= 1 {
			id = opts.Args[0]
		}
		err = cmd.RunDelete(cmd.DeleteOptions{C3Dir: c3Dir, ID: id, Store: s, DryRun: opts.DryRun}, w)
	case "query":
		queryTerm := ""
		if len(opts.Args) >= 1 {
			queryTerm = opts.Args[0]
		}
		err = cmd.RunQuery(cmd.QueryOptions{Store: s, Query: queryTerm, TypeFilter: opts.TypeFilter, Limit: opts.Limit, JSON: opts.JSON}, w)
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
		err = cmd.RunImpact(cmd.ImpactOptions{Store: s, EntityID: entityID, Depth: opts.Depth, JSON: opts.JSON}, w)
	case "export":
		outputDir := c3Dir
		if len(opts.Args) >= 1 {
			outputDir = opts.Args[0]
		}
		err = cmd.RunExport(cmd.ExportOptions{Store: s, OutputDir: outputDir, JSON: opts.JSON}, w)
	case "sync":
		subCmd := ""
		if len(opts.Args) >= 1 {
			subCmd = opts.Args[0]
		}
		switch subCmd {
		case "export":
			outputDir := c3Dir
			if len(opts.Args) >= 2 {
				outputDir = opts.Args[1]
			}
			err = cmd.RunSyncExport(cmd.ExportOptions{Store: s, OutputDir: outputDir, JSON: opts.JSON}, w)
		case "check":
			outputDir := c3Dir
			if len(opts.Args) >= 2 {
				outputDir = opts.Args[1]
			}
			err = cmd.RunSyncCheck(cmd.ExportOptions{Store: s, OutputDir: outputDir, JSON: opts.JSON}, w)
		default:
			cmd.ShowHelp("sync", w)
			return nil
		}
	case "nodes":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		err = cmd.RunNodes(cmd.NodesOptions{Store: s, EntityID: entityID, JSON: opts.JSON}, w)
	case "hash":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		err = cmd.RunHash(cmd.HashOptions{Store: s, EntityID: entityID, Recompute: opts.Recompute}, w)
	case "versions":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		err = cmd.RunVersions(cmd.VersionsOptions{Store: s, EntityID: entityID, JSON: opts.JSON}, w)
	case "version":
		entityID := ""
		versionNum := 0
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		if len(opts.Args) >= 2 {
			versionNum, _ = strconv.Atoi(opts.Args[1])
		}
		err = cmd.RunVersion(cmd.VersionOptions{Store: s, EntityID: entityID, Version: versionNum}, w)
	case "prune":
		entityID := ""
		if len(opts.Args) >= 1 {
			entityID = opts.Args[0]
		}
		err = cmd.RunPrune(cmd.PruneOptions{Store: s, EntityID: entityID, Keep: opts.Keep}, w)
	case "migrate":
		err = cmd.RunMigrateV2(cmd.MigrateV2Options{Store: s, DryRun: opts.DryRun}, w)
	default:
		return fmt.Errorf("error: unknown command '%s'\nhint: run 'c3x --help' to see available commands", opts.Command)
	}

	if err != nil {
		return err
	}
	if mutating {
		return cmd.AutoExportCanonical(s, c3Dir)
	}
	return nil
}

func commandMutatesCanonical(opts cmd.Options) bool {
	switch opts.Command {
	case "write", "add", "set", "wire", "unwire", "codemap":
		return true
	case "delete":
		return !opts.DryRun
	case "migrate":
		return !opts.DryRun
	default:
		return false
	}
}

func runAdd(opts cmd.Options, s *store.Store, w io.Writer) error {
	entityType := ""
	slug := ""
	if len(opts.Args) >= 1 {
		entityType = opts.Args[0]
	}
	if len(opts.Args) >= 2 {
		slug = opts.Args[1]
	}

	// Read body from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("error: c3x add requires body content via stdin\nhint: cat body.md | c3x add <type> <slug>\nhint: run 'c3x schema <type>' to see required sections")
	}

	var buf bytes.Buffer
	var addW io.Writer = w
	if opts.JSON {
		addW = &buf
	}

	err := cmd.RunAdd(entityType, slug, s, opts.Container, opts.Feature, os.Stdin, addW)
	if err != nil {
		return err
	}

	if opts.JSON {
		m := reAddID.FindStringSubmatch(buf.String())
		if len(m) >= 2 {
			result := cmd.AddResult{ID: m[1], Type: entityType}
			if sections := schema.ForType(entityType); sections != nil {
				for _, sec := range sections {
					result.Sections = append(result.Sections, sec.Name)
				}
			}
			enc := json.NewEncoder(w)
			if os.Getenv("C3X_MODE") != "agent" {
				enc.SetIndent("", "  ")
			}
			return enc.Encode(result)
		}
		w.Write(buf.Bytes())
	}
	return nil
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
	if opts.Field == "" && opts.Section == "" && !opts.Stdin && len(opts.Args) >= 2 {
		opts.Field = opts.Args[1]
		if len(opts.Args) >= 3 {
			value = opts.Args[2]
		}
	}
	if opts.Stdin {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("error: --stdin requires piped input\nhint: echo '{\"fields\":{...}}' | c3x set <id> --stdin")
		}
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error: reading stdin: %w", err)
		}
		value = string(data)
	}
	return cmd.RunSet(cmd.SetOptions{
		Store: s, C3Dir: c3Dir, ID: id,
		Field: opts.Field, Section: opts.Section,
		Value: value, Append: opts.Append, Remove: opts.Remove, Stdin: opts.Stdin,
	}, w)
}

func runWire(opts cmd.Options, s *store.Store, w io.Writer) error {
	if len(opts.Args) < 2 {
		return fmt.Errorf("error: usage: c3x wire <source> <target> [target2 ...]\nhint: c3x wire c3-101 ref-jwt ref-error-handling")
	}

	source := opts.Args[0]
	var targets []string
	relation := ""

	// Check if second arg is a relation type
	if len(opts.Args) >= 3 && opts.Args[1] == "cite" {
		relation = opts.Args[1]
		targets = opts.Args[2:]
	} else {
		targets = opts.Args[1:]
	}

	for _, target := range targets {
		if opts.Remove || opts.Command == "unwire" {
			if err := cmd.RunUnwire(s, source, relation, target, w); err != nil {
				return err
			}
		} else {
			if err := cmd.RunWire(s, source, relation, target, w); err != nil {
				return err
			}
		}
	}
	return nil
}

func runNoArgs(opts cmd.Options, w io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		cmd.ShowHelp("", w)
		return nil
	}
	c3Dir := config.ResolveC3Dir(cwd, opts.C3Dir)
	if c3Dir == "" {
		cmd.ShowHelp("", w)
		return nil
	}
	dbPath := filepath.Join(c3Dir, "c3.db")
	if !fileExists(dbPath) {
		cmd.ShowHelp("", w)
		return nil
	}
	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("error: opening database: %w", err)
	}
	defer s.Close()
	return cmd.RunStatus(cmd.StatusOptions{
		Store:        s,
		C3Dir:        c3Dir,
		ProjectDir:   config.ProjectDir(c3Dir),
		JSONExplicit: opts.JSONExplicit,
	}, w)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
