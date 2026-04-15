package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/toon"
)

// stripFrontmatter removes YAML frontmatter (---\n...\n---) from content,
// returning only the body after the closing ---.
func stripFrontmatter(s string) string {
	_, body := frontmatter.ParseFrontmatter(s)
	return body
}

// stripHTMLComments removes all HTML comment blocks from content.
func stripHTMLComments(s string) string {
	re := regexp.MustCompile(`(?s)<!--.*?-->`)
	result := re.ReplaceAllString(s, "")
	// Clean up any resulting blank lines (multiple consecutive newlines → max 2)
	reBlank := regexp.MustCompile(`\n{3,}`)
	return reBlank.ReplaceAllString(result, "\n\n")
}

const defaultTruncateLen = 1500

func isAgentMode() bool {
	return os.Getenv("C3X_MODE") == "agent"
}

func writeJSON(w io.Writer, v any) error {
	if isAgentMode() {
		out, err := toon.MarshalAny(v)
		if err != nil {
			return err
		}
		fmt.Fprint(w, out)
		return nil
	}
	var data []byte
	var err error
	data, err = json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(data))
	return nil
}
