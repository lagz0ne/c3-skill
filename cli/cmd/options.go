package cmd

import (
	"os"
	"strconv"
)

// Options holds parsed CLI flags and arguments.
type Options struct {
	Command       string
	Args          []string
	JSON          bool
	Flat          bool
	Compact       bool
	Feature       bool
	Append        bool
	Container     string
	C3Dir         string
	Field         string
	Section       string
	Help          bool
	Version       bool
	IncludeADR    bool
	Fix           bool
	Remove        bool
	DryRun        bool
	Depth         int
	Direction     string
	Format        string
	TypeFilter    string
	Mark          bool
	KeepOriginals bool
	Stdin         bool
	Limit         int
	Source        string
	Tag           string
	Recompute     bool
	Keep          int
}

// ParseArgs parses command-line arguments into Options.
func ParseArgs(argv []string) Options {
	var opts Options
	opts.Depth = 1
	opts.Limit = 20
	var args []string

	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch arg {
		case "--json":
			opts.JSON = true
		case "--flat":
			opts.Flat = true
		case "--compact":
			opts.Compact = true
		case "--feature":
			opts.Feature = true
		case "-h", "--help":
			opts.Help = true
		case "-v", "--version":
			opts.Version = true
		case "--container":
			if i+1 < len(argv) {
				i++
				opts.Container = argv[i]
			}
		case "--c3-dir":
			if i+1 < len(argv) {
				i++
				opts.C3Dir = argv[i]
			}
		case "--field":
			if i+1 < len(argv) {
				i++
				opts.Field = argv[i]
			}
		case "--section":
			if i+1 < len(argv) {
				i++
				opts.Section = argv[i]
			}
		case "--append":
			opts.Append = true
		case "--include-adr":
			opts.IncludeADR = true
		case "--fix":
			opts.Fix = true
		case "--remove":
			opts.Remove = true
		case "--dry-run":
			opts.DryRun = true
		case "--depth":
			if i+1 < len(argv) {
				i++
				opts.Depth, _ = strconv.Atoi(argv[i])
			}
		case "--direction":
			if i+1 < len(argv) {
				i++
				opts.Direction = argv[i]
			}
		case "--format":
			if i+1 < len(argv) {
				i++
				opts.Format = argv[i]
			}
		case "--type":
			if i+1 < len(argv) {
				i++
				opts.TypeFilter = argv[i]
			}
		case "--stdin":
			opts.Stdin = true
		case "--mark":
			opts.Mark = true
		case "--keep-originals":
			opts.KeepOriginals = true
		case "--limit":
			if i+1 < len(argv) {
				i++
				opts.Limit, _ = strconv.Atoi(argv[i])
			}
		case "--source":
			if i+1 < len(argv) {
				i++
				opts.Source = argv[i]
			}
		case "--tag":
			if i+1 < len(argv) {
				i++
				opts.Tag = argv[i]
			}
		case "--recompute":
			opts.Recompute = true
		case "--keep":
			if i+1 < len(argv) {
				i++
				opts.Keep, _ = strconv.Atoi(argv[i])
			}
		default:
			args = append(args, arg)
		}
	}

	if len(args) > 0 {
		opts.Command = args[0]
		opts.Args = args[1:]
	}
	// C3X_MODE env var: "agent" implies --json for commands that support it.
	// Explicit --json flag takes precedence (already set above).
	if !opts.JSON {
		if mode := os.Getenv("C3X_MODE"); mode == "agent" {
			opts.JSON = true
		}
	}

	return opts
}
