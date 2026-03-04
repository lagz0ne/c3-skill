package cmd

// Options holds parsed CLI flags and arguments.
type Options struct {
	Command   string
	Args      []string
	JSON      bool
	Flat      bool
	Compact   bool
	Feature   bool
	Append    bool
	Chain     bool
	Container string
	C3Dir     string
	Goal      string
	Summary   string
	Boundary  string
	Field     string
	Section   string
	Help      bool
	Version   bool
}

// ParseArgs parses command-line arguments into Options.
func ParseArgs(argv []string) Options {
	var opts Options
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
		case "--goal":
			if i+1 < len(argv) {
				i++
				opts.Goal = argv[i]
			}
		case "--summary":
			if i+1 < len(argv) {
				i++
				opts.Summary = argv[i]
			}
		case "--boundary":
			if i+1 < len(argv) {
				i++
				opts.Boundary = argv[i]
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
		case "--chain":
			opts.Chain = true
		default:
			args = append(args, arg)
		}
	}

	if len(args) > 0 {
		opts.Command = args[0]
		opts.Args = args[1:]
	}
	return opts
}
