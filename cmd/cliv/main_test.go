package main

import (
	"testing"

	"github.com/munakata-hisashi/cliv/internal/config"
)

func TestColorEnabled(t *testing.T) {
	cases := []struct {
		mode              string
		tty               bool
		noColor, termName string
		want              bool
	}{
		{"auto", true, "", "xterm", true},
		{"auto", false, "", "xterm", false},
		{"auto", true, "1", "xterm", false},
		{"auto", true, "", "dumb", false},
		{"always", false, "1", "dumb", true},
		{"never", true, "", "xterm", false},
	}
	for _, c := range cases {
		if got := colorEnabled(c.mode, c.tty, c.noColor, c.termName); got != c.want {
			t.Errorf("colorEnabled(%+v) = %v", c, got)
		}
	}
}

func TestUVCollectorEnabled(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		collectors := enabledCollectors(config.Config{Collectors: map[string]bool{"uv": enabled}})
		if enabled {
			if len(collectors) != 1 || collectors[0].Name() != "uv" {
				t.Fatalf("uv enabled: %v", collectors)
			}
		} else if len(collectors) != 0 {
			t.Fatalf("uv disabled: %v", collectors)
		}
	}
}

func TestCargoAndGoCollectorsEnabled(t *testing.T) {
	for _, name := range []string{"cargo", "go"} {
		for _, enabled := range []bool{false, true} {
			collectors := enabledCollectors(config.Config{Collectors: map[string]bool{name: enabled}})
			if enabled && (len(collectors) != 1 || collectors[0].Name() != name) {
				t.Fatalf("%s enabled: %v", name, collectors)
			}
			if !enabled && len(collectors) != 0 {
				t.Fatalf("%s disabled: %v", name, collectors)
			}
		}
	}
}
