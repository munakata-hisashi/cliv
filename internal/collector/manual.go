package collector

import (
	"os/exec"
	"strings"

	"github.com/munakata-hisashi/cliv/internal/config"
	"github.com/munakata-hisashi/cliv/internal/model"
)

type Manual struct {
	Tools []config.ManualTool
}

func (Manual) Name() string { return "manual" }

func (m Manual) Collect(all bool) ([]model.CLIEntry, error) {
	var entries []model.CLIEntry
	for _, t := range m.Tools {
		if strings.TrimSpace(t.Command) == "" {
			continue
		}
		pkg := t.Package
		if pkg == "" {
			pkg = t.Command
		}
		source := t.Source
		if source == "" {
			source = "manual"
		}
		version := "unknown"
		if t.VersionCommand != "" {
			version = manualVersion(t.VersionCommand)
		}
		entries = append(entries, model.CLIEntry{Command: t.Command, Package: pkg, Version: version, Source: source, Path: commandPath(t.Command), Direct: true})
	}
	return entries, nil
}

func manualVersion(command string) string {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return "unknown"
	}
	return strings.Fields(v)[len(strings.Fields(v))-1]
}
