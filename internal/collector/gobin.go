package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/munakata-hisashi/cliv/internal/model"
)

// Go lists binaries installed with `go install` in GOBIN (or the first GOPATH
// entry's bin). Tool is the main package path, i.e. what `go install` takes,
// rather than the binary name.
type Go struct{}

func (Go) Name() string { return "go" }

func (Go) Collect(all bool) ([]model.CLIEntry, error) {
	if !commandExists("go") {
		return nil, nil
	}
	env, err := run("go", "env", "GOBIN", "GOPATH")
	if err != nil {
		return nil, err
	}
	dir := goBinDir(string(env))
	if dir == "" {
		return nil, nil
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil, nil
	}
	// Directory mode reads build info of every Go binary and skips other files.
	data, err := run("go", "version", "-m", dir)
	if err != nil {
		return nil, err
	}
	return parseGoVersion(data)
}

func goBinDir(env string) string {
	lines := strings.Split(strings.ReplaceAll(env, "\r", ""), "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		return strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		for _, p := range filepath.SplitList(strings.TrimSpace(lines[1])) {
			if p != "" {
				return filepath.Join(p, "bin")
			}
		}
	}
	return ""
}

// parseGoVersion parses `go version -m` output: a "<file>: goX" header
// followed by tab-indented "path" and "mod" lines.
func parseGoVersion(data []byte) ([]model.CLIEntry, error) {
	var entries []model.CLIEntry
	seen := map[model.CLIEntry]bool{}
	var path, version string
	inBinary := false
	flush := func() {
		if inBinary && path != "" {
			e := model.CLIEntry{Tool: path, Version: firstNonEmpty(version), Source: "go", Direct: true}
			if !seen[e] {
				seen[e] = true
				entries = append(entries, e)
			}
		}
		path, version, inBinary = "", "", false
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		raw := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(raw) == "" {
			continue
		}
		if raw[0] != '\t' && raw[0] != ' ' {
			flush()
			if !strings.Contains(raw, ": go") && !strings.Contains(raw, ": devel") {
				return nil, fmt.Errorf("parse go version -m: unexpected line %q", raw)
			}
			inBinary = true
			continue
		}
		if !inBinary {
			return nil, fmt.Errorf("parse go version -m: unexpected line %q", raw)
		}
		fields := strings.Fields(raw)
		switch {
		case len(fields) >= 2 && fields[0] == "path":
			path = fields[1]
		case len(fields) >= 3 && fields[0] == "mod":
			version = strings.TrimPrefix(fields[2], "v")
		}
	}
	flush()
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse go version -m: %w", err)
	}
	return entries, nil
}
