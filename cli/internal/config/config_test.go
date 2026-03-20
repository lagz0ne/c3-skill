package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindC3Dir(t *testing.T) {
	t.Run("found in current dir", func(t *testing.T) {
		tmp := t.TempDir()
		c3 := filepath.Join(tmp, ".c3")
		if err := os.Mkdir(c3, 0755); err != nil {
			t.Fatal(err)
		}

		got := FindC3Dir(tmp)
		if got != c3 {
			t.Errorf("FindC3Dir() = %q, want %q", got, c3)
		}
	})

	t.Run("found in parent dir", func(t *testing.T) {
		tmp := t.TempDir()
		c3 := filepath.Join(tmp, ".c3")
		if err := os.Mkdir(c3, 0755); err != nil {
			t.Fatal(err)
		}
		child := filepath.Join(tmp, "src", "deep", "nested")
		if err := os.MkdirAll(child, 0755); err != nil {
			t.Fatal(err)
		}

		got := FindC3Dir(child)
		if got != c3 {
			t.Errorf("FindC3Dir() = %q, want %q", got, c3)
		}
	})

	t.Run("not found", func(t *testing.T) {
		tmp := t.TempDir()
		got := FindC3Dir(tmp)
		// Might find one in parent dirs of tmp, but that's OS-dependent.
		// Create a known isolated dir structure.
		isolated := filepath.Join(tmp, "isolated")
		if err := os.Mkdir(isolated, 0755); err != nil {
			t.Fatal(err)
		}
		got = FindC3Dir(isolated)
		// We can't guarantee no .c3 exists above tmp, so just check it's not inside isolated.
		if got != "" && filepath.Dir(got) == isolated {
			t.Errorf("should not find .c3 in isolated dir")
		}
	})

	t.Run("file named .c3 is not a directory", func(t *testing.T) {
		tmp := t.TempDir()
		// Create .c3 as a file, not directory
		if err := os.WriteFile(filepath.Join(tmp, ".c3"), []byte("not a dir"), 0644); err != nil {
			t.Fatal(err)
		}

		got := FindC3Dir(tmp)
		// Should not return a file, should keep searching upward
		if got == filepath.Join(tmp, ".c3") {
			t.Error("FindC3Dir() should not return a file")
		}
	})

	t.Run("override with explicit dir", func(t *testing.T) {
		tmp := t.TempDir()
		custom := filepath.Join(tmp, "my-c3")
		if err := os.Mkdir(custom, 0755); err != nil {
			t.Fatal(err)
		}

		got := ResolveC3Dir("", custom)
		if got != custom {
			t.Errorf("ResolveC3Dir() = %q, want %q", got, custom)
		}
	})

	t.Run("override takes precedence", func(t *testing.T) {
		tmp := t.TempDir()
		// Create both .c3 and a custom dir
		c3 := filepath.Join(tmp, ".c3")
		if err := os.Mkdir(c3, 0755); err != nil {
			t.Fatal(err)
		}
		custom := filepath.Join(tmp, "custom-c3")
		if err := os.Mkdir(custom, 0755); err != nil {
			t.Fatal(err)
		}

		got := ResolveC3Dir(tmp, custom)
		if got != custom {
			t.Errorf("ResolveC3Dir() = %q, want %q (override should win)", got, custom)
		}
	})

	t.Run("empty override uses auto-discovery", func(t *testing.T) {
		tmp := t.TempDir()
		c3 := filepath.Join(tmp, ".c3")
		if err := os.Mkdir(c3, 0755); err != nil {
			t.Fatal(err)
		}

		got := ResolveC3Dir(tmp, "")
		if got != c3 {
			t.Errorf("ResolveC3Dir() = %q, want %q", got, c3)
		}
	})
}

func TestProjectDir(t *testing.T) {
	got := ProjectDir("/home/user/project/.c3")
	want := "/home/user/project"
	if got != want {
		t.Errorf("ProjectDir() = %q, want %q", got, want)
	}
}
