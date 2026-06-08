//go:build !embedmodel

package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSemanticCacheDirUsesVersionedC3XCache(t *testing.T) {
	root := t.TempDir()
	t.Setenv("C3_SEMANTIC_CACHE_DIR", "")
	t.Setenv("XDG_CACHE_HOME", root)
	t.Setenv("C3X_VERSION", "9.9.1")

	got, err := SemanticCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "c3x", "9.9.1")
	if got != want {
		t.Fatalf("SemanticCacheDir() = %q, want %q", got, want)
	}
}

func TestEnsureCanonicalSemanticFileDownloadsPinnedAssetWithChecksum(t *testing.T) {
	data := []byte("model bytes")
	source := testSemanticModelSource(t, "model.onnx", data)

	target := filepath.Join(t.TempDir(), "model.onnx")
	if err := ensureCanonicalSemanticFile(context.Background(), target, source, true); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, target, data)
}

func TestSemanticModelSourceURLUsesPinnedHuggingFaceRevision(t *testing.T) {
	source := semanticModelSources()[0]
	want := "https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/" + semanticHFRevision + "/onnx/model.onnx"
	if got := semanticModelSourceURL(source); got != want {
		t.Fatalf("semanticModelSourceURL() = %q, want %q", got, want)
	}
}

func TestEnsureCanonicalSemanticFileDownloadErrorIsSemanticUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	source := testSourceWithURL("model.onnx", "model.onnx", []byte("model bytes"), server.URL+"/model.onnx")
	target := filepath.Join(t.TempDir(), "model.onnx")
	err := ensureCanonicalSemanticFile(context.Background(), target, source, true)
	if !errors.Is(err, ErrSemanticUnavailable) {
		t.Fatalf("err = %v, want ErrSemanticUnavailable", err)
	}
	if !strings.Contains(err.Error(), "Hugging Face") {
		t.Fatalf("err = %v, want Hugging Face hint", err)
	}
}

func TestEnsureCanonicalSemanticFileRejectsChecksumMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("actual bytes"))
	}))
	t.Cleanup(server.Close)

	source := semanticModelSource{
		assetName: "model.onnx",
		cacheName: "model.onnx",
		url:       server.URL + "/model.onnx",
		sha256:    fmt.Sprintf("%064x", 1),
	}
	target := filepath.Join(t.TempDir(), "model.onnx")
	err := ensureCanonicalSemanticFile(context.Background(), target, source, true)
	if err == nil {
		t.Fatal("expected checksum failure")
	}
	if !errors.Is(err, ErrSemanticUnavailable) {
		t.Fatalf("err = %v, want ErrSemanticUnavailable", err)
	}
	if !strings.Contains(err.Error(), "sha256 mismatch") {
		t.Fatalf("expected sha256 mismatch, got %v", err)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("target should not be committed on checksum failure, stat err = %v", statErr)
	}
}

func TestPrepareEmbeddedSemanticModelAssetsUsesVerifiedLegacyCache(t *testing.T) {
	model := []byte("model bytes")
	vocab := []byte("[PAD]\n[UNK]\n")
	sources := []semanticModelSource{
		testSource("model-asset.onnx", "model.onnx", model),
		testSource("vocab-asset.txt", "vocab.txt", vocab),
	}

	primary := t.TempDir()
	legacy := t.TempDir()
	legacyModelDir := semanticModelCacheDir(legacy)
	if err := os.MkdirAll(legacyModelDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyModelDir, "model.onnx"), model, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyModelDir, "vocab.txt"), vocab, 0644); err != nil {
		t.Fatal(err)
	}

	embedDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(embedDir, "model.onnx"), []byte("stub"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(embedDir, "vocab.txt"), []byte("stub"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := prepareEmbeddedSemanticModelAssets(context.Background(), embedDir, []string{primary, legacy}, sources); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, filepath.Join(embedDir, "model.onnx"), model)
	assertFileBytes(t, filepath.Join(embedDir, "vocab.txt"), vocab)
	assertFileBytes(t, filepath.Join(semanticModelCacheDir(primary), "model.onnx"), model)
	assertFileBytes(t, filepath.Join(semanticModelCacheDir(primary), "vocab.txt"), vocab)
}

func TestPrepareReleaseSemanticModelAssetsWritesChecksums(t *testing.T) {
	model := []byte("model bytes")
	source := testSource("model-asset.onnx", "model.onnx", model)
	cacheDir := t.TempDir()
	modelDir := semanticModelCacheDir(cacheDir)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "model.onnx"), model, 0644); err != nil {
		t.Fatal(err)
	}

	outDir := t.TempDir()
	if err := prepareReleaseSemanticModelAssets(context.Background(), outDir, []string{cacheDir}, []semanticModelSource{source}); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, filepath.Join(outDir, "model-asset.onnx"), model)
	checksum, err := os.ReadFile(filepath.Join(outDir, "model-asset.onnx.sha256"))
	if err != nil {
		t.Fatal(err)
	}
	want := source.sha256 + "  model-asset.onnx\n"
	if string(checksum) != want {
		t.Fatalf("checksum = %q, want %q", string(checksum), want)
	}
}

func TestPrepareEmbeddedSemanticRuntimeUsesCache(t *testing.T) {
	runtimeBytes := []byte("runtime bytes")
	asset := runtimeAsset{
		LibName:  "libonnxruntime.test",
		Platform: "linux-test",
	}
	cacheDir := t.TempDir()
	runtimeDir := semanticRuntimeCacheDir(cacheDir, asset)
	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtimeDir, asset.LibName), runtimeBytes, 0755); err != nil {
		t.Fatal(err)
	}

	embedDir := t.TempDir()
	if err := prepareEmbeddedSemanticRuntime(context.Background(), embedDir, asset, []string{cacheDir}); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, filepath.Join(embedDir, "runtime", asset.LibName), runtimeBytes)
}

func testSemanticModelSource(t *testing.T, name string, data []byte) semanticModelSource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	return testSourceWithURL(name, name, data, server.URL+"/"+name)
}

func testSource(assetName, cacheName string, data []byte) semanticModelSource {
	return testSourceWithURL(assetName, cacheName, data, "")
}

func testSourceWithURL(assetName, cacheName string, data []byte, url string) semanticModelSource {
	sum := sha256.Sum256(data)
	return semanticModelSource{
		assetName: assetName,
		cacheName: cacheName,
		url:       url,
		sha256:    hex.EncodeToString(sum[:]),
	}
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s = %q, want %q", path, string(got), string(want))
	}
}
