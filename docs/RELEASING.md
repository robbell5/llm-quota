# Releasing

`llm-quota` releases are automated with [GoReleaser](https://goreleaser.com).
Pushing a semver tag builds the binaries, publishes a GitHub Release, and updates
the Homebrew formula. There is nothing to build or upload by hand.

## Cut a release

From an up-to-date `main`, tag and push (replace `v0.1.1` with the version you
are releasing):

```sh
git tag -a v0.1.1 -m "v0.1.1"
git push origin v0.1.1
```

The tag must match `v*` (for example `v1.2.3`). GoReleaser strips the leading
`v`, so the release and formula version become `1.2.3`.

## What happens automatically

The `.github/workflows/release.yml` workflow runs on the tag push and:

1. runs `go test -race ./...` as a gate — a test failure aborts the release;
2. cross-compiles `darwin` and `linux` for `amd64` and `arm64`, stamping the
   version into the binary (`llm-quota --version`);
3. publishes a GitHub Release with the four archives and `checksums.txt`;
4. updates `Formula/llm-quota.rb` in the
   [`robbell5/homebrew-tap`](https://github.com/robbell5/homebrew-tap) repository.

Users then upgrade with:

```sh
brew update && brew upgrade robbell5/tap/llm-quota
```

## Verify

```sh
gh release view v0.1.1 --repo robbell5/llm-quota
brew update && brew upgrade robbell5/tap/llm-quota && llm-quota --version
```

The release should list four `*.tar.gz` archives plus `checksums.txt`, and
`llm-quota --version` should report the new version.

## Prerequisites

A repository secret `HOMEBREW_TAP_GITHUB_TOKEN` must exist on
`robbell5/llm-quota`: a fine-grained personal access token with **Contents:
Read and write** on `robbell5/homebrew-tap`. It is already configured and only
needs attention if it expires.

## Troubleshooting

If the GoReleaser step fails with `403` or `resource not accessible by
integration`, the `HOMEBREW_TAP_GITHUB_TOKEN` token is missing, under-scoped, or
expired. Regenerate it, update the secret, then re-tag:

```sh
git push --delete origin v0.1.1 && git tag -d v0.1.1
# fix the secret, then tag and push again
```

Every run logs a `DEPRECATED: brews` warning. That is expected: the project
intentionally keeps the formula (`brews`) section so Homebrew works on Linux as
well as macOS. The release still succeeds.
