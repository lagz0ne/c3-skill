package cmd

import (
	"fmt"
	"io"
)

type TemplateOptions struct {
	C3Dir         string
	JSON          bool
	Sub           string
	ID            string
	Body          io.Reader
	StdinTerminal bool
}

func RunTemplate(_ TemplateOptions, _ io.Writer) error {
	return fmt.Errorf("error: c3x template has been retired\nhint: use c3x canvas list, c3x canvas read adr, or c3x canvas write adr --file canvas.md")
}
