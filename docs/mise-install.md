# mise 経由のインストール

## 利用者向け手順

```sh
mise use -g github:munakata-hisashi/cliv@0.2.0
cliv --version
# cliv 0.2.0
```

mise のシェル有効化が済んでいなければ、次でも起動できる。

```sh
mise exec github:munakata-hisashi/cliv@0.2.0 -- cliv --version
```

`mise use -g` は利用者のグローバル設定にツールを追加する。追加せず一度実行したい場合は `mise exec` を使う（未導入なら自動インストールされる）。

GitHub Releases バックエンドを使うため、専用プラグインや Go は不要。レジストリへの短縮名登録はまだないので、`mise use -g cliv` ではなく完全な `github:` ID を指定する。

## 検証結果

- 環境：macOS arm64、mise 2026.10.7。
- 対象：GitHub Release v0.2.0。
- `github:munakata-hisashi/cliv@0.2.0` だけで `cliv_v0.2.0_darwin_arm64.tar.gz` が選択された。
- アーカイブ内のトップレベルディレクトリは適切に処理され、インストール先の `cliv` を検出できた。
- `mise exec ... -- cliv --version` は `cliv 0.2.0` を返した。
- 分離した設定で `mise use -g` も成功し、`mise ls --json` で `installed: true`・`active: true` を確認した。
- `asset_pattern`・`bin_path`・`strip_components` の追加指定は不要だった。

Linux・Intel Mac・Windows は公開バイナリが存在するが、mise 経由のインストールはまだ実機検証していない。リリース資産の名前やアーカイブ構造を変更した場合は再検証する。

## 開発時の安全な再検証

通常のグローバル設定、キャッシュ、インストール済みツールを変更しないよう、専用の一時環境を使う。以下は POSIX シェル向け。

```sh
# サブシェルなので HOME などの変更を現在のシェルに残さない
(
  set -eu
  mise_bin="$(command -v mise)"
  sandbox="$(mktemp -d)"
  echo "Sandbox: $sandbox"
  mkdir -p "$sandbox/home" "$sandbox/work"
  cd "$sandbox/work"
  export HOME="$sandbox/home"
  export MISE_DATA_DIR="$sandbox/data"
  export MISE_CACHE_DIR="$sandbox/cache"
  export MISE_CONFIG_DIR="$sandbox/config"
  export MISE_STATE_DIR="$sandbox/state"
  export XDG_CONFIG_HOME="$sandbox/config"
  export XDG_DATA_HOME="$sandbox/data"
  export XDG_CACHE_HOME="$sandbox/cache"
  export XDG_STATE_HOME="$sandbox/state"
  export MISE_YES=1

  "$mise_bin" install github:munakata-hisashi/cliv@0.2.0
  "$mise_bin" exec github:munakata-hisashi/cliv@0.2.0 -- cliv --version
  "$mise_bin" use -g github:munakata-hisashi/cliv@0.2.0
  "$mise_bin" exec -- cliv --version
  "$mise_bin" ls --json
)
```

独自の `MISE_CONFIG_FILE` など、設定の場所を上書きする環境変数を使っている場合は、それらも一時環境に向けるか解除してから検証する。既存の mise 環境を更新したり、実環境で `mise use -g` を試さない。

公式資料：https://mise.jdx.dev/dev-tools/backends/github.html
