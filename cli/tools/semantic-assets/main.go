package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func main() {
	var embedDir string
	var releaseDir string
	var targetOS string
	var targetArch string
	flag.StringVar(&embedDir, "embed-dir", "", "write model.onnx, vocab.txt, and target runtime for go:embed")
	flag.StringVar(&releaseDir, "release-dir", "", "write release-named semantic assets and checksums")
	flag.StringVar(&targetOS, "os", "", "target GOOS for embedded runtime")
	flag.StringVar(&targetArch, "arch", "", "target GOARCH for embedded runtime")
	flag.Parse()

	if (embedDir == "") == (releaseDir == "") {
		fail("pass exactly one of --embed-dir or --release-dir")
	}

	ctx := context.Background()
	var err error
	if embedDir != "" {
		if targetOS == "" || targetArch == "" {
			fail("--embed-dir requires --os and --arch")
		}
		err = store.PrepareEmbeddedSemanticAssets(ctx, embedDir, targetOS, targetArch)
	} else {
		err = store.PrepareReleaseSemanticModelAssets(ctx, releaseDir)
	}
	if err != nil {
		fail(err.Error())
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
