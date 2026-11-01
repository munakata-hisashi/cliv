package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/munakata-hisashi/cliv/internal/model"
)

// Cargo lists crates installed with `cargo install`.
// `cargo install --list` is the documented interface; ~/.cargo/.crates2.json is
// an internal file whose location depends on CARGO_HOME / --root.
type Cargo struct{}

func (Cargo) Name() string { return "cargo" }

func (Cargo) Collect(all bool) ([]model.CLIEntry, error) {
	if !commandExists("cargo") {
		return nil, nil
	}
	data, err := run("cargo", "install", "--list")
	if err != nil {
		return nil, err
	}
	return parseCargo(data)
}

// parseCargo reads lines like "ripgrep v14.1.1:" or "tool v0.1.0 (/path):".
// Indented lines list executables and are not shown.
func parseCargo(data []byte) ([]model.CLIEntry, error) {
	var entries []model.CLIEntry
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		raw := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(raw) == "" || raw[0] == ' ' || raw[0] == '\t' {
			continue
		}
		line := strings.TrimSpace(raw)
		fields := strings.Fields(strings.TrimSuffix(line, ":"))
		if !strings.HasSuffix(line, ":") || len(fields) < 2 || len(fields[1]) < 2 || fields[1][0] != 'v' || fields[1][1] < '0' || fields[1][1] > '9' {
			return nil, fmt.Errorf("parse cargo install --list: unexpected line %q", line)
		}
		entries = append(entries, model.CLIEntry{
			Tool: fields[0], Version: strings.TrimSuffix(strings.TrimPrefix(fields[1], "v"), ":"), Source: "cargo", Direct: true,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse cargo install --list: %w", err)
	}
	return entries, nil
}
