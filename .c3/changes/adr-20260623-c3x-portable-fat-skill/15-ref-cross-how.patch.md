---
target: ref-cross-compiled-binary
scope: block
base: ref-cross-compiled-binary#n736@v1:sha256:4cc30a30543cf1a4667937fb8883febff3cdcafe8bf92ec5cc9d529b4e2a48be
---
A `v*` version tag triggers the distribute workflow, which runs the release build variant: thin plus full-fat binaries for each platform in the matrix, and a `CGO_ENABLED=0` portable binary for each Linux arch. Release assembly publishes the standard binaries for the npm manager and packages full-fat and Linux portable binaries into their matching skill ZIPs.
