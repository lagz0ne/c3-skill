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
	semanticONNXRuntimeVer = "1.26.0"
	semanticHTTPTimeout    = 15 * time.Minute
	semanticReleaseRepo    = "https://github.com/lagz0ne/c3-design/releases/download"
)

const semanticModelAsset = "c3x-semantic-model-all-MiniLM-L6-v2-" + semanticHFRevision + ".onnx"
const semanticVocabAsset = "c3x-semantic-vocab-all-MiniLM-L6-v2-" + semanticHFRevision + ".txt"

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
	modelDir := filepath.Join(cacheDir, "models", "all-MiniLM-L6-v2-"+semanticHFRevision)
	runtimeDir := filepath.Join(cacheDir, "onnxruntime", semanticONNXRuntimeVer, runtime.GOOS+"-"+runtime.GOARCH)
	asset, err := currentRuntimeAsset()
	if err != nil {
		return semanticAssets{}, err
	}
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
	if !allowDownload {
		return semanticAssets{}, ErrSemanticUnavailable
	}
	if err := downloadRuntimeLib(ctx, asset, runtimeDir, assets.RuntimeLibPath); err != nil {
		return semanticAssets{}, err
	}
	return assets, nil
}

func ensureSemanticModelFiles(ctx context.Context, assets semanticAssets, allowDownload bool) error {
	if ok, err := materializeEmbeddedSemanticAssets(assets.ModelPath, assets.VocabPath); err != nil {
		return err
	} else if ok {
		return nil
	}
	if err := ensureReleaseFile(ctx, assets.ModelPath, semanticModelAsset, allowDownload); err != nil {
		return err
	}
	return ensureReleaseFile(ctx, assets.VocabPath, semanticVocabAsset, allowDownload)
}

func currentRuntimeAsset() (runtimeAsset, error) {
	base := "https://github.com/microsoft/onnxruntime/releases/download/v" + semanticONNXRuntimeVer + "/"
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		name := "onnxruntime-linux-x64-" + semanticONNXRuntimeVer + ".tgz"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "tgz",
			MemberEnd: "/lib/libonnxruntime.so." + semanticONNXRuntimeVer,
			LibName:   "libonnxruntime.so." + semanticONNXRuntimeVer,
		}, nil
	case "linux/arm64":
		name := "onnxruntime-linux-aarch64-" + semanticONNXRuntimeVer + ".tgz"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "tgz",
			MemberEnd: "/lib/libonnxruntime.so." + semanticONNXRuntimeVer,
			LibName:   "libonnxruntime.so." + semanticONNXRuntimeVer,
		}, nil
	case "darwin/amd64":
		name := "onnxruntime-osx-x86_64-" + semanticONNXRuntimeVer + ".tgz"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "tgz",
			MemberEnd: "/lib/libonnxruntime." + semanticONNXRuntimeVer + ".dylib",
			LibName:   "libonnxruntime." + semanticONNXRuntimeVer + ".dylib",
		}, nil
	case "darwin/arm64":
		name := "onnxruntime-osx-arm64-" + semanticONNXRuntimeVer + ".tgz"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "tgz",
			MemberEnd: "/lib/libonnxruntime." + semanticONNXRuntimeVer + ".dylib",
			LibName:   "libonnxruntime." + semanticONNXRuntimeVer + ".dylib",
		}, nil
	case "windows/amd64":
		name := "onnxruntime-win-x64-" + semanticONNXRuntimeVer + ".zip"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "zip",
			MemberEnd: "/lib/onnxruntime.dll",
			LibName:   "onnxruntime.dll",
		}, nil
	case "windows/arm64":
		name := "onnxruntime-win-arm64-" + semanticONNXRuntimeVer + ".zip"
		return runtimeAsset{
			URL:       base + name,
			Archive:   "zip",
			MemberEnd: "/lib/onnxruntime.dll",
			LibName:   "onnxruntime.dll",
		}, nil
	default:
		return runtimeAsset{}, fmt.Errorf("%w: unsupported platform %s/%s", ErrSemanticUnavailable, runtime.GOOS, runtime.GOARCH)
	}
}

func ensureReleaseFile(ctx context.Context, path, assetName string, allowDownload bool) error {
	if fileExistsAndNonEmpty(path) {
		return nil
	}
	if !allowDownload {
		return ErrSemanticUnavailable
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	want, err := downloadAssetChecksum(ctx, assetName)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := downloadURL(ctx, releaseAssetURL(assetName), tmp); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("download semantic asset %s: %w\nhint: connect once to GitHub Releases or use the fat C3 build for offline semantic search", assetName, err)
	}
	if err := verifyFileSHA256(tmp, want); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("verify semantic asset %s: %w", assetName, err)
	}
	return os.Rename(tmp, path)
}

func releaseAssetURL(assetName string) string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("C3_SEMANTIC_RELEASE_BASE_URL")), "/")
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(os.Getenv("C3X_RELEASE_BASE_URL")), "/")
	}
	if base == "" {
		version := strings.TrimSpace(os.Getenv("C3X_VERSION"))
		if version == "" {
			version = "dev"
		}
		base = semanticReleaseRepo + "/v" + version
	}
	return base + "/" + assetName
}

func downloadAssetChecksum(ctx context.Context, assetName string) (string, error) {
	body, err := downloadBytes(ctx, releaseAssetURL(assetName+".sha256"))
	if err != nil {
		return "", fmt.Errorf("download checksum for %s: %w", assetName, err)
	}
	for _, field := range strings.Fields(string(body)) {
		if len(field) != sha256.Size*2 {
			continue
		}
		if _, err := hex.DecodeString(field); err == nil {
			return strings.ToLower(field), nil
		}
	}
	return "", fmt.Errorf("checksum for %s: missing sha256 digest", assetName)
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
