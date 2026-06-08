//go:build !embedmodel

package store

func materializeEmbeddedSemanticAssets(modelPath, vocabPath string) (bool, error) {
	return false, nil
}

func materializeEmbeddedSemanticRuntime(runtimePath, libName string) (bool, error) {
	return false, nil
}
