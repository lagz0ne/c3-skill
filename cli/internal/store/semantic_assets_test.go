//go:build !embedmodel

package store

import (
	"context"
	"crypto/sha256"
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

func TestEnsureSemanticModelFilesDownloadsReleaseAssetsWithChecksums(t *testing.T) {
	model := []byte("model bytes")
	vocab := []byte("[PAD]\n[UNK]\n")
	server := semanticAssetServer(t, map[string][]byte{
		semanticModelAsset: model,
		semanticVocabAsset: vocab,
	})
	t.Setenv("C3_SEMANTIC_RELEASE_BASE_URL", server.URL)

	dir := t.TempDir()
	assets := semanticAssets{
		ModelPath: filepath.Join(dir, "model.onnx"),
		VocabPath: filepath.Join(dir, "vocab.txt"),
	}
	if err := ensureSemanticModelFiles(context.Background(), assets, true); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, assets.ModelPath, model)
	assertFileBytes(t, assets.VocabPath, vocab)
}

func TestEnsureReleaseFileRejectsChecksumMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".sha256") {
			fmt.Fprintf(w, "%064x  %s\n", 1, semanticModelAsset)
			return
		}
		_, _ = w.Write([]byte("actual bytes"))
	}))
	t.Cleanup(server.Close)
	t.Setenv("C3_SEMANTIC_RELEASE_BASE_URL", server.URL)

	target := filepath.Join(t.TempDir(), "model.onnx")
	err := ensureReleaseFile(context.Background(), target, semanticModelAsset, true)
	if err == nil {
		t.Fatal("expected checksum failure")
	}
	if !strings.Contains(err.Error(), "sha256 mismatch") {
		t.Fatalf("expected sha256 mismatch, got %v", err)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("target should not be committed on checksum failure, stat err = %v", statErr)
	}
}

func semanticAssetServer(t *testing.T, assets map[string][]byte) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if strings.HasSuffix(name, ".sha256") {
			assetName := strings.TrimSuffix(name, ".sha256")
			data, ok := assets[assetName]
			if !ok {
				http.NotFound(w, r)
				return
			}
			sum := sha256.Sum256(data)
			fmt.Fprintf(w, "%x  %s\n", sum, assetName)
			return
		}
		data, ok := assets[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	return server
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
