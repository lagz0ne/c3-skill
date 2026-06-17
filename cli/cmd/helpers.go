package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/lagz0ne/c3-design/cli/internal/toon"
)

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
