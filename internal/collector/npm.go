// npmのグローバルパッケージ名とバージョンを集めます。
package collector

import (
	"encoding/json"

	"github.com/munakata-hisashi/cliv/internal/model"
)

type NPM struct{}

func (NPM) Name() string { return "npm" }

func (NPM) Collect(all bool) ([]model.CLIEntry, error) {
	if !commandExists("npm") {
		return nil, nil
	}
	out, err := run("npm", "list", "-g", "--depth=0", "--json")
	if err != nil && len(out) == 0 {
		return nil, err
	}
	return parseNPM(out)
}

func parseNPM(data []byte) ([]model.CLIEntry, error) {
	var list struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	var entries []model.CLIEntry
	for pkg, dep := range list.Dependencies {
		entries = append(entries, model.CLIEntry{
			Tool: pkg, Version: firstNonEmpty(dep.Version),
			Source: "npm", Direct: true,
		})
	}
	return entries, nil
}
