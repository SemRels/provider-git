# provider-git

Local Git subprocess plugin for SemRel.

The plugin runs after SemRel creates a release tag, then pushes the requested tag and branch to a configured remote with `git push`.

## Repository Layout

~~~text
cmd/plugin/              Plugin entry point
internal/plugin/         Git subprocess client and configuration
.github/workflows/       CI, release, and security automation
~~~

## Development

~~~bash
go build ./cmd/plugin
go test ./...
~~~

## Runtime Environment

- `SEMREL_TAG_NAME` - tag to push
- `SEMREL_BRANCH` - branch to update with `git push <remote> HEAD:<branch>`
- `SEMREL_PLUGIN_REMOTE` - remote name, defaults to `origin`
- `SEMREL_DRY_RUN` - when set to `true`, log the intended pushes without executing git
