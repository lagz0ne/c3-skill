---
target: ref-fat-thin-distribution
scope: block
base: ref-fat-thin-distribution#n754@v1:sha256:48ca49a642b2e43c8e60b60f8c82448cf054333e4893d5d700a47aa31fad727f
---
CI assembles skill archives from one tree that includes `.gitattributes`, `.claude-plugin/`, `.codex-plugin/`, and `skills/`: the platform-neutral `c3-skill-v{VERSION}.zip` removes all `skills/c3/bin/c3x-*` binaries, each full-fat `c3-skill-{platform}-v{VERSION}.zip` adds exactly the matching semantic-capable fat binary, and each Linux portable `c3-skill-linux-{arch}-portable-v{VERSION}.zip` adds exactly the matching pure-Go portable binary. The npm manager pins or resolves C3X_VERSION and downloads verified runtime assets from the GitHub Release.
