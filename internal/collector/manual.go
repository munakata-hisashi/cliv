// 設定ファイルに書かれた手動登録CLIを集めます。
// curl | shなど由来を自動判定しづらいものを、MVPでは明示登録で扱います。

package collector

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"

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
		// Configured entries are always manual; they must not masquerade as a
		// package manager's metadata or bypass --source filtering.
		source := "manual"
		version := "unknown"
		if t.VersionCommand != "" {
			version = manualVersion(t.VersionCommand)
		}
		entries = append(entries, model.CLIEntry{Command: t.Command, Package: pkg, Version: version, Source: source, Path: commandPath(t.Command), Direct: true})
	}
	return entries, nil
}

func manualVersion(command string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return "unknown"
	}
	// Version output is not standardized: e.g. "2.1.63 (Claude Code)" or
	// "claude 1.0.113". Prefer a version-like token over the last word.
	if version := regexp.MustCompile(`\b[vV]?[0-9]+(?:\.[0-9]+)+(?:[-+._][0-9A-Za-z]+)*\b`).FindString(v); version != "" {
		return strings.TrimPrefix(strings.TrimPrefix(version, "v"), "V")
	}
	return "unknown"
}
