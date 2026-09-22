package collector

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/munakata-hisashi/cliv/internal/model"
)

type Brew struct{}

func (Brew) Name() string { return "brew" }

func (Brew) Collect(all bool) ([]model.CLIEntry, error) {
	if !commandExists("brew") {
		return nil, nil
	}

	direct := map[string]bool{}
	if !all {
		out, err := run("brew", "leaves")
		if err != nil {
			return nil, err
		}
		s := bufio.NewScanner(bytes.NewReader(out))
		for s.Scan() {
			name := strings.TrimSpace(s.Text())
			if name != "" {
				direct[name] = true
			}
		}
	}

	out, err := run("brew", "list", "--formula", "--versions")
	if err != nil {
		return nil, err
	}

	var entries []model.CLIEntry
	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) < 1 {
			continue
		}
		pkg := fields[0]
		if !all && !direct[pkg] {
			continue
		}
		version := "unknown"
		if len(fields) > 1 {
			version = fields[1]
		}
		commands := brewCommands(pkg)
		if len(commands) == 0 {
			commands = []string{pkg}
		}
		for _, cmd := range commands {
			entries = append(entries, model.CLIEntry{
				Command: cmd,
				Package: pkg,
				Version: version,
				Source:  "brew",
				Path:    commandPath(cmd),
				Direct:  all || direct[pkg],
			})
		}
	}
	return entries, nil
}

func brewCommands(pkg string) []string {
	out, err := run("brew", "--prefix", pkg)
	if err != nil {
		return nil
	}
	prefix := strings.TrimSpace(string(out))
	if prefix == "" {
		return nil
	}
	return executableFiles(prefix + "/bin")
}
