// miseが管理しているツールを集めます。
// MVPでは1ツール1代表コマンドとして扱い、副次的なコマンドは追いません。

package collector

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/munakata-hisashi/cliv/internal/model"
)

type Mise struct{}

func (Mise) Name() string { return "mise" }

func (Mise) Collect(all bool) ([]model.CLIEntry, error) {
	if !commandExists("mise") {
		return nil, nil
	}
	out, err := run("mise", "ls", "--json")
	if err != nil {
		return nil, err
	}
	return parseMise(out)
}

func parseMise(data []byte) ([]model.CLIEntry, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse mise ls: %w", err)
	}
	var entries []model.CLIEntry
	switch v := raw.(type) {
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				entries = append(entries, miseEntry(m))
			}
		}
	case map[string]any:
		// miseのバージョンによりJSON形状が違うため、代表的な2形式を受けます。
		for name, val := range v {
			switch vv := val.(type) {
			case []any:
				for _, item := range vv {
					if m, ok := item.(map[string]any); ok {
						m["name"] = name
						entries = append(entries, miseEntry(m))
					}
				}
			case map[string]any:
				vv["name"] = name
				entries = append(entries, miseEntry(vv))
			}
		}
	default:
		return nil, fmt.Errorf("parse mise ls: unexpected JSON shape")
	}
	out := entries[:0]
	for _, e := range entries {
		if e.Command != "" {
			out = append(out, e)
		}
	}
	return out, nil
}

func miseEntry(m map[string]any) model.CLIEntry {
	if installed, ok := m["installed"].(bool); ok && !installed {
		return model.CLIEntry{}
	}
	name := stringField(m, "name", "tool", "plugin", "short")
	command := name
	if strings.Contains(name, ":") {
		// Backend-qualified tool IDs are package IDs, not shell commands.
		// Use the installed bin directory to pick one representative command.
		bins := executableFiles(filepath.Join(stringField(m, "install_path"), "bin"))
		if len(bins) == 0 {
			return model.CLIEntry{}
		}
		command = bins[0]
	}
	version := stringField(m, "version", "installed_version")
	if version == "" {
		version = stringField(m, "requested_version")
	}
	// バージョン以外の補足が混ざる場合があるため、先頭トークンだけを採用します。
	fields := strings.Fields(versionString(version))
	if len(fields) > 0 {
		version = fields[0]
	}
	if version == "" {
		version = "unknown"
	}
	return model.CLIEntry{Command: command, Package: name, Version: version, Source: "mise", Path: commandPath(command), Direct: true}
}

func versionString(s string) string {
	return strings.TrimSpace(s)
}

func stringField(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch x := v.(type) {
			case string:
				return strings.TrimSpace(x)
			case float64:
				return strings.TrimSpace(strings.TrimRight(strings.TrimRight(jsonNumber(x), "0"), "."))
			}
		}
	}
	return ""
}

func jsonNumber(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}
