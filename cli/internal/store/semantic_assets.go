package store

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
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
)

const semanticModelURL = "https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/" + semanticHFRevision + "/onnx/model.onnx"
const semanticVocabURL = "https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/" + semanticHFRevision + "/vocab.txt"

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
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("semantic cache dir: %w", err)
	}
	return filepath.Join(base, "c3", "semantic"), nil
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

	if err := ensureFile(ctx, assets.ModelPath, semanticModelURL, allowDownload); err != nil {
		return semanticAssets{}, err
	}
	if err := ensureFile(ctx, assets.VocabPath, semanticVocabURL, allowDownload); err != nil {
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

func ensureFile(ctx context.Context, path, url string, allowDownload bool) error {
	if fileExistsAndNonEmpty(path) {
		return nil
	}
	if !allowDownload {
		return ErrSemanticUnavailable
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := downloadURL(ctx, url, tmp); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: semanticHTTPTimeout}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("download %s: status %s", url, res.Status)
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, res.Body); err != nil {
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
