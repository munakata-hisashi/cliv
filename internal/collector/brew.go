package collector

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/munakata-hisashi/cliv/internal/model"
)

type Brew struct{}

func (Brew) Name() string { return "brew" }

type brewInfo struct {
	Formulae []struct {
		Name      string `json:"name"`
		Installed []struct {
			Version            string `json:"version"`
			InstalledOnRequest *bool  `json:"installed_on_request"`
		} `json:"installed"`
	} `json:"formulae"`
}

func (Brew) Collect(all bool) ([]model.CLIEntry, error) {
	if !commandExists("brew") {
		return nil, nil
	}
	// installed_on_request distinguishes explicitly installed formulae from dependencies;
	// brew leaves does not (a requested formula may also be a dependency).
	data, err := run("brew", "info", "--json=v2", "--installed")
	if err != nil {
		return nil, err
	}
	prefix, err := run("brew", "--prefix")
	if err != nil {
		return nil, err
	}
	return parseBrew(data, filepath.Clean(strings.TrimSpace(string(prefix))), all)
}

func parseBrew(data []byte, prefix string, all bool) ([]model.CLIEntry, error) {
	var info brewInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parse brew info: %w", err)
	}
	if info.Formulae == nil {
		return nil, fmt.Errorf("parse brew info: missing formulae")
	}
	var entries []model.CLIEntry
	for _, f := range info.Formulae {
		if f.Name == "" {
			continue
		}
		for _, installed := range f.Installed {
			direct := installed.InstalledOnRequest != nil && *installed.InstalledOnRequest
			if !all && !direct {
				continue
			}
			// Inspect the keg itself so unlinked formulae are included. Formulae
			// without an executable in bin or sbin do not provide a CLI entry.
			keg := filepath.Join(prefix, "Cellar", f.Name, installed.Version)
			seen := make(map[string]bool)
			for _, dir := range []string{"bin", "sbin"} {
				for _, command := range executableFiles(filepath.Join(keg, dir)) {
					if seen[command] {
						continue
					}
					seen[command] = true
					entries = append(entries, model.CLIEntry{
						Command: command, Package: f.Name, Version: firstNonEmpty(installed.Version),
						Source: "brew", Path: commandPath(command), Direct: direct,
					})
				}
			}
		}
	}
	return entries, nil
}
