---
target: ref-fat-thin-distribution
scope: block
base: ref-fat-thin-distribution#n754@v1:sha256:6979a60c4383212f3411f0db3f479f0d7da4197637a29111ad04fefa5955a124
---
CI assembles skill archives from one tree that includes `.gitattributes`, `.claude-plugin/`, and `skills/`: the platform-neutral `c3-skill-v{VERSION}.zip` removes all `skills/c3/bin/c3x-*` binaries, each full-fat `c3-skill-{platform}-v{VERSION}.zip` adds exactly the matching semantic-capable fat binary, and each Linux portable `c3-skill-linux-{arch}-portable-v{VERSION}.zip` adds exactly the matching pure-Go portable binary. The npm manager honors C3X_VERSION when externally set for development/tests, otherwise resolves project or latest runtime versions and downloads verified runtime assets from the GitHub Release.
