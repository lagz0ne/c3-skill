---
target: c3-203
scope: block
base: c3-203#n612@v1:sha256:d5016a0107e4fcf30ec4c36598b66c0bf4d393c16ce94526e40d2f6c16c61b01
---
Carry bin/c3x.sh and bin/VERSION: read VERSION, lowercase the OS and normalize the arch (x86_64 to amd64, aarch64/arm64 to arm64), reject any platform outside linux amd64/arm64 and darwin arm64 with a hint, compute the `c3x-${VERSION}-${OS}-${ARCH}` asset name, exec the full bundled binary when present, on Linux exec `c3x-${VERSION}-${OS}-${ARCH}-portable` when present, otherwise build the full binary from cli/ with `go build` (embedmodel tag, version ldflag) when Go and go.mod exist, otherwise answer passive root help/version locally or delegate real commands to `npm exec --yes --package @c3x/cli@${VERSION} -- c3x` without forcing C3X_VERSION so no-binary skill installs use the pinned manager package while preserving the manager's project/latest runtime resolution. Non-goals: implementing any C3 command (the binary), making project metadata executable authority, or shaping `.c3/` content.
