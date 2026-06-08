//go:build embedmodel

package store

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed semantic_model/model.onnx semantic_model/vocab.txt semantic_model/runtime/*
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

func materializeEmbeddedSemanticRuntime(runtimePath, libName string) (bool, error) {
	name := "semantic_model/runtime/" + libName
	data, err := embeddedSemanticModelFS.ReadFile(name)
	if err != nil {
		return false, nil
	}
	if isEmbeddedSemanticStub(data) {
		return false, nil
	}
	if err := writeEmbeddedSemanticAssetData(runtimePath, name, data, 0755); err != nil {
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
	return writeEmbeddedSemanticAssetData(target, name, data, mode)
}

func writeEmbeddedSemanticAssetData(target, name string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	if err := os.Chmod(tmp, mode); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, target)
}

func isEmbeddedSemanticStub(data []byte) bool {
	trimmed := bytes.ToLower(bytes.TrimSpace(data))
	return len(trimmed) == 0 || bytes.HasPrefix(trimmed, []byte("stub"))
}
