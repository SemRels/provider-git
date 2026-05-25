# provider-git

Local Git provider plugin for SemRel.

Provides commit, branch, and tag access from a local Git repository during SemRel execution.

## Documentation

- SemRel docs (planned): <https://github.com/SemRels/semrel/tree/main/docs/plugins/provider-git>
- Plugin template: <https://github.com/SemRels/plugin-template>
- Registry: <https://registry.semrel.io>

## Repository Layout

~~~text
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
~~~

## Development

~~~bash
go build ./cmd/plugin
go test ./...
~~~

## Configuration Example

~~~yaml
plugins:
  - name: provider-git
    type: provider
    config:
      repository_path: .
      release_branch: main
      annotated_tags: true
~~~

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.
