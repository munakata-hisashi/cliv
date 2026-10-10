package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/munakata-hisashi/cliv/internal/model"
)

type UV struct{}

func (UV) Name() string { return "uv" }

func (UV) Collect(all bool) ([]model.CLIEntry, error) {
	if !commandExists("uv") {
		return nil, nil
	}
	// Disable color to keep the human-readable output parseable.
	data, err := run("uv", "tool", "list", "--color", "never")
	if err != nil {
		return nil, err
	}
	return parseUV(data)
}

func parseUV(data []byte) ([]model.CLIEntry, error) {
	var entries []model.CLIEntry
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "- ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || len(fields[1]) < 2 || fields[1][0] != 'v' || fields[1][1] < '0' || fields[1][1] > '9' {
			return nil, fmt.Errorf("parse uv tool list: unexpected line %q", line)
		}
		entries = append(entries, model.CLIEntry{
			Tool: fields[0], Version: strings.TrimPrefix(fields[1], "v"), Source: "uv", Direct: true,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse uv tool list: %w", err)
	}
	return entries, nil
}
