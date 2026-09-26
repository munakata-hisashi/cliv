# cliv

`cliv` lists CLI commands installed through multiple package managers in one read-only view.

```sh
cliv
```

Default output:

```text
COMMAND   VERSION   SOURCE
rg        14.1.1    brew
node      24.8.0    mise
codex     0.42.0    npm
```

## Install / run

```sh
go run ./cmd/cliv
# or
go install ./cmd/cliv
```

Prebuilt binaries for Linux, macOS and Windows (amd64/arm64) are available from
[GitHub Releases](https://github.com/munakata-hisashi/cliv/releases). Download the
archive for your OS/architecture, verify it against `checksums.txt`, and place
`cliv` (or `cliv.exe`) on your PATH. Release binaries report their tag version
with `cliv --version`; local source builds report `dev`.

## Options

```sh
cliv --source brew   # filter by source
cliv --all           # include dependency packages when supported
cliv --json          # JSON output
cliv --version       # cliv version
cliv --help          # help
```

## Supported sources

- Homebrew (`brew`): direct formulae by default; `--all` includes dependencies.
- mise (`mise`): installed tools from `mise ls --json`.
- npm (`npm`): global top-level packages and their `package.json` `bin` commands.
- manual: optional entries from `~/.config/cliv/config.toml`.

## Config

```toml
[collectors]
brew = true
mise = true
npm = true

[[tools]]
command = "claude"
package = "claude"
source = "manual"
version_command = "claude --version"
```

`cliv` is read-only and does not install, update, remove, or repair tools.

## CI and releases

GitHub Actions runs tests, vet and build checks on pushes to `main` and pull
requests (Linux and macOS). To publish a release, create and push a version tag
on a tested commit:

```sh
git tag v0.1.0
git push origin v0.1.0
```

The release workflow checks the tag format (`vMAJOR.MINOR.PATCH`), runs tests
and vet again, builds cross-platform archives with the tag version embedded,
and publishes them with `checksums.txt` to GitHub Releases. It requires GitHub
Actions to have permission to create releases via `GITHUB_TOKEN` (repository
Settings → Actions → General → Workflow permissions; the workflow requests
`contents: write` for publishing). User-specific configuration in
`~/.config/cliv/config.toml` is not part of a release.
