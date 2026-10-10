package collector

import (
	"testing"

	"github.com/munakata-hisashi/cliv/internal/config"
)

func TestBrewDirectAndAll(t *testing.T) {
	data := []byte(`{"formulae":[{"name":"cmake","installed":[{"version":"4.2.1","installed_on_request":true}]},{"name":"library","installed":[{"version":"1.0","installed_on_request":false}]}]}`)
	entries, err := parseBrew(data, false)
	if err != nil || len(entries) != 1 || entries[0].Tool != "cmake" || entries[0].Version != "4.2.1" || !entries[0].Direct {
		t.Fatalf("direct: %+v, %v", entries, err)
	}
	entries, err = parseBrew(data, true)
	if err != nil || len(entries) != 2 || entries[1].Tool != "library" || entries[1].Direct {
		t.Fatalf("all: %+v, %v", entries, err)
	}
	for _, invalid := range []string{`{}`, `invalid`} {
		if _, err := parseBrew([]byte(invalid), true); err == nil {
			t.Fatal("expected invalid metadata error")
		}
	}
}

func TestMiseInstalledOnly(t *testing.T) {
	entries, err := parseMise([]byte(`{"node":[{"version":"24.8.0","installed":true},{"requested_version":"25","installed":false}],"go":[{"version":"1.24","installed":true}]}`))
	if err != nil || len(entries) != 2 {
		t.Fatalf("entries: %+v, %v", entries, err)
	}
	versions := map[string]string{}
	for _, e := range entries {
		versions[e.Tool] = e.Version
	}
	if versions["node"] != "24.8.0" || versions["go"] != "1.24" {
		t.Fatalf("versions: %+v", versions)
	}
	if _, err := parseMise([]byte(`invalid`)); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestMiseToolNamesWithoutExecutables(t *testing.T) {
	for _, name := range []string{"npm:@openai/codex", "npm:@earendil-works/pi-coding-agent", "pipx:httpie", "node"} {
		t.Run(name, func(t *testing.T) {
			entry := miseEntry(map[string]any{"name": name, "version": "1.1.0", "installed": true})
			if entry.Tool != name || entry.Version != "1.1.0" {
				t.Fatalf("entry: %+v", entry)
			}
		})
	}
}

func TestMiseMultipleInstalledVersions(t *testing.T) {
	entries, err := parseMise([]byte(`{"npm:@openai/codex":[{"version":"0.155.1","installed":true,"active":false},{"version":"0.162.0","installed":true,"active":true}]}`))
	if err != nil || len(entries) != 2 {
		t.Fatalf("entries: %+v, %v", entries, err)
	}
	for _, entry := range entries {
		if entry.Tool != "npm:@openai/codex" {
			t.Fatalf("entry: %+v", entry)
		}
	}
}

func TestManualVersion(t *testing.T) {
	if got := manualVersion(`printf '2.1.63 (Claude Code)\n'`); got != "2.1.63" {
		t.Fatalf("version = %q", got)
	}
}

func TestManualToolName(t *testing.T) {
	entries, err := (Manual{Tools: []config.ManualTool{
		{Command: "first-command", Package: "tool-name"},
		{Command: "fallback"},
		{Command: "second-command", Package: "tool-name"},
	}}).Collect(false)
	if err != nil || len(entries) != 2 || entries[0].Tool != "tool-name" || entries[1].Tool != "fallback" {
		t.Fatalf("entries: %+v, %v", entries, err)
	}
}

func TestNPMPackageNames(t *testing.T) {
	entries, err := parseNPM([]byte(`{"dependencies":{"@example/tool":{"version":"1.2.3"},"library":{}}}`))
	if err != nil || len(entries) != 2 {
		t.Fatalf("entries: %+v, %v", entries, err)
	}
	versions := map[string]string{}
	for _, e := range entries {
		versions[e.Tool] = e.Version
	}
	if versions["@example/tool"] != "1.2.3" || versions["library"] != "unknown" {
		t.Fatalf("versions: %+v", versions)
	}
	if _, err := parseNPM([]byte(`invalid`)); err == nil {
		t.Fatal("expected parse error")
	}
}
