package cmd

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name string
		argv []string
		want Options
	}{
		{
			name: "simple command",
			argv: []string{"list"},
			want: Options{Command: "list"},
		},
		{
			name: "command with flags",
			argv: []string{"list", "--flat", "--json"},
			want: Options{Command: "list", Flat: true, JSON: true},
		},
		{
			name: "add with args",
			argv: []string{"add", "container", "api"},
			want: Options{Command: "add", Args: []string{"container", "api"}},
		},
		{
			name: "add component with container flag",
			argv: []string{"add", "component", "auth", "--container", "c3-1", "--feature"},
			want: Options{Command: "add", Args: []string{"component", "auth"}, Container: "c3-1", Feature: true},
		},
		{
			name: "version flag",
			argv: []string{"-v"},
			want: Options{Version: true},
		},
		{
			name: "help flag",
			argv: []string{"--help"},
			want: Options{Help: true},
		},
		{
			name: "c3-dir flag",
			argv: []string{"list", "--c3-dir", "/tmp/my-c3"},
			want: Options{Command: "list", C3Dir: "/tmp/my-c3"},
		},
		{
			name: "empty args",
			argv: []string{},
			want: Options{},
		},
		{
			name: "include-adr flag",
			argv: []string{"list", "--include-adr"},
			want: Options{Command: "list", IncludeADR: true},
		},
		{
			name: "continue flag",
			argv: []string{"migrate", "--continue"},
			want: Options{Command: "migrate", Continue: true},
		},
		{
			name: "check only repeatable",
			argv: []string{"check", "--only", "c3-101", "--only", "refs/ref-jwt.md"},
			want: Options{Command: "check", Only: []string{"c3-101", "refs/ref-jwt.md"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseArgs(tt.argv)
			if got.Command != tt.want.Command {
				t.Errorf("Command = %q, want %q", got.Command, tt.want.Command)
			}
			if got.JSON != tt.want.JSON {
				t.Errorf("JSON = %v, want %v", got.JSON, tt.want.JSON)
			}
			if got.Flat != tt.want.Flat {
				t.Errorf("Flat = %v, want %v", got.Flat, tt.want.Flat)
			}
			if got.Feature != tt.want.Feature {
				t.Errorf("Feature = %v, want %v", got.Feature, tt.want.Feature)
			}
			if got.Container != tt.want.Container {
				t.Errorf("Container = %q, want %q", got.Container, tt.want.Container)
			}
			if got.C3Dir != tt.want.C3Dir {
				t.Errorf("C3Dir = %q, want %q", got.C3Dir, tt.want.C3Dir)
			}
			if got.Help != tt.want.Help {
				t.Errorf("Help = %v, want %v", got.Help, tt.want.Help)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.want.Version)
			}
			if got.IncludeADR != tt.want.IncludeADR {
				t.Errorf("IncludeADR = %v, want %v", got.IncludeADR, tt.want.IncludeADR)
			}
			if got.Continue != tt.want.Continue {
				t.Errorf("Continue = %v, want %v", got.Continue, tt.want.Continue)
			}
			if len(got.Only) != len(tt.want.Only) {
				t.Errorf("Only len = %d, want %d", len(got.Only), len(tt.want.Only))
			} else {
				for i, only := range tt.want.Only {
					if got.Only[i] != only {
						t.Errorf("Only[%d] = %q, want %q", i, got.Only[i], only)
					}
				}
			}
			if len(got.Args) != len(tt.want.Args) {
				t.Errorf("Args len = %d, want %d", len(got.Args), len(tt.want.Args))
			} else {
				for i, a := range tt.want.Args {
					if got.Args[i] != a {
						t.Errorf("Args[%d] = %q, want %q", i, got.Args[i], a)
					}
				}
			}
		})
	}
}

func TestParseArgs_Extended(t *testing.T) {
	tests := []struct {
		name  string
		argv  []string
		check func(t *testing.T, got Options)
	}{
		{
			name: "section and append",
			argv: []string{"set", "c3-101", "--section", "Goal", "--append"},
			check: func(t *testing.T, got Options) {
				if got.Section != "Goal" {
					t.Errorf("Section = %q", got.Section)
				}
				if !got.Append {
					t.Error("Append should be true")
				}
			},
		},
		{
			name: "field flag",
			argv: []string{"set", "--field", "goal"},
			check: func(t *testing.T, got Options) {
				if got.Field != "goal" {
					t.Errorf("Field = %q", got.Field)
				}
			},
		},
		{
			name: "depth flag",
			argv: []string{"graph", "c3-1", "--depth", "3"},
			check: func(t *testing.T, got Options) {
				if got.Depth != 3 {
					t.Errorf("Depth = %d", got.Depth)
				}
			},
		},
		{
			name: "direction flag",
			argv: []string{"graph", "c3-1", "--direction", "forward"},
			check: func(t *testing.T, got Options) {
				if got.Direction != "forward" {
					t.Errorf("Direction = %q", got.Direction)
				}
			},
		},
		{
			name: "format flag",
			argv: []string{"graph", "c3-1", "--format", "mermaid"},
			check: func(t *testing.T, got Options) {
				if got.Format != "mermaid" {
					t.Errorf("Format = %q", got.Format)
				}
			},
		},
		{
			name: "type filter",
			argv: []string{"query", "auth", "--type", "component"},
			check: func(t *testing.T, got Options) {
				if got.TypeFilter != "component" {
					t.Errorf("TypeFilter = %q", got.TypeFilter)
				}
			},
		},
		{
			name: "mark flag",
			argv: []string{"diff", "--mark"},
			check: func(t *testing.T, got Options) {
				if !got.Mark {
					t.Error("Mark should be true")
				}
			},
		},
		{
			name: "keep-originals",
			argv: []string{"migrate", "--keep-originals"},
			check: func(t *testing.T, got Options) {
				if !got.KeepOriginals {
					t.Error("KeepOriginals should be true")
				}
			},
		},
		{
			name: "limit flag",
			argv: []string{"query", "auth", "--limit", "5"},
			check: func(t *testing.T, got Options) {
				if got.Limit != 5 {
					t.Errorf("Limit = %d", got.Limit)
				}
			},
		},
		{
			name: "compact flag",
			argv: []string{"list", "--compact"},
			check: func(t *testing.T, got Options) {
				if !got.Compact {
					t.Error("Compact should be true")
				}
			},
		},
		{
			name: "fix flag",
			argv: []string{"check", "--fix"},
			check: func(t *testing.T, got Options) {
				if !got.Fix {
					t.Error("Fix should be true")
				}
			},
		},
		{
			name: "remove flag",
			argv: []string{"wire", "--remove", "c3-101", "ref-jwt"},
			check: func(t *testing.T, got Options) {
				if !got.Remove {
					t.Error("Remove should be true")
				}
			},
		},
		{
			name: "dry-run flag",
			argv: []string{"delete", "c3-101", "--dry-run"},
			check: func(t *testing.T, got Options) {
				if !got.DryRun {
					t.Error("DryRun should be true")
				}
			},
		},
		{
			name: "short help",
			argv: []string{"-h"},
			check: func(t *testing.T, got Options) {
				if !got.Help {
					t.Error("Help should be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseArgs(tt.argv)
			tt.check(t, got)
		})
	}
}

func TestParseArgs_C3XMode(t *testing.T) {
	t.Setenv("C3X_MODE", "agent")
	got := ParseArgs([]string{"list"})
	if !got.JSON {
		t.Error("C3X_MODE=agent should request machine output")
	}
	if got.JSONExplicit {
		t.Error("C3X_MODE=agent should not mark --json explicit")
	}
}
