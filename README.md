# cliv

`cliv` lists tools and package versions installed through multiple package managers in one read-only view.

```sh
cliv
```

Default output:

```text
brew
TOOL      VERSION
ripgrep   14.1.1

mise
TOOL   VERSION
node      24.8.0

npm
TOOL            VERSION
@openai/codex   0.42.0
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
cliv --source brew   # filter by source (brew, mise, npm, uv, manual)
cliv --all           # include Homebrew dependency formulae
cliv --json          # JSON output (ungrouped array)
cliv --color always  # force colored headings (auto / always / never)
cliv --version       # cliv version
cliv --help          # help
```

Headings are colored automatically when stdout is a terminal. Pipes and redirected
output remain plain. A non-empty `NO_COLOR` or `TERM=dumb` disables automatic
color; `--color always` overrides these settings. JSON is never colored.

## Supported sources

- Homebrew (`brew`): installed formula names and versions; direct formulae by default, with `--all` also including dependencies. Formulae without executables are included.
- mise (`mise`): installed tool IDs and versions from `mise ls --json`, including backend prefixes such as `npm:` and all installed versions. Executable discovery is not required.
- npm (`npm`): global top-level package names and versions, including packages without executables.
- uv (`uv`): tool names and versions installed with `uv tool install`, from `uv tool list`. Virtual-environment packages (`uv pip`) and temporary `uvx` executions are not included.
- manual: optional entries from `~/.config/cliv/config.toml`; displays `package`, falling back to `command` when unspecified.

`cliv --json` outputs an array with `tool`, `version`, `source`, and `direct` fields. `tool` is the displayed tool/package name. `direct` distinguishes explicitly installed Homebrew formulae from dependencies; other sources report `true`. The former `command`, `package`, and `path` fields have been removed.

```json
[
  {
    "tool": "npm:@openai/codex",
    "version": "0.162.0",
    "source": "mise",
    "direct": true
  }
]
```

## Config

```toml
[collectors]
brew = true
mise = true
npm = true
uv = true

[[tools]]
command = "claude"
package = "claude"
source = "manual"
version_command = "claude --version"
```

`cliv` is read-only and does not install, update, remove, or repair tools.

## Failure handling

If collection fails, cliv keeps available results but exits with status `1` and
reports the failed sources and an incomplete-inventory warning on stderr.
JSON remains an array; check the exit code before treating it as a complete inventory.
Missing package managers and empty inventories are not errors. Invalid arguments
exit with status `2`. See [failure behavior](docs/failure-behavior.md) for details
and safe shell usage.

## Collection metrics

Build the binary to verify and compare its JSON inventory against native package-manager lists:

```sh
go build -o /tmp/cliv-check ./cmd/cliv
python3 scripts/check_collection.py --cliv /tmp/cliv-check
```

The checker reports per-source coverage, missing/extra entries, and duplicates.
See [collection metrics](docs/collection-metrics.md) for evidence capture, exit codes,
and fixture tests.

## Runtime metrics

Measure the compiled binary's total and per-source wall-clock runtime:

```sh
go build -o /tmp/cliv-check ./cmd/cliv
python3 scripts/measure_runtime.py --cliv /tmp/cliv-check
```

The first run is reported separately from the median/min/max of five subsequent runs.
Use `--targets`, `--runs`, and `--output-dir` to customize measurement and save evidence.
See [runtime metrics](docs/runtime-metrics.md) for methodology and comparison caveats.

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
