package model

// CLIEntry represents one command provided by an installed package.
type CLIEntry struct {
	Command string `json:"command"`
	Package string `json:"package"`
	Version string `json:"version"`
	Source  string `json:"source"`
	Path    string `json:"path,omitempty"`
	Direct  bool   `json:"direct"`
}
