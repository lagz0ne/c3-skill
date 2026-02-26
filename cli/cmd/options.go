package cmd

// Options holds parsed CLI flags and arguments.
type Options struct {
	Command   string
	Args      []string
	JSON      bool
	Flat      bool
	Feature   bool
	Container string
	C3Dir     string
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
