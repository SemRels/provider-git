# provider-git

[![Latest Release](https://img.shields.io/github/v/release/SemRels/provider-git?label=version&color=blue)](https://github.com/SemRels/provider-git/releases/latest)
[![CI](https://github.com/SemRels/provider-git/actions/workflows/ci.yml/badge.svg)](https://github.com/SemRels/provider-git/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/SemRels/provider-git)](LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/SemRels/provider-git/badge)](https://scorecard.dev/viewer/?uri=github.com/SemRels/provider-git)
[![REUSE status](https://api.reuse.software/badge/github.com/SemRels/provider-git)](https://api.reuse.software/info/github.com/SemRels/provider-git)

Pushes the semrel tag and optional branch updates to a Git remote.

This plugin is distributed as the standalone Go binary `semrel-plugin-provider-git`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

### Binary

```bash
go install github.com/SemRels/provider-git/cmd/plugin@latest
```

### Docker

Pre-built, multi-platform images (linux/amd64, linux/arm64) are published to the GitHub Container Registry on every release:

```bash
docker pull ghcr.io/semrels/provider-git:latest
```

Images are signed with [cosign](https://github.com/sigstore/cosign) and include a full SBOM attestation. Verify the signature:

```bash
cosign verify ghcr.io/semrels/provider-git:latest \
  --certificate-identity-regexp 'https://github.com/SemRels/provider-git/.github/workflows/container.yml.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```


## Configuration

```yaml
plugins:
  - uses: "@semrel/provider-git"
    phase: release
    args:
      remote: "origin"
      signing_key: "C:\\keys\\release-signing.asc"
      push_branch: "true"
```

## Options

Every option can be set as an entry under `args` in `.semrel.yaml`. semrel passes each entry to the plugin as the environment variable shown in the table, so you can also set that variable directly (for example in your CI environment) instead of using `args`.

| Argument | Environment variable | Required | Description | Default |
| --- | --- | --- | --- | --- |
| `remote` | `SEMREL_PLUGIN_REMOTE` | Optional | Git remote name to push to. | origin |
| `signing_key` | `SEMREL_PLUGIN_SIGNING_KEY` | Optional | Path to a GPG key used when signing the tag. | None |
| `push_branch` | `SEMREL_PLUGIN_PUSH_BRANCH` | Optional | Push the current branch in addition to the tag. | false |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_BRANCH` | Git branch associated with the current release run. |
| `SEMREL_DRY_RUN` | Whether semrel is running in dry-run mode. |

## Example behavior

The plugin pushes the release tag to the configured remote and can optionally push the branch as part of the same release transaction.

## License

Apache-2.0
