package collector

import (
	"encoding/json"
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
	return parseMise(out), nil
}

func parseMise(data []byte) []model.CLIEntry {
	var raw any
	if json.Unmarshal(data, &raw) != nil {
		return nil
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
		// Some mise versions return {"node": [{...}], "python": [{...}]}.
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
	}
	out := entries[:0]
	for _, e := range entries {
		if e.Command != "" {
			out = append(out, e)
		}
	}
	return out
}

func miseEntry(m map[string]any) model.CLIEntry {
	name := stringField(m, "name", "tool", "plugin", "short")
	version := stringField(m, "version", "installed_version")
	if version == "" {
		version = stringField(m, "requested_version")
	}
	// If version contains source/status suffixes, keep the first token as MVP metadata.
	fields := strings.Fields(versionString(version))
	if len(fields) > 0 {
		version = fields[0]
	}
	if version == "" {
		version = "unknown"
	}
	return model.CLIEntry{Command: name, Package: name, Version: version, Source: "mise", Path: commandPath(name), Direct: true}
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
