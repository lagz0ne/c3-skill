//go:build embedmodel

package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbedModelMaterializesModelAssets(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.onnx")
	vocabPath := filepath.Join(dir, "vocab.txt")

	ok, err := materializeEmbeddedSemanticAssets(modelPath, vocabPath)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("embedmodel build should materialize embedded assets")
	}
	for _, path := range []string{modelPath, vocabPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() == 0 {
			t.Fatalf("%s should not be empty", path)
		}
	}
}
