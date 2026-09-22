// ~/.config/cliv/config.tomlを読み込む簡易設定ローダーです。
// 依存を増やさないため、MVPで必要なTOMLの形だけを軽く解釈します。

package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Collectors map[string]bool
	Tools      []ManualTool
}

type ManualTool struct {
	Command        string
	Package        string
	Source         string
	VersionCommand string
}

func Load() Config {
	cfg := Config{Collectors: map[string]bool{"brew": true, "mise": true, "npm": true}}
	h, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	path := filepath.Join(h, ".config", "cliv", "config.toml")
	f, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer f.Close()

	section := ""
	var current *ManualTool
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch line {
		case "[collectors]":
			section = "collectors"
			current = nil
		case "[[tools]]":
			section = "tools"
			cfg.Tools = append(cfg.Tools, ManualTool{Source: "manual"})
			current = &cfg.Tools[len(cfg.Tools)-1]
		default:
			k, v, ok := parseKV(line)
			if !ok {
				continue
			}
			if section == "collectors" {
				cfg.Collectors[k] = parseBool(v, true)
			} else if section == "tools" && current != nil {
				switch k {
				case "command":
					current.Command = v
				case "package":
					current.Package = v
				case "source":
					current.Source = v
				case "version_command":
					current.VersionCommand = v
				}
			}
		}
	}
	return cfg
}

func parseKV(line string) (string, string, bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	k := strings.TrimSpace(parts[0])
	v := strings.TrimSpace(parts[1])
	v = strings.Trim(v, `"'`)
	return k, v, k != ""
}

func parseBool(v string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true":
		return true
	case "false":
		return false
	default:
		return def
	}
}
