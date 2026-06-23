---
target: c3-203
scope: block
base: c3-203#n616@v1:sha256:0ab994bab6c71d3f692691e55c8cec5bdb1f3e89dd3ff6e0e358c326f58a31cd
---
| ref-cross-compiled-binary | ref | The full per-platform asset name, Linux portable asset name, and supported OS/arch matrix the wrapper resolves to before exec | Wrapper selects only assets the build matrix produces | c3x.sh maps uname to amd64/arm64, gates on linux amd64/arm64 and darwin arm64, tries c3x-${VERSION}-${OS}-${ARCH}, and on Linux also tries c3x-${VERSION}-${OS}-${ARCH}-portable. |
