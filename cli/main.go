package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/config"
	"github.com/lagz0ne/c3-design/cli/internal/coord"
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
	stdinTerminal := true
	if stat, err := os.Stdin.Stat(); err == nil {
		stdinTerminal = (stat.Mode() & os.ModeCharDevice) != 0
	}
	return runWithIO(argv, os.Stdin, stdinTerminal, w, os.Stderr, true)
}

func runWithIO(argv []string, stdin io.Reader, stdinTerminal bool, w io.Writer, stderr io.Writer, coordinate bool) error {
	opts := cmd.ParseArgs(argv)

	if opts.File != "" && commandAcceptsFile(opts.Command) {
		f, err := os.Open(opts.File)
		if err != nil {
			return fmt.Errorf("error: opening --file %s: %w", opts.File, err)
		}
		defer f.Close()
		stdin = f
		stdinTerminal = false
	}

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

	if opts.Command == "git" {
		return runGit(opts, config.ProjectDir(c3Dir), c3Dir, w)
	}

	dbPath := filepath.Join(c3Dir, "c3.db")
	hasDB := fileExists(dbPath)
	hasCanonical := hasCanonicalDocs(c3Dir)
	mutates := commandMutatesCanonical(opts)
	if coordinate && mutates {
		return runThroughCoordinator(argv, stdin, stdinTerminal, c3Dir, w, stderr)
	}

	// Mutations bypass preverify (ADR mutation-preverify-repair-bypass): the
	// mutation itself may be the fix.
	if hasCanonical {
		if mutates {
			if err := cmd.EnsureLocalCache(c3Dir, opts.IncludeADR, opts.Only, io.Discard); err != nil {
				return fmt.Errorf("error: refresh cache before %q: %w", opts.Command, err)
			}
		} else if err := cmd.RunVerify(cmd.VerifyOptions{C3Dir: c3Dir, JSON: opts.JSON, IncludeADR: opts.IncludeADR, Only: opts.Only}, io.Discard); err != nil {
			fmt.Fprintln(stderr, "warning: .c3/ drift detected; run 'c3x check' to reconcile")
		}
		hasDB = fileExists(dbPath)
	}

	if !hasDB {
		return fmt.Errorf("error: local C3 cache unavailable at %s\nhint: run 'c3x check' to rebuild from canonical .c3/, or 'c3x init' if this project is not onboarded", dbPath)
	}

	var rollback *mutationSnapshot
	if mutates {
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

	cmdErr := runCommand(opts, s, c3Dir, stdin, stdinTerminal, w)
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

func runThroughCoordinator(argv []string, stdin io.Reader, stdinTerminal bool, c3Dir string, w io.Writer, stderr io.Writer) error {
	if os.Getenv("C3X_COORDINATOR") == "0" {
		return runWithIO(argv, stdin, stdinTerminal, w, stderr, false)
	}
	data, err := readCoordinatorStdin(stdin, stdinTerminal)
	if err != nil {
		return fmt.Errorf("error: reading stdin: %w", err)
	}
	cwd, _ := os.Getwd()
	req := coord.Request{
		Argv:          append([]string(nil), argv...),
		Stdin:         data,
		StdinTerminal: stdinTerminal,
		CWD:           cwd,
		C3XMode:       os.Getenv("C3X_MODE"),
	}
	if resp, handled, err := coord.TryForward(c3Dir, req); handled {
		writeCoordinatorResponse(resp, w, stderr)
		if resp.Error != "" {
			return fmt.Errorf("%s", resp.Error)
		}
		return err
	}

	leader, err := coord.NewLeader(c3Dir)
	if err != nil {
		if resp, handled, retryErr := coord.ForwardWithRetry(c3Dir, req, 2*time.Second); handled {
			writeCoordinatorResponse(resp, w, stderr)
			if resp.Error != "" {
				return fmt.Errorf("%s", resp.Error)
			}
			return retryErr
		}
		if err == coord.ErrBusy {
			return fmt.Errorf("error: write coordinator busy for %s", c3Dir)
		}
		return runWithIO(argv, bytes.NewReader(data), stdinTerminal, w, stderr, false)
	}
	defer leader.Close()
	resp := leader.Serve(req, func(queued coord.Request) coord.Response {
		return runQueuedRequest(queued)
	})
	writeCoordinatorResponse(resp, w, stderr)
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func runQueuedRequest(req coord.Request) coord.Response {
	var stdout, stderr bytes.Buffer
	oldMode, hadMode := os.LookupEnv("C3X_MODE")
	if req.C3XMode != "" {
		_ = os.Setenv("C3X_MODE", req.C3XMode)
	} else {
		_ = os.Unsetenv("C3X_MODE")
	}
	oldWD, wdErr := os.Getwd()
	if req.CWD != "" {
		_ = os.Chdir(req.CWD)
	}
	err := runWithIO(req.Argv, bytes.NewReader(req.Stdin), req.StdinTerminal, &stdout, &stderr, false)
	if req.CWD != "" && wdErr == nil {
		_ = os.Chdir(oldWD)
	}
	if hadMode {
		_ = os.Setenv("C3X_MODE", oldMode)
	} else {
		_ = os.Unsetenv("C3X_MODE")
	}
	resp := coord.Response{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func writeCoordinatorResponse(resp coord.Response, w io.Writer, stderr io.Writer) {
	if resp.Stdout != "" {
		_, _ = io.WriteString(w, resp.Stdout)
	}
	if resp.Stderr != "" {
		_, _ = io.WriteString(stderr, resp.Stderr)
	}
}

func readCoordinatorStdin(stdin io.Reader, stdinTerminal bool) ([]byte, error) {
	if stdinTerminal || stdin == nil {
		return nil, nil
	}
	return io.ReadAll(stdin)
}

// runCommand dispatches to the appropriate command handler.
func runCommand(opts cmd.Options, s *store.Store, c3Dir string, stdin io.Reader, stdinTerminal bool, w io.Writer) error {
	projectDir := config.ProjectDir(c3Dir)
	mutating := commandMutatesCanonical(opts)

	var err error
	switch opts.Command {
	case "list":
		err = cmd.RunList(cmd.ListOptions{Store: s, JSON: opts.JSON, Flat: opts.Flat, Compact: opts.Compact, C3Dir: c3Dir, IncludeADR: opts.IncludeADR, JSONExplicit: opts.JSONExplicit}, w)
	case "check":
		var verifyOut bytes.Buffer
		if err = cmd.RunVerify(cmd.VerifyOptions{C3Dir: c3Dir, JSON: opts.JSON, IncludeADR: opts.IncludeADR, Only: opts.Only}, &verifyOut); err != nil {
			_, _ = io.Copy(w, &verifyOut)
			return fmt.Errorf("%w\nhint: run c3x check again after resolving", err)
		}
		err = cmd.RunCheckV2(cmd.CheckOptions{
			Store:      s,
			JSON:       opts.JSON,
			ProjectDir: projectDir,
			C3Dir:      c3Dir,
			IncludeADR: opts.IncludeADR,
			Fix:        opts.Fix,
			Only:       opts.Only,
			Rules:      opts.Rules,
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
		if stdinTerminal {
			return fmt.Errorf("error: no input on stdin\nhint: pipe content: echo '...' | c3x write <id>, or: c3x read <id> | c3x write <id>")
		}
		var content []byte
		content, err = io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("error: reading stdin: %w", err)
		}
		err = cmd.RunWrite(cmd.WriteOptions{Store: s, ID: entityID, Section: opts.Section, Content: string(content)}, w)
	case "add":
		err = runAdd(opts, s, stdin, stdinTerminal, w)
	case "set":
		err = runSet(opts, s, c3Dir, stdin, stdinTerminal, w)
	case "wire":
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
	case "repair":
		err = cmd.RunRepair(cmd.RepairOptions{C3Dir: c3Dir, JSON: opts.JSON, IncludeADR: opts.IncludeADR, Only: opts.Only}, w)
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
	case "write", "add", "set", "wire", "delete", "repair":
		if opts.Command == "delete" {
			return !opts.DryRun
		}
		return true
	case "check":
		return opts.Fix
	default:
		return false
	}
}

func runAdd(opts cmd.Options, s *store.Store, stdin io.Reader, stdinTerminal bool, w io.Writer) error {
	entityType := ""
	slug := ""
	if len(opts.Args) >= 1 {
		entityType = opts.Args[0]
	}
	if len(opts.Args) >= 2 {
		slug = opts.Args[1]
	}

	// Read body from stdin
	if stdinTerminal {
		return fmt.Errorf("error: c3x add requires body content via stdin\nhint: cat body.md | c3x add <type> <slug>\nhint: run 'c3x schema <type>' to see required sections")
	}

	if opts.DryRun {
		return cmd.RunAddDryRun(entityType, slug, s, opts.Container, opts.Feature, stdin, w)
	}

	format := cmd.FormatHuman
	if opts.JSON {
		format = cmd.ResolveFormat(opts.JSONExplicit, os.Getenv("C3X_MODE") == "agent")
	}
	return cmd.RunAddFormatted(entityType, slug, s, opts.Container, opts.Feature, stdin, w, format)
}

func runSet(opts cmd.Options, s *store.Store, c3Dir string, stdin io.Reader, stdinTerminal bool, w io.Writer) error {
	if opts.Section != "" {
		return fmt.Errorf("error: c3x set no longer accepts --section\nhint: use 'c3x write <id> --section <name>' (body via stdin or --file)")
	}
	if opts.Stdin {
		return fmt.Errorf("error: c3x set no longer accepts --stdin batch mode\nhint: use 'c3x write <id>' for body or multiple 'c3x set <id> <field> <value>' for fields")
	}
	id := ""
	value := ""
	if len(opts.Args) >= 1 {
		id = opts.Args[0]
	}
	if len(opts.Args) >= 2 {
		value = opts.Args[1]
	}
	if opts.Field == "" && len(opts.Args) >= 2 {
		opts.Field = opts.Args[1]
		if len(opts.Args) >= 3 {
			value = opts.Args[2]
		}
	}
	return cmd.RunSet(cmd.SetOptions{
		Store: s, C3Dir: c3Dir, ID: id,
		Field: opts.Field,
		Value: value, Append: opts.Append, Remove: opts.Remove,
	}, w)
}

func commandAcceptsFile(cmd string) bool {
	switch cmd {
	case "write", "add", "set":
		return true
	}
	return false
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
		if opts.Remove {
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
	cmd.ShowHelp("", w)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
