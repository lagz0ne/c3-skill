//go:build !cgo

package store

import (
	"context"
	"fmt"
)

type onnxMiniLMProvider struct{}

func (p *onnxMiniLMProvider) Embed(ctx context.Context, text string, allowDownload bool) ([]float32, bool, error) {
	return nil, false, fmt.Errorf("%w: ONNX semantic search requires a CGO-enabled build", ErrSemanticUnavailable)
}
