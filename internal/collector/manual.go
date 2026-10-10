// 設定ファイルに書かれた手動登録CLIを集めます。
// curl | shなど由来を自動判定しづらいものを、MVPでは明示登録で扱います。

package collector

import (
	"context"
	"errors"
	"fmt"
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
	var failures []error
	seen := make(map[[2]string]bool)
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
			var err error
			version, err = manualVersion(t.VersionCommand)
			if err != nil {
				failures = append(failures, fmt.Errorf("%s: %w", pkg, err))
			}
		}
		key := [2]string{pkg, version}
		if seen[key] {
			continue
		}
		seen[key] = true
		entries = append(entries, model.CLIEntry{Tool: pkg, Version: version, Source: source, Direct: true})
	}
	return entries, errors.Join(failures...)
}

func manualVersion(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "unknown", fmt.Errorf("version command: %w", ctx.Err())
		}
		var exit *exec.ExitError
		if errors.As(err, &exit) && strings.TrimSpace(string(exit.Stderr)) != "" {
			err = fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exit.Stderr)))
		}
		return "unknown", fmt.Errorf("version command: %w", err)
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return "unknown", fmt.Errorf("version command returned empty output")
	}
	// Version output is not standardized: e.g. "2.1.63 (Claude Code)" or
	// "claude 1.0.113". Prefer a version-like token over the last word.
	if version := regexp.MustCompile(`\b[vV]?[0-9]+(?:\.[0-9]+)+(?:[-+._][0-9A-Za-z]+)*\b`).FindString(v); version != "" {
		return strings.TrimPrefix(strings.TrimPrefix(version, "v"), "V"), nil
	}
	return "unknown", fmt.Errorf("version command returned no recognizable version")
}
