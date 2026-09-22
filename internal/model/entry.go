package model

// cliv内で扱うデータ構造を定義します。
// package名とcommand名は一致しないため、両方を保持します。
// CLIEntryは、インストール済みpackageが提供する1つのcommandを表します。
type CLIEntry struct {
	Command string `json:"command"`
	Package string `json:"package"`
	Version string `json:"version"`
	Source  string `json:"source"`
	Path    string `json:"path,omitempty"`
	Direct  bool   `json:"direct"`
}
