package main

import (
	"testing"

	"github.com/munakata-hisashi/cliv/internal/config"
)

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
