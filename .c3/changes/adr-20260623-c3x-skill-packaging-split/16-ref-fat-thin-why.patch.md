---
target: ref-fat-thin-distribution
scope: block
base: ref-fat-thin-distribution#n752@v1:sha256:979f39ec55737c64e890a8bef1e52407c3f054dbed1dc75fee9a76aa79b5624f
---
Sandboxed and offline skill installs need a self-contained binary and embedded semantic model; plugin and skills CLI installers need a platform-neutral package that does not carry binaries; npm users want a small package that pulls only the runtime assets they need. The split keeps one underlying Go binary while matching each install environment's constraints.
