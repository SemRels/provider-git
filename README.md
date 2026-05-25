# provider-git

Git provider plugin for Semantic Release.

Provides local Git repository integration for Semantic Release workflows.

## Documentation

- Docs (coming soon): <https://github.com/SemRels/semrel/tree/main/docs/plugins/provider-git>
- Template source: <https://github.com/SemRels/plugin-template>

## Repository Layout

`	ext
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
`

## Development

`ash
go build ./cmd/plugin
go test ./...
`

## Configuration Example

`yaml
plugins:
  - name: provider-git
    type: provider
    config:
      repository_path: .
      remote_name: origin
      base_branch: main
`

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.