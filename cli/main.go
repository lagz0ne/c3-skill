package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/config"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

var version = "dev"

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

	// init is special — creates .c3/ with local cache + canonical files, no store needed
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

	// template is special — outputs fillable scaffolds, no .c3/ needed
	if opts.Command == "template" {
		entityType := ""
		if len(opts.Args) >= 1 {
			entityType = opts.Args[0]
		}
		return cmd.RunTemplate(entityType, w)
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
	if opts.Command == "cache" {
		return runCache(opts, c3Dir, w)
	}

	dbPath := filepath.Join(c3Dir, "c3.db")
	hasDB := fileExists(dbPath)
	hasCanonical := hasCanonicalDocs(c3Dir)
	skipPreVerify := commandMutatesCanonical(opts) ||
		(opts.Command == "sync" && len(opts.Args) >= 1 && opts.Args[0] == "export") ||
		opts.Command == "migrate"

	// v9 workflow treats canonical .c3/ markdown as submitted truth.
	// When canonical files exist, refresh/rebuild local cache from them before dispatch.
	if hasCanonical && !skipPreVerify {
		if err := cmd.RunVerify(cmd.VerifyOptions{C3Dir: c3Dir, JSON: opts.JSON}, io.Discard); err != nil {
			return fmt.Errorf("error: canonical .c3/ is not ready for %q: %w", opts.Command, err)
		}
		hasDB = fileExists(dbPath)
	}

	if !hasDB {
		return fmt.Errorf("error: local C3 cache unavailable at %s\nhint: run 'c3x verify' to rebuild from canonical .c3/, or 'c3x init' if this project is not onboarded", dbPath)
	}

	var rollback *mutationSnapshot
	if commandMutatesCanonical(opts) {
		rollback, err = newMutationSnapshot(c3Dir)
		if err != nil {
			return fmt.Errorf("error: create mutation rollback snapshot: %w", err)
		}
		defer rollback.cleanup()
	}

	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("error: opening database: %w", err)
	}

	cmdErr := runCommand(opts, s, c3Dir, w)
	closeErr := s.Close()
	if cmdErr != nil {
		if rollback != nil {
			if restoreErr := rollback.restore(); restoreErr != nil {
				return fmt.Errorf("%w\nrollback failed: %v", cmdErr, restoreErr)
			}
		}
		return cmdErr
	}
	if closeErr != nil {
		if rollback != nil {
			if restoreErr := rollback.restore(); restoreErr != nil {
				return fmt.Errorf("error: closing database: %w\nrollback failed: %v", closeErr, restoreErr)
			}
		}
		return fmt.Errorf("error: closing database: %w", closeErr)
	}
	return nil
}

type mutationSnapshot struct {
	c3Dir     string
	backupDir string
}

func newMutationSnapshot(c3Dir string) (*mutationSnapshot, error) {
	tmpDir, err := os.MkdirTemp("", "c3-mutation-rollback-")
	if err != nil {
		return nil, err
	}
	snap := &mutationSnapshot{c3Dir: c3Dir, backupDir: tmpDir}
	if err := copyTree(c3Dir, filepath.Join(tmpDir, "c3")); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}
	return snap, nil
}

func (s *mutationSnapshot) cleanup() {
	if s != nil {
		_ = os.RemoveAll(s.backupDir)
	}
}

func (s *mutationSnapshot) restore() error {
	if s == nil {
		return nil
	}
	backup := filepath.Join(s.backupDir, "c3")
	if err := os.RemoveAll(s.c3Dir); err != nil {
		return err
	}
	return copyTree(backup, s.c3Dir)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func copyFile(src, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
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

func hasCanonicalDocs(c3Dir string) bool {
	if fileExists(filepath.Join(c3Dir, "README.md")) {
		return true
	}
	if fileExists(filepath.Join(c3Dir, "code-map.yaml")) {
		return true
	}
	matches, err := filepath.Glob(filepath.Join(c3Dir, "adr", "*.md"))
	if err == nil && len(matches) > 0 {
		return true
	}
	matches, err = filepath.Glob(filepath.Join(c3Dir, "refs", "*.md"))
	if err == nil && len(matches) > 0 {
		return true
	}
	matches, err = filepath.Glob(filepath.Join(c3Dir, "rules", "*.md"))
	if err == nil && len(matches) > 0 {
		return true
	}
	matches, err = filepath.Glob(filepath.Join(c3Dir, "recipes", "*.md"))
	if err == nil && len(matches) > 0 {
		return true
	}
	matches, err = filepath.Glob(filepath.Join(c3Dir, "c3-*", "README.md"))
	return err == nil && len(matches) > 0
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
		var content []byte
		content, err = io.ReadAll(os.Stdin)
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
		err = cmd.RunQuery(cmd.QueryOptions{Store: s, Query: queryTerm, TypeFilter: opts.TypeFilter, Limit: opts.Limit, JSON: opts.JSON, IncludeADR: opts.IncludeADR}, w)
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
		subCmd := ""
		if len(opts.Args) >= 1 {
			subCmd = opts.Args[0]
		}
		switch subCmd {
		case "repair-plan":
			err = cmd.RunMigrateRepairPlan(s, w)
		case "repair":
			id := ""
			if len(opts.Args) >= 2 {
				id = opts.Args[1]
			}
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("error: no input on stdin\nhint: pipe content: cat section.md | c3x migrate repair <id> --section <name>")
			}
			content, readErr := io.ReadAll(os.Stdin)
			if readErr != nil {
				return fmt.Errorf("error: reading stdin: %w", readErr)
			}
			err = cmd.RunMigrateRepairSection(s, id, opts.Section, string(content), w)
		default:
			err = cmd.RunMigrateV2(cmd.MigrateV2Options{Store: s, DryRun: opts.DryRun, JSON: opts.JSON, Continue: opts.Continue}, w)
		}
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
		if len(opts.Args) >= 1 && opts.Args[0] == "repair-plan" {
			return false
		}
		return !opts.DryRun
	default:
		return false
	}
}

func runCache(opts cmd.Options, c3Dir string, w io.Writer) error {
	subCmd := ""
	if len(opts.Args) >= 1 {
		subCmd = opts.Args[0]
	}
	switch subCmd {
	case "clear":
		return cmd.RunCacheClear(c3Dir, w)
	default:
		cmd.ShowHelp("cache", w)
		return nil
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

	if opts.DryRun {
		return cmd.RunAddDryRun(entityType, slug, s, opts.Container, opts.Feature, os.Stdin, w)
	}

	format := cmd.FormatHuman
	if opts.JSON {
		format = cmd.ResolveFormat(opts.JSONExplicit, os.Getenv("C3X_MODE") == "agent")
	}
	return cmd.RunAddFormatted(entityType, slug, s, opts.Container, opts.Feature, os.Stdin, w, format)
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
