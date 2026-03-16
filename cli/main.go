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
	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/config"
	"github.com/lagz0ne/c3-design/cli/internal/index"
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

	// init is special — creates .c3/, no graph needed
	if opts.Command == "init" {
		if err := cmd.RunInit(mustCwd(), w); err != nil {
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

	walkResult, err := walker.WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: walking .c3/: %v\n", err)
		os.Exit(1)
	}
	docs := walkResult.Docs
	graph := walker.BuildGraph(docs)

	// Resolve project root (parent of .c3/)
	projectDir := config.ProjectDir(c3Dir)

	switch opts.Command {
	case "list":
		err = cmd.RunList(cmd.ListOptions{Graph: graph, JSON: opts.JSON, Flat: opts.Flat, Compact: opts.Compact, C3Dir: c3Dir, IncludeADR: opts.IncludeADR}, w)
	case "check":
		checkOpts := cmd.CheckOptions{
			Graph:         graph,
			Docs:          docs,
			JSON:          opts.JSON,
			ProjectDir:    projectDir,
			C3Dir:         c3Dir,
			ParseWarnings: walkResult.Warnings,
			IncludeADR:    opts.IncludeADR,
			Fix:           opts.Fix,
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
				Graph:      graph,
				Container:  opts.Container,
				Feature:    opts.Feature,
				Goal:       opts.Goal,
				Summary:    opts.Summary,
				Boundary:   opts.Boundary,
			}
			err = cmd.RunAddRich(addOpts, addW)
		} else {
			err = cmd.RunAdd(entityType, slug, c3Dir, graph, opts.Container, opts.Feature, addW)
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
			err = cmd.RunUnwire(c3Dir, source, relation, target, w)
		} else {
			err = cmd.RunWire(c3Dir, source, relation, target, w)
		}
	case "lookup":
		if len(opts.Args) < 1 {
			fmt.Fprintln(os.Stderr, "error: lookup requires a <file-path> argument")
			fmt.Fprintln(os.Stderr, "hint: run 'c3x lookup --help' for usage")
			os.Exit(1)
		}
		err = cmd.RunLookup(cmd.LookupOptions{
			Graph:      graph,
			FilePath:   opts.Args[0],
			JSON:       opts.JSON,
			ProjectDir: projectDir,
			C3Dir:      c3Dir,
		}, w)
	case "codemap":
		err = cmd.RunCodemap(cmd.CodemapOptions{
			C3Dir: c3Dir,
			Graph: graph,
			JSON:  opts.JSON,
		}, w)
	case "coverage":
		err = cmd.RunCoverage(cmd.CoverageOptions{
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
			Graph:     graph,
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
			Graph:  graph,
			DryRun: opts.DryRun,
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

	// Rebuild structural index after mutating commands (best-effort, silent)
	switch opts.Command {
	case "add", "set", "wire", "unwire", "delete":
		cm, _ := codemap.ParseCodeMap(filepath.Join(c3Dir, "code-map.yaml"))
		idx := index.Build(graph, cm, c3Dir)
		_ = index.WriteTo(c3Dir, idx)
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
