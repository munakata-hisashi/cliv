package collector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBrewDirectAndAll(t *testing.T) {
	prefix := t.TempDir()
	bin := filepath.Join(prefix, "Cellar", "ripgrep", "14.1", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "rg-real"), []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("rg-real", filepath.Join(bin, "rg")); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{"formulae":[{"name":"ripgrep","installed":[{"version":"14.1","installed_on_request":true}]},{"name":"library","installed":[{"version":"1.0","installed_on_request":false}]}]}`)
	entries, err := parseBrew(data, prefix, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Command != "rg" || entries[1].Command != "rg-real" || !entries[0].Direct {
		t.Fatalf("direct: %+v", entries)
	}
	entries, err = parseBrew(data, prefix, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 || entries[2].Command != "library" || entries[2].Direct {
		t.Fatalf("all: %+v", entries)
	}
	if _, err := parseBrew([]byte(`{}`), prefix, true); err == nil {
		t.Fatal("expected invalid metadata error")
	}
}

func TestMiseInstalledOnly(t *testing.T) {
	entries, err := parseMise([]byte(`{"node":[{"version":"24.8.0","installed":true},{"requested_version":"25","installed":false}],"go":[{"version":"1.24","installed":true}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries: %+v", entries)
	}
	versions := map[string]string{}
	for _, e := range entries {
		versions[e.Command] = e.Version
	}
	if versions["node"] != "24.8.0" || versions["go"] != "1.24" {
		t.Fatalf("versions: %+v", versions)
	}
	if _, err := parseMise([]byte(`invalid`)); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestMiseBackendCommand(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	entry := miseEntry(map[string]any{"name": "npm:@openai/codex", "install_path": filepath.Dir(bin), "version": "0.42.0", "installed": true})
	if entry.Command != "codex" || entry.Package != "npm:@openai/codex" {
		t.Fatalf("entry: %+v", entry)
	}
}

func TestManualVersion(t *testing.T) {
	if got := manualVersion(`printf '2.1.63 (Claude Code)\n'`); got != "2.1.63" {
		t.Fatalf("version = %q", got)
	}
}

func TestNPMBins(t *testing.T) {
	root := t.TempDir()
	for pkg, bin := range map[string]any{"@openai/codex": map[string]string{"codex": "bin/codex.js"}, "typescript": map[string]string{"tsc": "bin/tsc", "tsserver": "bin/tsserver"}, "foo": "cli.js"} {
		dir := filepath.Join(root, filepath.FromSlash(pkg))
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		data, _ := json.Marshal(map[string]any{"name": pkg, "bin": bin})
		if err := os.WriteFile(filepath.Join(dir, "package.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	for pkg, want := range map[string]int{"@openai/codex": 1, "typescript": 2, "foo": 1} {
		if got := npmBins(root, pkg); len(got) != want {
			t.Fatalf("%s: %v", pkg, got)
		}
	}
	if got := npmBins(root, "@openai/codex"); got[0] != "codex" {
		t.Fatal(got)
	}
}
