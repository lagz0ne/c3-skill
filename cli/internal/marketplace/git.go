package marketplace

import (
	"fmt"
	"os/exec"
)

// Clone performs a shallow git clone of url into dest.
func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", url, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone %s: %w\n%s", url, err, out)
	}
	return nil
}

// Pull runs git pull in the given directory.
func Pull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull", "--ff-only")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull in %s: %w\n%s", dir, err, out)
	}
	return nil
}
