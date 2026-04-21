package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lagz0ne/c3-design/cli/internal/config"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

type VerifyOptions struct {
	C3Dir      string
	JSON       bool
	IncludeADR bool
	Only       []string
}

type RepairOptions struct {
	C3Dir      string
	JSON       bool
	IncludeADR bool
	Only       []string
}

func RunVerify(opts VerifyOptions, w io.Writer) error {
	if err := reportBrokenSeals(opts.C3Dir, opts.IncludeADR, opts.Only, w); err != nil {
		return err
	}
	if err := ensureLocalCache(opts.C3Dir, opts.IncludeADR, opts.Only, w); err != nil {
		return err
	}
	return runVerificationSuite(opts.C3Dir, opts.JSON, opts.IncludeADR, opts.Only, w)
}

func RunRepair(opts RepairOptions, w io.Writer) error {
	if err := RunImport(ImportOptions{C3Dir: opts.C3Dir, Force: true, SkipBackup: true}, io.Discard); err != nil {
		return err
	}
	s, err := store.Open(filepath.Join(opts.C3Dir, "c3.db"))
	if err != nil {
		return fmt.Errorf("repair: open rebuilt cache: %w", err)
	}
	defer s.Close()
	if err := RunSyncExport(ExportOptions{Store: s, OutputDir: opts.C3Dir, JSON: opts.JSON}, io.Discard); err != nil {
		return err
	}
	fmt.Fprintf(w, "Rebuilt local C3 cache from canonical .c3/\n")
	fmt.Fprintf(w, "Resealed canonical .c3/ tree\n")
	return runVerificationSuite(opts.C3Dir, opts.JSON, opts.IncludeADR, opts.Only, w)
}

func ensureLocalCache(c3Dir string, includeADR bool, only []string, w io.Writer) error {
	dbPath := filepath.Join(c3Dir, "c3.db")
	if !pathExists(dbPath) {
		if err := RunImport(ImportOptions{C3Dir: c3Dir, SkipBackup: true, AllowADRDrift: !includeADR, Only: only}, io.Discard); err != nil {
			return fmt.Errorf("verify: rebuild local cache: %w", err)
		}
		fmt.Fprintf(w, "Rebuilt local C3 cache from canonical .c3/\n")
		return nil
	}

	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("verify: open cache: %w", err)
	}
	defer s.Close()
	if err := RunSyncCheck(ExportOptions{Store: s, OutputDir: c3Dir, IncludeADR: includeADR, Only: only}, io.Discard); err == nil {
		return nil
	}

	if err := RunImport(ImportOptions{C3Dir: c3Dir, Force: true, SkipBackup: true}, io.Discard); err != nil {
		return fmt.Errorf("verify: refresh local cache: %w", err)
	}
	fmt.Fprintf(w, "Refreshed local C3 cache from canonical .c3/\n")
	return nil
}

func runVerificationSuite(c3Dir string, json bool, includeADR bool, only []string, w io.Writer) error {
	s, err := store.Open(filepath.Join(c3Dir, "c3.db"))
	if err != nil {
		return fmt.Errorf("verify: open cache: %w", err)
	}
	defer s.Close()

	var checkOut bytes.Buffer
	if err := RunCheckV2(CheckOptions{
		Store:      s,
		JSON:       json,
		ProjectDir: config.ProjectDir(c3Dir),
		C3Dir:      c3Dir,
		IncludeADR: includeADR,
		Only:       only,
	}, &checkOut); err != nil {
		if checkOut.Len() > 0 {
			if _, copyErr := io.Copy(w, &checkOut); copyErr != nil {
				return copyErr
			}
		}
		return err
	}
	if _, err := io.Copy(w, &checkOut); err != nil {
		return err
	}
	return RunSyncCheck(ExportOptions{Store: s, OutputDir: c3Dir, JSON: json, IncludeADR: includeADR, Only: only}, w)
}

func reportBrokenSeals(c3Dir string, includeADR bool, only []string, w io.Writer) error {
	_, broken, err := snapshotCanonicalTree(c3Dir, true)
	if err != nil {
		return fmt.Errorf("verify: read canonical tree: %w", err)
	}
	if !includeADR {
		broken = filterADRPaths(broken)
	}
	if len(only) > 0 {
		broken = filterPathsByTargets(broken, only)
	}
	if len(broken) == 0 {
		return nil
	}
	for _, path := range broken {
		fmt.Fprintf(w, "BROKEN_SEAL %s\n", path)
	}
	return fmt.Errorf("verify failed: canonical markdown has broken seals\nhint: resolve the .c3/ text, then run 'c3x repair'")
}

func AutoExportCanonical(s *store.Store, c3Dir string) error {
	if s == nil || c3Dir == "" {
		return nil
	}
	if _, err := os.Stat(c3Dir); err != nil {
		return err
	}
	return RunSyncExport(ExportOptions{Store: s, OutputDir: c3Dir}, io.Discard)
}
