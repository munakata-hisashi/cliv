package collector

import (
	"strings"
	"testing"

	"github.com/munakata-hisashi/cliv/internal/config"
)

func TestManualFailuresKeepOtherTools(t *testing.T) {
	entries, err := (Manual{Tools: []config.ManualTool{
		{Command: "bad", VersionCommand: "exit 1"},
		{Command: "good", VersionCommand: "printf '1.2.3\\n'"},
	}}).Collect(false)
	if err == nil || !strings.Contains(err.Error(), "bad") || len(entries) != 2 || entries[0].Version != "unknown" || entries[1].Version != "1.2.3" {
		t.Fatalf("entries=%+v err=%v", entries, err)
	}
}

func TestManualUnrecognizedVersionIsFailure(t *testing.T) {
	for _, cmd := range []string{"printf ''", "printf 'no version'"} {
		if version, err := manualVersion(cmd); err == nil || version != "unknown" {
			t.Fatalf("version=%q err=%v", version, err)
		}
	}
}

func TestNPMErrorMetadata(t *testing.T) {
	for _, data := range []string{
		`{"error":{"code":"FAIL"},"dependencies":{"tool":{"version":"1.2.3"}}}`,
		`{"problems":["missing dependency"],"dependencies":{"tool":{"version":"1.2.3"}}}`,
	} {
		entries, err := parseNPM([]byte(data))
		if err == nil || len(entries) != 1 || entries[0].Tool != "tool" {
			t.Fatalf("entries=%+v err=%v", entries, err)
		}
	}
}
