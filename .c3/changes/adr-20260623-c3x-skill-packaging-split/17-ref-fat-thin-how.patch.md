---
target: ref-fat-thin-distribution
scope: block
base: ref-fat-thin-distribution#n754@v1:sha256:cc9044a4d3376240919c15a1d007e55249293555a9b81512535cb2c4e27403f0
---
CI assembles skill archives from one tree that includes `.gitattributes`, `.claude-plugin/`, `.codex-plugin/`, and `skills/`: the platform-neutral `c3-skill-v{VERSION}.zip` removes all `skills/c3/bin/c3x-*` binaries, while each `c3-skill-{platform}-v{VERSION}.zip` adds exactly the matching fat binary. The npm manager pins or resolves C3X_VERSION and downloads verified runtime assets from the GitHub Release.
