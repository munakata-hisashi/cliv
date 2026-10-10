# 実行時間の計測

## 指標

ビルド済み cliv のプロセス起動から終了までの経過時間（wall-clock time）を、Python の `time.perf_counter()` で計測する。

- `total`：すべての有効な collector と表／JSON 出力を含む全体時間。
- ソース別：`--source brew|mise|npm|uv|manual` の実行時間。collector 本体だけでなく、cliv の起動・設定読み込み・出力も含む。
- `FIRST`：各対象の初回実行時間。
- `MEDIAN`・`MIN`・`MAX`：初回とは別に繰り返した実行の中央値・最小値・最大値。

期待一覧を取得する収集率チェックや、ビルド時間は含めない。stdout はパイプで取得し、色は無効にして、端末の描画速度の影響を避ける。出力の生成・パイプへの書き込み・取得とプロセス起動のオーバーヘッドは含む。CPU 時間やメモリ使用量の計測ではない。

## 実行方法

Python 3.9 以降を使用。追加 Python パッケージは不要。

```sh
go build -o /tmp/cliv-check ./cmd/cliv

# 全体と全ソース。各対象を初回1回＋繰り返し5回実行
python3 scripts/measure_runtime.py --cliv /tmp/cliv-check

# 対象と繰り返し回数を限定
python3 scripts/measure_runtime.py --cliv /tmp/cliv-check \
  --targets total brew mise npm uv --runs 10

# 証跡を保存（出力先は新規ディレクトリ）
python3 scripts/measure_runtime.py --cliv /tmp/cliv-check \
  --output-dir measurements/runtime-001

# JSON／Homebrew の依存を含む条件も計測可能
python3 scripts/measure_runtime.py --cliv /tmp/cliv-check --json --all
```

`--runs` は初回を除く回数（1以上）。`--timeout` は各コマンドのタイムアウト秒数（デフォルト60）。Windows ではビルド先に適切な `.exe` パスを指定する。

## 集計と失敗の扱い

- 非ゼロ終了、タイムアウト、起動失敗、非空の cliv stderr は `ERROR`。終了コード0でも collector の失敗があり得るため stderr を確認する。
- 1回でも失敗した対象は残りの繰り返しを中止し、中央値などを表示しない。不完全な実行を「高速化」として扱わない。他の対象の計測は続行する。
- 管理ツールが未導入ならそのソースは `SKIP`。`total` と `manual` は実行する。
- 設定で無効なソースやツール未登録の場合は空一覧になり得る。空一覧の対象には注意書きを出す。設定と収集率を確認してから時間を比較する。
- 終了コードは正常計測が `0`、計測失敗または全対象 SKIP が `2`。速度悪化の閾値判定はまだ行わない。

## 証跡

`--output-dir` に `report.json` を保存する。

- 計測日時、OS・Python、作業ディレクトリ
- cliv の実行パス・SHA-256・バージョン
- cliv 設定ファイルの SHA-256（ファイルなし／読み取り不可なら null。設定内容は保存しない）
- 管理ツールのバージョン
- 条件（対象、回数、タイムアウト、出力形式、`--all`）
- 各実行の時間、コマンド、終了コード、stdout、stderr、集計結果

バージョン情報の取得は計測後に行い、計測前に管理ツールを起動しない。バージョン取得結果は参考情報として保存し、その失敗だけでは計測全体を ERROR にしない。

実環境の出力はユーザー固有の情報を含み得るため、そのままコミットしない。`measurements/` は gitignore 対象。

## 比較時の注意

- `FIRST` は「この計測での初回」であり、コールドキャッシュを保証しない。OS や管理ツールのキャッシュは消去しない。
- 対象は指定順に実行する。デフォルトは total → brew → mise → npm → uv → manual。total の実行で後続ソースのキャッシュが温まる可能性がある。比較時は対象と順序をそろえる。
- 同じマシン・設定・インストール状態・出力形式・対象・回数で比較する。計測中のインストールや更新、重いバックグラウンド処理を避ける。
- 環境変数や外部キャッシュの状態すべては保存しない。必要な条件は別途記録する。
- ソース別時間の合計は全体時間と一致するとは限らない。ソース別ではプロセス起動などがそれぞれ発生する。
- manual の時間は設定した `version_command` の数と内容に依存する。比較のために設定を書き換えない。
- まず収集率が一致していることを確認する。時間だけでは正しく収集できていることを保証できない。

## テスト

```sh
python3 -m unittest discover -s scripts -p 'test_*.py'
```

初回の除外、中央値などの計算、collector 失敗時の集計無効化、タイムアウト、起動失敗を固定データ／モックで検証する。既存 CI の Python テスト対象に含まれる。実環境の時間は CI runner の負荷に左右されるため、現時点では CI の合否判定には使わない。
