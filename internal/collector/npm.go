// npm global packageからbinコマンドを集めます。
// package名とcommand名が違うため、package.jsonのbinを正として使います。

package collector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	var list struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, err
	}
	rootBytes, err := run("npm", "root", "-g")
	if err != nil {
		return nil, err
	}
	root := strings.TrimSpace(string(rootBytes))
	var entries []model.CLIEntry
	for pkg, dep := range list.Dependencies {
		bins := npmBins(root, pkg)
		if len(bins) == 0 {
			continue
		}
		sort.Strings(bins)
		for _, cmd := range bins {
			entries = append(entries, model.CLIEntry{Command: cmd, Package: pkg, Version: firstNonEmpty(dep.Version), Source: "npm", Path: commandPath(cmd), Direct: true})
		}
	}
	return entries, nil
}

func npmBins(root, pkg string) []string {
	pkgJSON := filepath.Join(root, filepath.FromSlash(pkg), "package.json")
	b, err := os.ReadFile(pkgJSON)
	if err != nil {
		return nil
	}
	var p struct {
		Name string          `json:"name"`
		Bin  json.RawMessage `json:"bin"`
	}
	if json.Unmarshal(b, &p) != nil || len(p.Bin) == 0 || string(p.Bin) == "null" {
		return nil
	}
	var binStr string
	if json.Unmarshal(p.Bin, &binStr) == nil && binStr != "" {
		name := p.Name
		if name == "" {
			name = pkg
		}
		return []string{strings.TrimPrefix(filepath.Base(name), "@")}
	}
	var binMap map[string]any
	if json.Unmarshal(p.Bin, &binMap) == nil {
		out := make([]string, 0, len(binMap))
		for name := range binMap {
			out = append(out, name)
		}
		return out
	}
	return nil
}
