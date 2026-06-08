//go:build !embedmodel

package store

func materializeEmbeddedSemanticAssets(modelPath, vocabPath string) (bool, error) {
	return false, nil
}
