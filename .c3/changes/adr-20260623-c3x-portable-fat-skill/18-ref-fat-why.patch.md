---
target: ref-fat-thin-distribution
scope: block
base: ref-fat-thin-distribution#n752@v1:sha256:996191edd4b1736940a5cd74fdf18997057dc5cc59dfd55f3615551210193c27
---
Sandboxed and offline skill installs need a self-contained binary, but not every sandbox can run the native semantic runtime path. Full-fat artifacts preserve embedded semantic search for supported environments, portable Linux artifacts favor distro compatibility with keyword/graph fallback, plugin and skills CLI installers need a platform-neutral package that does not carry binaries, and npm users want a small package that pulls only the runtime assets they need.
