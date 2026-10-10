package model

// CLIEntryはインストール済みツール／パッケージの1つのバージョンを表します。
type CLIEntry struct {
	Tool    string `json:"tool"`
	Version string `json:"version"`
	Source  string `json:"source"`
	Direct  bool   `json:"direct"`
}
