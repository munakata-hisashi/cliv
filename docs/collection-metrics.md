# 収集率の計測

## 目的と指標

管理ツールのネイティブ一覧と `cliv --json` を、`(source, tool, version)` の組で照合する。

- 収集率：期待一覧と実際の一覧の共通件数 / 期待件数。期待0件は `N/A`。
- 欠落：期待一覧にあるが cliv にない組。
- 余分：cliv にあるが期待一覧にない組。
- 重複：cliv に同じ組が2回以上出る場合。表の件数は重複した組の数。
- 目標：収集率100%、欠落・余分・重複0件。空一覧同士も一致として扱う。

`direct` は収集率の比較キーに含めないが、固定フィクスチャのテストでは検証する。

## 実環境での計測

Python 3.9 以降と Go が必要。Python の追加パッケージは不要。

```sh
go build -o /tmp/cliv-check ./cmd/cliv
python3 scripts/check_collection.py --cliv /tmp/cliv-check

# ソース限定／Homebrew の依存を含める
python3 scripts/check_collection.py --cliv /tmp/cliv-check --sources mise uv
python3 scripts/check_collection.py --cliv /tmp/cliv-check --all

# 証跡の保存（出力先は新しいディレクトリを指定）
python3 scripts/check_collection.py --cliv /tmp/cliv-check \
  --output-dir measurements/run-001
```

`--cliv` を必須とし、PATH 上の古いバイナリを誤って計測しないようにする。
Windows ではビルド先を適宜 `.exe` のパスに変更する。

デフォルトの対象は brew・mise・npm・uv。未導入の管理ツールは `SKIP` で計測から除外する。すべて SKIP の場合は計測不能。cliv 側の collector が設定で無効なら、ネイティブ一覧との欠落として現れる。計測対象は `~/.config/cliv/config.toml` で有効にするか、`--sources` で限定する。スクリプトは設定を変更しない。

期待一覧の基準：

| ソース | 取得コマンドと対象 |
|---|---|
| brew | `brew info --json=v2 --installed`。通常は `installed_on_request: true` の formula。`--all` で依存も含む |
| mise | `mise ls --json`。`installed: false` を除外し、非アクティブな版も含む |
| npm | `npm list -g --depth=0 --json`。トップレベルのパッケージ |
| uv | `uv tool list --color never`。ツールの行のみ。コマンドを示す `- ...` の行は除外 |

manual は任意のシェルコマンドでバージョンを取得するため、実環境の収集率計測から除外し、通常のユニットテストで確認する。

## 結果と終了コード

ソースごとに EXPECTED・MATCHED・MISSING・EXTRA・DUPLICATES・COVERAGE と PASS/DIFF/SKIP/ERROR を表示する。差分の具体的な組も表示する。

- `0`：計測した全ソースが一致（未導入ソースの SKIP は許容）。
- `1`：欠落・余分・重複のいずれかが存在する。
- `2`：コマンド失敗、出力解析失敗、全ソース SKIP などで計測不能。差分と計測不能が混在すれば2を優先する。

cliv の収集失敗は終了コード1で判別する。旧バイナリでの計測も考慮し、cliv の stderr が非空なら計測不能とする。ネイティブコマンドの非ゼロ終了も計測不能とし、不完全な一覧を期待値にしない。タイムアウトはコマンドごとに60秒で、`--timeout` で変更できる。

`--output-dir` では `report.json` に以下を保存する：

- 計測時刻、cliv の実行パスとバージョン、管理ツールのバージョン
- ネイティブコマンドと cliv のコマンド、終了コード、stdout、stderr
- 正規化した期待一覧、比較結果、計測の終了コード

ユーザー固有の情報を含み得るため、証跡はそのままコミットしない。`measurements/` は gitignore 対象。

## 再現性と限界

- 計測中にツールのインストール・削除・更新をしない。ネイティブ一覧と cliv は順番に取得するため、環境が変化すると偽の差分が生じる。
- 期待一覧の解析は Python で独立実装し、cliv の collector を呼ばない。ただし同じ仕様の誤解までは防げない。
- この計測は管理ツールが報告する一覧との一致を測るもので、管理ツール自身の不具合やディスク上の完全性を保証しない。
- uv の期待一覧もテキスト出力の解析なので、将来の出力形式変更時には見直す。

## 固定データの検証と CI

```sh
go test ./...
python3 -m unittest discover -s scripts -p 'test_*.py'
```

`internal/collector/testdata/` にネイティブ出力と、人が確認した期待一覧を保存する。`fixture_test.go` が全件一致を検証する。期待一覧はパーサーの出力から自動生成しない。

現在のフィクスチャは、brew の直接／依存 formula、mise のバックエンド ID と複数版／未インストール版、npm のパッケージ、uv の複数コマンドを持つツールを対象とする。空一覧・不正形式などは既存のユニットテストと Python のテストで補う。

CI は Go と Python の固定データテストを実行する。実環境照合は CI runner のインストール状態に依存するため、手元で必要時に実行する。
