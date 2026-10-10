# cliv の作業コンテキスト

## 目的と基本方針

- `cliv` は複数の管理元から、インストール済みの **ツール／パッケージ名とバージョン** を一覧表示する Go 製 CLI。
- コマンド名の一覧ではない。1パッケージが複数の実行コマンドを提供しても、それぞれを別行にしない。
- 読み取り専用。ツールのインストール・更新・削除・修復は行わない。
- Go のバージョンは `go.mod` を参照。端末判定に `golang.org/x/term` を使用している。それ以外は基本的に標準ライブラリを使用する。
- 現在の仕様は実装と `README.md` を参照する。`docs/MVP.md` は初期設計資料で、コマンド単位の表示など現在と異なる記述がある。

## コードの構成

- `cmd/cliv/main.go`：フラグ、collector の選択、フィルタリング、安定ソート、出力の呼び出し。
- `internal/collector/`：brew・mise・npm・uv・manual の情報収集と解析。
- `internal/model/entry.go`：共通の `CLIEntry`。
- `internal/output/`：ソース別の表と JSON 出力。
- `internal/config/config.go`：`~/.config/cliv/config.toml` の簡易ローダー。汎用 TOML パーサーではない。

## 表示とデータの仕様

- デフォルトは `brew` → `mise` → `npm` → `uv` → `manual` のセクションに分け、各セクションは `TOOL`・`VERSION` の2列。
- 色は `--color auto|always|never` で制御。auto は stdout が端末の場合のみ有効で、非空の `NO_COLOR` または `TERM=dumb` で無効。always は環境変数より優先。JSON は常に無着色。
- ANSI 色付けは列幅の計算後に行い、表の位置揃えを崩さない。
- 各セクション内はツール名順。存在しないソースのセクションは表示しない。
- 表のグルーピング／ソートで入力スライスを変更しない。JSON の並び順に影響させない。
- `--json` はグルーピングしない配列。フィールドは `tool`・`version`・`source`・`direct`。
- 旧 JSON フィールドの `command`・`package`・`path` は削除済み。実行ファイル用のフィールドを復活させない。
- `--source` は `brew`・`mise`・`npm`・`uv`・`manual` のいずれか。`--all` は現在 Homebrew の依存 formula を含めるための指定。

## ソースごとの収集方針

### Homebrew

- `brew info --json=v2 --installed` の formula 名とインストール済みバージョンを使用する。
- 通常は `installed_on_request` が true の formula のみ。`--all` では依存 formula も含める。
- `direct` はこの明示インストール判定を表す。`brew leaves` は同じ判定ではないので代用しない。
- 実行ファイルを持たない formula も対象。Cellar の `bin`・`sbin` の走査や `brew --prefix` は不要。

### mise

- `mise ls --json` のツール ID とバージョンを使用する。`npm:`・`pipx:`・`spm:` などの接頭辞を保持する。
- `installed: false` は除外するが、`active: false` は除外しない。古い版を含むインストール済みの複数バージョンを表示する。
- JSON の配列形式と、ツール名をキーにしたオブジェクト形式に対応している。
- インストール先や実行ファイルの存在を表示条件にしない。
- 今回の検出漏れの原因：旧実装は `install_path/bin` のみを探索していたが、mise の内蔵 aube による npm インストールは `install_path/node_modules/.bin` を使う。パッケージ名表示への変更により、この配置の違いに依存しなくなった。
- ホーム配下の固定パスや、特定の mise／Node バージョンを前提にしない。

### npm

- `npm list -g --depth=0 --json` のトップレベルパッケージ名とバージョンを使用する。
- `package.json` の `bin` や `npm root -g` に依存しない。実行ファイルを持たないパッケージも表示する。
- `direct` は true。

### uv

- `uv tool list --color never` のツール名とバージョンを使用する。実行コマンドを示す `- ...` の行は表示対象にしない。
- 現在はテキスト出力を解析している。フォーマット変更時はフィクスチャと解析処理を確認し、未知の形式を黙って誤表示しない。
- `uv tool install` の対象のみ。仮想環境の `uv pip` パッケージや一時的な `uvx` 実行は含めない。
- デフォルトで有効。設定の `[collectors] uv = false` で無効化できる。`direct` は true。

### manual

- 設定の `package` を表示名とし、未指定なら `command` にフォールバックする。
- 設定ファイルの `command`・`package`・`version_command` は引き続き有効。JSON フィールド変更と混同しない。
- バージョンは `version_command` から取得する。同名・同バージョンの登録は重複排除する。
- 設定の `source` にかかわらず収集結果は `manual` とする。他の管理元を装ったり、`--source` の選別を回避させない。
- `direct` は true。

## 変更後の確認

```sh
# 変更した Go ファイルに gofmt を実行する
go test ./...
go vet ./...
go build ./cmd/cliv

# ソースから実際の表示を確認する
go run ./cmd/cliv
go run ./cmd/cliv --source mise
go run ./cmd/cliv --json
```

- 収集率の計測は `scripts/check_collection.py --cliv <検証用バイナリ>`。期待一覧をネイティブ管理ツールから独立に取得し、`(source, tool, version)` で照合する。詳細は `docs/collection-metrics.md` を参照。
- 実行時間は `scripts/measure_runtime.py --cliv <検証用バイナリ>` で計測。全体・ソース別に初回と繰り返しの中央値／最小／最大を分ける。失敗した実行を高速化とみなさない。詳細は `docs/runtime-metrics.md` を参照。
- `python3 -m unittest discover -s scripts -p 'test_*.py'` で計測スクリプトも検証する。Go の期待一覧フィクスチャは `internal/collector/testdata/` にあり、パーサーから自動生成しない。
- 実環境の証跡は `measurements/` など gitignore 対象に保存し、ユーザー固有の情報をコミットしない。
- ユニットテストは実機の管理ツールやインストール構造に依存させず、JSON フィクスチャなどで解析・フィルタ・出力を確認する。
- 実行確認では、ソースから起動したものと PATH 上のインストール済み `cliv` を区別する。コード変更だけでは既存バイナリは更新されない。
- インストールする場合は `command -v cliv` と `go env GOPATH GOBIN` で配置先を確認してから `go install ./cmd/cliv` を実行し、インストール済みコマンドでも確認する。
- ローカルビルドの `cliv --version` は `dev`。リリース版は ldflags でタグのバージョンを埋め込む。
- 表示・JSON・オプションの仕様を変更したら、関連テストと `README.md` も更新する。
- ユーザーの実環境設定やインストール済みツールは、確認のために勝手に変更しない。
