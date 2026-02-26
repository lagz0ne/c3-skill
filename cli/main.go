package main

import (
	"fmt"
	"os"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/config"
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

	docs, err := walker.WalkC3Docs(c3Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: walking .c3/: %v\n", err)
		os.Exit(1)
	}
	graph := walker.BuildGraph(docs)

	// Resolve project root (parent of .c3/)
	projectDir := config.ProjectDir(c3Dir)

	switch opts.Command {
	case "list":
		err = cmd.RunList(graph, opts.JSON, opts.Flat, w)
	case "check":
		checkOpts := cmd.CheckOptions{
			Graph:      graph,
			Docs:       docs,
			JSON:       opts.JSON,
			ProjectDir: projectDir,
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
			err = cmd.RunAddRich(addOpts, w)
		} else {
			err = cmd.RunAdd(entityType, slug, c3Dir, graph, opts.Container, opts.Feature, w)
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
	case "wire":
		source, relation, target := "", "", ""
		if len(opts.Args) >= 1 {
			source = opts.Args[0]
		}
		if len(opts.Args) >= 2 {
			relation = opts.Args[1]
		}
		if len(opts.Args) >= 3 {
			target = opts.Args[2]
		}
		err = cmd.RunWire(c3Dir, source, relation, target, w)
	case "unwire":
		source, relation, target := "", "", ""
		if len(opts.Args) >= 1 {
			source = opts.Args[0]
		}
		if len(opts.Args) >= 2 {
			relation = opts.Args[1]
		}
		if len(opts.Args) >= 3 {
			target = opts.Args[2]
		}
		err = cmd.RunUnwire(c3Dir, source, relation, target, w)
	case "schema":
		entityType := ""
		if len(opts.Args) >= 1 {
			entityType = opts.Args[0]
		}
		err = cmd.RunSchema(entityType, opts.JSON, w)
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
