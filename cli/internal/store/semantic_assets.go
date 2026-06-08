package store

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	semanticHFRevision     = "1110a243fdf4706b3f48f1d95db1a4f5529b4d41"
	semanticHFRepo         = "sentence-transformers/all-MiniLM-L6-v2"
	semanticONNXRuntimeVer = "1.26.0"
	semanticHTTPTimeout    = 15 * time.Minute
)

const semanticModelAsset = "c3x-semantic-model-all-MiniLM-L6-v2-" + semanticHFRevision + ".onnx"
const semanticVocabAsset = "c3x-semantic-vocab-all-MiniLM-L6-v2-" + semanticHFRevision + ".txt"

const (
	semanticModelHFPath = "onnx/model.onnx"
	semanticVocabHFPath = "vocab.txt"
	semanticModelSHA256 = "6fd5d72fe4589f189f8ebc006442dbb529bb7ce38f8082112682524616046452"
	semanticVocabSHA256 = "07eced375cec144d27c900241f3e339478dec958f92fddbc551f295c992038a3"
)

type semanticModelSource struct {
	assetName string
	cacheName string
	hfPath    string
	url       string
	sha256    string
}

type semanticAssets struct {
	ModelPath      string
	VocabPath      string
	RuntimeLibPath string
	CacheDir       string
}

type runtimeAsset struct {
	URL       string
	Archive   string
	MemberEnd string
	LibName   string
	Platform  string
}

// SemanticCacheDir returns the cache root used for downloaded semantic assets.
func SemanticCacheDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("C3_SEMANTIC_CACHE_DIR")); dir != "" {
		return dir, nil
	}
	base := strings.TrimSpace(os.Getenv("XDG_CACHE_HOME"))
	if base == "" {
		var err error
		base, err = os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("semantic cache dir: %w", err)
		}
	}
	version := strings.TrimSpace(os.Getenv("C3X_VERSION"))
	if version == "" {
		version = "dev"
	}
	return filepath.Join(base, "c3x", version), nil
}

func ensureSemanticAssets(ctx context.Context, allowDownload bool) (semanticAssets, error) {
	if os.Getenv("C3_SEMANTIC_OFFLINE") != "" {
		allowDownload = false
	}
	cacheDir, err := SemanticCacheDir()
	if err != nil {
		return semanticAssets{}, err
	}
	modelDir := semanticModelCacheDir(cacheDir)
	asset, err := currentRuntimeAsset()
	if err != nil {
		return semanticAssets{}, err
	}
	runtimeDir := semanticRuntimeCacheDir(cacheDir, asset)
	assets := semanticAssets{
		ModelPath:      filepath.Join(modelDir, "model.onnx"),
		VocabPath:      filepath.Join(modelDir, "vocab.txt"),
		RuntimeLibPath: filepath.Join(runtimeDir, asset.LibName),
		CacheDir:       cacheDir,
	}

	if err := ensureSemanticModelFiles(ctx, assets, allowDownload); err != nil {
		return semanticAssets{}, err
	}
	if fileExistsAndNonEmpty(assets.RuntimeLibPath) {
		return assets, nil
	}
	if ok, err := materializeEmbeddedSemanticRuntime(assets.RuntimeLibPath, asset.LibName); err != nil {
		return semanticAssets{}, err
	} else if ok {
		return assets, nil
	}
	if !allowDownload {
		return semanticAssets{}, ErrSemanticUnavailable
	}
	if err := downloadRuntimeLib(ctx, asset, runtimeDir, assets.RuntimeLibPath); err != nil {
		return semanticAssets{}, fmt.Errorf("%w: download onnxruntime: %w", ErrSemanticUnavailable, err)
	}
	return assets, nil
}

func ensureSemanticModelFiles(ctx context.Context, assets semanticAssets, allowDownload bool) error {
	if ok, err := materializeEmbeddedSemanticAssets(assets.ModelPath, assets.VocabPath); err != nil {
		return err
	} else if ok {
		return nil
	}
	sources := semanticModelSources()
	if err := ensureCanonicalSemanticFile(ctx, assets.ModelPath, sources[0], allowDownload); err != nil {
		return err
	}
	return ensureCanonicalSemanticFile(ctx, assets.VocabPath, sources[1], allowDownload)
}

// PrepareEmbeddedSemanticAssets writes verified canonical model and runtime
// files into the go:embed directory used by fat release builds.
func PrepareEmbeddedSemanticAssets(ctx context.Context, embedDir, targetOS, targetArch string) error {
	if err := PrepareEmbeddedSemanticModelAssets(ctx, embedDir); err != nil {
		return err
	}
	asset, err := runtimeAssetFor(targetOS, targetArch)
	if err != nil {
		return err
	}
	cacheDir, err := SemanticCacheDir()
	if err != nil {
		return err
	}
	return prepareEmbeddedSemanticRuntime(ctx, embedDir, asset, semanticModelCacheCandidateDirs(cacheDir))
}

// PrepareEmbeddedSemanticModelAssets writes the verified canonical model files
// into the go:embed directory used by fat release builds.
func PrepareEmbeddedSemanticModelAssets(ctx context.Context, embedDir string) error {
	cacheDir, err := SemanticCacheDir()
	if err != nil {
		return err
	}
	return prepareEmbeddedSemanticModelAssets(ctx, embedDir, semanticModelCacheCandidateDirs(cacheDir), semanticModelSources())
}

// PrepareReleaseSemanticModelAssets writes verified model release assets and
// matching .sha256 files for distribution packaging.
func PrepareReleaseSemanticModelAssets(ctx context.Context, outDir string) error {
	cacheDir, err := SemanticCacheDir()
	if err != nil {
		return err
	}
	return prepareReleaseSemanticModelAssets(ctx, outDir, semanticModelCacheCandidateDirs(cacheDir), semanticModelSources())
}

func semanticModelSources() []semanticModelSource {
	return []semanticModelSource{
		{
			assetName: semanticModelAsset,
			cacheName: "model.onnx",
			hfPath:    semanticModelHFPath,
			sha256:    semanticModelSHA256,
		},
		{
			assetName: semanticVocabAsset,
			cacheName: "vocab.txt",
			hfPath:    semanticVocabHFPath,
			sha256:    semanticVocabSHA256,
		},
	}
}

func semanticModelCacheDir(cacheDir string) string {
	return filepath.Join(cacheDir, "models", "all-MiniLM-L6-v2-"+semanticHFRevision)
}

func semanticRuntimeCacheDir(cacheDir string, asset runtimeAsset) string {
	return filepath.Join(cacheDir, "onnxruntime", semanticONNXRuntimeVer, asset.Platform)
}

func semanticModelSourceURL(source semanticModelSource) string {
	if source.url != "" {
		return source.url
	}
	return "https://huggingface.co/" + semanticHFRepo + "/resolve/" + semanticHFRevision + "/" + source.hfPath
}

func semanticModelCacheCandidateDirs(primary string) []string {
	dirs := []string{primary}
	if legacy, err := legacySemanticCacheDir(); err == nil && legacy != "" && legacy != primary {
		dirs = append(dirs, legacy)
	}
	return dirs
}

func legacySemanticCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("legacy semantic cache dir: %w", err)
	}
	return filepath.Join(base, "c3", "semantic"), nil
}

func prepareEmbeddedSemanticModelAssets(ctx context.Context, embedDir string, cacheDirs []string, sources []semanticModelSource) error {
	if err := os.MkdirAll(embedDir, 0755); err != nil {
		return err
	}
	for _, source := range sources {
		cachePath, err := ensureCanonicalSemanticCachedFile(ctx, source, cacheDirs)
		if err != nil {
			return err
		}
		target := filepath.Join(embedDir, source.cacheName)
		if err := copyVerifiedSemanticFile(cachePath, target, source.sha256); err != nil {
			return fmt.Errorf("stage embedded semantic asset %s: %w", source.assetName, err)
		}
	}
	return nil
}

func prepareReleaseSemanticModelAssets(ctx context.Context, outDir string, cacheDirs []string, sources []semanticModelSource) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	for _, source := range sources {
		cachePath, err := ensureCanonicalSemanticCachedFile(ctx, source, cacheDirs)
		if err != nil {
			return err
		}
		target := filepath.Join(outDir, source.assetName)
		if err := copyVerifiedSemanticFile(cachePath, target, source.sha256); err != nil {
			return fmt.Errorf("write release semantic asset %s: %w", source.assetName, err)
		}
		sumPath := target + ".sha256"
		if err := os.WriteFile(sumPath, []byte(source.sha256+"  "+source.assetName+"\n"), 0644); err != nil {
			return fmt.Errorf("write checksum %s: %w", filepath.Base(sumPath), err)
		}
	}
	return nil
}

func prepareEmbeddedSemanticRuntime(ctx context.Context, embedDir string, asset runtimeAsset, cacheDirs []string) error {
	cachePath, err := ensureSemanticRuntimeCachedFile(ctx, asset, cacheDirs)
	if err != nil {
		return err
	}
	target := filepath.Join(embedDir, "runtime", asset.LibName)
	if err := copySemanticFile(cachePath, target); err != nil {
		return fmt.Errorf("stage embedded runtime asset %s: %w", asset.LibName, err)
	}
	return nil
}

func ensureSemanticRuntimeCachedFile(ctx context.Context, asset runtimeAsset, cacheDirs []string) (string, error) {
	if len(cacheDirs) == 0 {
		return "", errors.New("semantic cache candidates: empty")
	}
	primary := filepath.Join(semanticRuntimeCacheDir(cacheDirs[0], asset), asset.LibName)
	for _, cacheDir := range cacheDirs {
		candidate := filepath.Join(semanticRuntimeCacheDir(cacheDir, asset), asset.LibName)
		if !fileExistsAndNonEmpty(candidate) {
			continue
		}
		if candidate != primary {
			if err := copySemanticFile(candidate, primary); err != nil {
				return "", fmt.Errorf("copy runtime asset from legacy cache: %w", err)
			}
		}
		return primary, nil
	}
	if err := downloadRuntimeLib(ctx, asset, filepath.Dir(primary), primary); err != nil {
		return "", err
	}
	return primary, nil
}

func ensureCanonicalSemanticCachedFile(ctx context.Context, source semanticModelSource, cacheDirs []string) (string, error) {
	if len(cacheDirs) == 0 {
		return "", errors.New("semantic cache candidates: empty")
	}
	primary := filepath.Join(semanticModelCacheDir(cacheDirs[0]), source.cacheName)
	for _, cacheDir := range cacheDirs {
		candidate := filepath.Join(semanticModelCacheDir(cacheDir), source.cacheName)
		if err := verifyFileSHA256(candidate, source.sha256); err != nil {
			continue
		}
		if candidate != primary {
			if err := copyVerifiedSemanticFile(candidate, primary, source.sha256); err != nil {
				return "", fmt.Errorf("copy semantic asset from legacy cache: %w", err)
			}
		}
		return primary, nil
	}
	if err := ensureCanonicalSemanticFile(ctx, primary, source, true); err != nil {
		return "", err
	}
	return primary, nil
}

func ensureCanonicalSemanticFile(ctx context.Context, path string, source semanticModelSource, allowDownload bool) error {
	if err := verifyFileSHA256(path, source.sha256); err == nil {
		return nil
	}
	if !allowDownload {
		return ErrSemanticUnavailable
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := downloadURL(ctx, semanticModelSourceURL(source), tmp); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("%w: download semantic asset %s: %w\nhint: connect once to Hugging Face or use the fat C3 build for offline semantic search", ErrSemanticUnavailable, source.assetName, err)
	}
	if err := verifyFileSHA256(tmp, source.sha256); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("%w: verify semantic asset %s: %w", ErrSemanticUnavailable, source.assetName, err)
	}
	return os.Rename(tmp, path)
}

func copyVerifiedSemanticFile(source, target, wantHex string) error {
	if err := verifyFileSHA256(source, wantHex); err != nil {
		return fmt.Errorf("verify source %s: %w", source, err)
	}
	if err := copySemanticFile(source, target); err != nil {
		return err
	}
	if err := verifyFileSHA256(target, wantHex); err != nil {
		return fmt.Errorf("verify copied %s: %w", target, err)
	}
	return nil
}

func copySemanticFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	tmp := target + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, target)
}

func currentRuntimeAsset() (runtimeAsset, error) {
	return runtimeAssetFor(runtime.GOOS, runtime.GOARCH)
}

func runtimeAssetFor(goos, goarch string) (runtimeAsset, error) {
	base := "https://github.com/microsoft/onnxruntime/releases/download/v" + semanticONNXRuntimeVer + "/"
	switch goos + "/" + goarch {
	case "linux/amd64":
		name := "onnxruntime-linux-x64-" + semanticONNXRuntimeVer + ".tgz"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "tgz",
			MemberEnd: "/lib/libonnxruntime.so." + semanticONNXRuntimeVer,
			LibName:   "libonnxruntime.so." + semanticONNXRuntimeVer,
			Platform:  "linux-amd64",
		}, nil
	case "linux/arm64":
		name := "onnxruntime-linux-aarch64-" + semanticONNXRuntimeVer + ".tgz"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "tgz",
			MemberEnd: "/lib/libonnxruntime.so." + semanticONNXRuntimeVer,
			LibName:   "libonnxruntime.so." + semanticONNXRuntimeVer,
			Platform:  "linux-arm64",
		}, nil
	case "darwin/amd64":
		name := "onnxruntime-osx-x86_64-" + semanticONNXRuntimeVer + ".tgz"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "tgz",
			MemberEnd: "/lib/libonnxruntime." + semanticONNXRuntimeVer + ".dylib",
			LibName:   "libonnxruntime." + semanticONNXRuntimeVer + ".dylib",
			Platform:  "darwin-amd64",
		}, nil
	case "darwin/arm64":
		name := "onnxruntime-osx-arm64-" + semanticONNXRuntimeVer + ".tgz"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "tgz",
			MemberEnd: "/lib/libonnxruntime." + semanticONNXRuntimeVer + ".dylib",
			LibName:   "libonnxruntime." + semanticONNXRuntimeVer + ".dylib",
			Platform:  "darwin-arm64",
		}, nil
	case "windows/amd64":
		name := "onnxruntime-win-x64-" + semanticONNXRuntimeVer + ".zip"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "zip",
			MemberEnd: "/lib/onnxruntime.dll",
			LibName:   "onnxruntime.dll",
			Platform:  "windows-amd64",
		}, nil
	case "windows/arm64":
		name := "onnxruntime-win-arm64-" + semanticONNXRuntimeVer + ".zip"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "zip",
			MemberEnd: "/lib/onnxruntime.dll",
			LibName:   "onnxruntime.dll",
			Platform:  "windows-arm64",
		}, nil
	default:
		return runtimeAsset{}, fmt.Errorf("%w: unsupported platform %s/%s", ErrSemanticUnavailable, goos, goarch)
	}
}

func downloadRuntimeLib(ctx context.Context, asset runtimeAsset, dir, target string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "onnxruntime-*."+asset.Archive)
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	defer os.Remove(tmpPath)

	if err := downloadURL(ctx, asset.URL, tmpPath); err != nil {
		return err
	}
	switch asset.Archive {
	case "tgz":
		return extractTarGzMember(tmpPath, asset.MemberEnd, target)
	case "zip":
		return extractZipMember(tmpPath, asset.MemberEnd, target)
	default:
		return fmt.Errorf("unsupported runtime archive %q", asset.Archive)
	}
}

func downloadURL(ctx context.Context, url, target string) error {
	body, err := downloadBytes(ctx, url)
	if err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := out.Write(body); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if !fileExistsAndNonEmpty(target) {
		return fmt.Errorf("download %s: empty file", url)
	}
	return nil
}

func downloadBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: semanticHTTPTimeout}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("download %s: status %s", url, res.Status)
	}
	return io.ReadAll(res.Body)
}

func verifyFileSHA256(path, wantHex string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, wantHex) {
		return errors.New("sha256 mismatch")
	}
	return nil
}

func extractTarGzMember(archivePath, memberEnd, target string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag != tar.TypeReg || !strings.HasSuffix(filepath.ToSlash(header.Name), memberEnd) {
			continue
		}
		return writeExtractedFile(target, tr, 0755)
	}
	return fmt.Errorf("runtime library %s not found in %s", memberEnd, archivePath)
}

func extractZipMember(archivePath, memberEnd, target string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !strings.HasSuffix(filepath.ToSlash(f.Name), memberEnd) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = writeExtractedFile(target, rc, 0755)
		_ = rc.Close()
		return err
	}
	return fmt.Errorf("runtime library %s not found in %s", memberEnd, archivePath)
}

func writeExtractedFile(target string, r io.Reader, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	tmp := target + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Chmod(tmp, mode); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, target)
}

func fileExistsAndNonEmpty(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}
