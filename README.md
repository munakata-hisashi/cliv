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
