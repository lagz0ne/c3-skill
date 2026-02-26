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
