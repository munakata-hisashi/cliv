# cliv MVP 設計書

## 1. 概要

`cliv` は、ローカルマシンにインストールされているCLIツールを、インストール方法に依存せず一覧表示するための読み取り専用CLIツールである。

主な目的は、Homebrew、mise、npmなど複数の経路でインストールされたCLIを、1つのコマンドでまとめて確認できるようにすること。

コンセプト:

> **mise list for your entire CLI environment**

MVPでは、一覧表示以外の機能を極力持たせない。

---

## 2. 想定利用

基本操作は以下のみ。

```bash
cliv
```

出力例:

```text
COMMAND   VERSION    SOURCE
rg        14.1.1     brew
jq        1.8.1      brew
node      24.8.0     mise
python    3.13.7     mise
codex     0.42.0     npm
claude    1.0.113    manual
```

ユーザーは、

> このマシンに、どのCLIが、どのバージョンで、どこ経由で入っているか

を確認できる。

---

## 3. MVPの責務

MVPでは以下のみを行う。

- CLIツールの検出
- CLI名の取得
- バージョンの取得
- インストール元の取得
- 一覧表示
- JSON出力
- Sourceによるフィルタ

---

## 4. 非目標

MVPでは以下を実装しない。

- CLIのインストール
- CLIのアンインストール
- CLIのアップデート
- CLIの自動修復
- Package Managerのラッパー
- 重複インストール診断
- `doctor`
- Snapshot
- History
- Diff
- Shell integration
- Update check
- Environment fingerprint
- Project lockfile
- PATH競合の詳細解析
- symlink chain解析
- 高度なOwnership Resolution

`cliv` は、CLI環境を変更しない。

---

## 5. コマンド仕様

### 5.1 基本

```bash
cliv
```

インストールされているCLI一覧を表示する。

`list` サブコマンドはMVPでは設けない。

```bash
cliv list
```

ではなく、

```bash
cliv
```

を基本操作とする。

---

### 5.2 Source filter

```bash
cliv --source brew
```

例:

```text
COMMAND   VERSION   SOURCE
rg        14.1.1    brew
jq        1.8.1     brew
git       2.51.0    brew
```

複数指定についてはMVPでは不要。

---

### 5.3 All

```bash
cliv --all
```

通常一覧から除外する依存packageなどを含めて表示する。

主にHomebrew用。

---

### 5.4 JSON

```bash
cliv --json
```

例:

```json
[
  {
    "command": "rg",
    "package": "ripgrep",
    "version": "14.1.1",
    "source": "brew"
  },
  {
    "command": "node",
    "package": "node",
    "version": "24.8.0",
    "source": "mise"
  }
]
```

---

### 5.5 Version

```bash
cliv --version
```

`cliv` 自身のバージョンを表示する。

---

### 5.6 Help

```bash
cliv --help
```

利用可能なオプションを表示する。

---

## 6. 出力形式

デフォルトはTable形式。

```text
COMMAND   VERSION    SOURCE
rg        14.1.1     brew
jq        1.8.1      brew
node      24.8.0     mise
python    3.13.7     mise
codex     0.42.0     npm
```

必要に応じてPackage名を表示できるよう、内部データには保持する。

初期表示では、ユーザーが実際に入力する `COMMAND` を最も重要な列とする。

MVPのデフォルト列:

```text
COMMAND
VERSION
SOURCE
```

---

## 7. PackageとCommand

Package名とCommand名は一致するとは限らない。

例:

```text
PACKAGE             COMMAND
ripgrep             rg
@openai/codex       codex
typescript          tsc
typescript          tsserver
```

そのため内部では両方保持する。

ただしMVPの通常表示ではCommandを主軸にする。

---

## 8. データモデル

MVPではシンプルな1レコード形式とする。

```text
CLIEntry
- command
- package
- version
- source
- path?
- direct
```

例:

```json
{
  "command": "codex",
  "package": "@openai/codex",
  "version": "0.42.0",
  "source": "npm",
  "path": "/opt/homebrew/bin/codex",
  "direct": true
}
```

`path` は取得できる場合のみ保持する。

MVPでは `PackageInstallation` と `CommandInstallation` を別モデルに分離しなくてもよい。

将来的に必要になった段階で正規化する。

---

## 9. Source

MVPで正式対応するSource:

```text
brew
mise
npm
```

余裕があれば追加:

```text
cargo
uv
```

手動登録についてはMVPでは設定ファイルによる簡易対応のみ検討する。

---

## 10. Collector

各SourceについてCollectorを持つ。

```text
collectors/
  brew
  mise
  npm
```

共通インターフェースの概念:

```text
Collect() -> []CLIEntry
```

Collectorは読み取り専用処理のみを行う。

---

## 11. Homebrew Collector

対象:

- Homebrew Formula
- ユーザーが直接インストールしたPackage

デフォルトでは依存packageを除外する。

取得候補:

```bash
brew leaves
brew list --formula --versions
brew info --json=v2
```

最初に直接インストールされたFormula集合を取得し、そのPackageが提供するCommandを取得する。

例:

```text
package: ripgrep
version: 14.1.1
command: rg
source: brew
```

MVPではCommandの完全な列挙が難しい場合、以下の順で取得する。

1. Homebrew metadata
2. packageの `bin` directory
3. package名をfallback

---

## 12. mise Collector

取得候補:

```bash
mise ls --json
```

miseで管理されているツールを取得する。

例:

```text
package: node
version: 24.8.0
command: node
source: mise
```

MVPでは1Packageにつき代表Commandを1つ表示する。

例:

```text
node
python
go
ruby
```

`npm` や `npx` など副次的CommandはMVPでは一覧に出さなくてもよい。

---

## 13. npm Collector

対象:

- npm global package
- top-level packageのみ

取得候補:

```bash
npm list -g --depth=0 --json
```

Command名は `package.json` の `bin` から取得する。

例:

```json
{
  "name": "@openai/codex",
  "version": "0.42.0",
  "bin": {
    "codex": "bin/codex.js"
  }
}
```

変換結果:

```text
command: codex
package: @openai/codex
version: 0.42.0
source: npm
```

1Packageが複数Commandを提供する場合、それぞれ別 `CLIEntry` として扱う。

---

## 14. Manual Entry

専用installerや `curl | sh` 経由で導入されたCLIについては、MVPでは自動判定しない。

必要なら設定ファイルに手動登録できる。

例:

```toml
[[tools]]
command = "claude"
package = "claude"
version_command = "claude --version"
source = "manual"
```

設定ファイル:

```text
~/.config/cliv/config.toml
```

手動登録されたCommandについては `version_command` を実行してバージョンを取得する。

---

## 15. PATHの扱い

MVPではPATH全体のOwnership Resolutionは行わない。

ただし、表示対象のCommandが現在実行可能か確認するため、

```bash
command -v <command>
```

相当の処理を行ってもよい。

取得できたPathは内部データとして保持する。

例:

```text
command: node
path: ~/.local/share/mise/shims/node
```

MVPでは、

- Active Provider判定
- duplicate provider解析
- symlink追跡
- shim解決

は行わない。

---

## 16. 重複

異なるSourceから同じCommandが検出された場合、MVPでは単純に複数行表示する。

例:

```text
COMMAND   VERSION   SOURCE
node      24.8.0    mise
node      23.11.0   brew
```

どちらが正しいか、どちらを削除すべきかは判断しない。

将来的に必要なら `ACTIVE` 列追加を検討する。

---

## 17. 並び順

デフォルトではCommand名の昇順。

例:

```text
claude
codex
jq
node
python
rg
```

同じCommandが複数ある場合:

```text
command
source
version
```

の順で安定ソートする。

---

## 18. バージョン取得

優先順位:

```text
1. Package Manager metadata
2. Manual entry の version_command
3. unknown
```

MVPでは任意のCommandに対して、

```bash
<command> --version
```

を自動実行しない。

理由:

- 出力形式が不定
- 副作用の可能性
- パフォーマンス低下
- エラー処理が複雑化

---

## 19. エラー処理

特定Collectorが失敗しても、他Collectorの結果は表示する。

例:

```text
brew: OK
mise: OK
npm: failed
```

通常出力では可能な限りノイズを抑える。

詳細エラーはstderrへ出す。

未インストールのPackage Managerはエラー扱いしない。

例:

```text
cargo not installed
```

は通常表示しない。

---

## 20. パフォーマンス

目標:

```text
cliv
```

の実行を通常環境で1秒前後に収める。

Collectorは独立しているため、必要であれば並列実行する。

MVPではまず実装単純性を優先し、遅い場合に並列化を検討する。

---

## 21. キャッシュ

MVPでは必須としない。

実行速度に問題が出た場合のみ、

```text
~/.cache/cliv/
```

へのキャッシュを検討する。

キャッシュは観測結果のみ保存する。

---

## 22. 設定ファイル

```text
~/.config/cliv/config.toml
```

例:

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

MVPでは設定項目を増やしすぎない。

---

## 23. アーキテクチャ

```text
         ┌─────────────┐
         │ Collectors  │
         │             │
         │ brew        │
         │ mise        │
         │ npm         │
         └──────┬──────┘
                │
                ▼
           []CLIEntry
                │
                ▼
           Normalize
                │
                ▼
             Sort
                │
        ┌───────┴───────┐
        ▼               ▼
      Table            JSON
```

非常に単純な構造を維持する。

---

## 24. ディレクトリ構成案

```text
cliv/
├── cmd/
│   └── cliv/
│       └── main.*
├── internal/
│   ├── collector/
│   │   ├── brew.*
│   │   ├── mise.*
│   │   └── npm.*
│   ├── model/
│   │   └── entry.*
│   ├── config/
│   │   └── config.*
│   └── output/
│       ├── table.*
│       └── json.*
└── README.md
```

言語によって具体的な構成は変更可能。

---

## 25. MVPのCLI UX

基本:

```bash
cliv
```

Filter:

```bash
cliv --source brew
```

依存を含む:

```bash
cliv --all
```

JSON:

```bash
cliv --json
```

Help:

```bash
cliv --help
```

Version:

```bash
cliv --version
```

MVPではサブコマンドを持たない。

---

## 26. MVP完了条件

以下を満たせばMVP完成とする。

### Homebrew

```text
direct Formula
version
representative command
```

を取得できる。

### mise

```text
tool
version
representative command
```

を取得できる。

### npm

```text
global top-level package
version
bin command
```

を取得できる。

### Unified output

以下の形式で一覧表示できる。

```text
COMMAND   VERSION   SOURCE
```

### JSON

同じデータをJSONで出力できる。

### Read-only

外部Package ManagerやCLI環境を一切変更しない。

---

## 27. MVP後に必要性を見て追加するもの

実際に使って不足を感じた場合のみ追加する。

候補:

```text
cargo collector
uv collector
ACTIVE column
duplicate detection
PATH detail
snapshot
diff
doctor
```

「あると便利そう」という理由だけでは追加しない。

---

## 28. MVPの中心価値

`cliv` の価値は、以下の1コマンドに集約する。

```bash
cliv
```

これにより、

```text
COMMAND   VERSION    SOURCE
rg        14.1.1     brew
node      24.8.0     mise
python    3.13.7     mise
codex     0.42.0     npm
claude    1.0.113    manual
```

のように、異なる方法でインストールされたCLIを1つの一覧として確認できる。

MVPでは、

> **一覧を見る**

というユースケースだけを高品質に実現する。
