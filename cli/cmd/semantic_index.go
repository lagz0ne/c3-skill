package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

// SemanticIndexOptions configures semantic index rebuild.
type SemanticIndexOptions struct {
	Store *store.Store
	JSON  bool
}

// SemanticIndexOutput is the structured semantic index response.
type SemanticIndexOutput struct {
	Model    string `json:"model"`
	Dims     int    `json:"dims"`
	Count    int    `json:"count"`
	CacheDir string `json:"cache_dir"`
}

// RunSemanticIndex downloads local ONNX assets if needed and rebuilds vectors.
func RunSemanticIndex(opts SemanticIndexOptions, w io.Writer) error {
	if opts.Store == nil {
		return fmt.Errorf("error: semantic index store is required")
	}
	if err := opts.Store.RebuildSemanticIndexWithOptions(context.Background(), store.SemanticIndexOptions{AllowDownload: true}); err != nil {
		return err
	}
	count, err := opts.Store.SemanticIndexCount()
	if err != nil {
		return err
	}
	cacheDir, err := store.SemanticCacheDir()
	if err != nil {
		return err
	}
	out := SemanticIndexOutput{
		Model:    store.SemanticEmbeddingModel,
		Dims:     384,
		Count:    count,
		CacheDir: cacheDir,
	}
	format := ResolveFormat(opts.JSON, isAgentMode())
	if format != FormatHuman {
		return WriteObjectOutput(w, out, format, nil)
	}
	_, err = fmt.Fprintf(w, "Indexed %d entities with %s (%d dims)\nCache: %s\n", out.Count, out.Model, out.Dims, out.CacheDir)
	return err
}
