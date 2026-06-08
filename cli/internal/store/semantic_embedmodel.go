//go:build embedmodel

package store

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed semantic_model/model.onnx semantic_model/vocab.txt
var embeddedSemanticModelFS embed.FS

func materializeEmbeddedSemanticAssets(modelPath, vocabPath string) (bool, error) {
	if err := writeEmbeddedSemanticAsset(modelPath, "semantic_model/model.onnx", 0644); err != nil {
		return false, err
	}
	if err := writeEmbeddedSemanticAsset(vocabPath, "semantic_model/vocab.txt", 0644); err != nil {
		return false, err
	}
	return true, nil
}

func writeEmbeddedSemanticAsset(target, name string, mode os.FileMode) error {
	if fileExistsAndNonEmpty(target) {
		return nil
	}
	data, err := embeddedSemanticModelFS.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read embedded semantic asset %s: %w", name, err)
	}
	if len(data) == 0 {
		return fmt.Errorf("embedded semantic asset %s is empty", name)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}
