package collector

import (
	"testing"

	"github.com/munakata-hisashi/cliv/internal/model"
)

func TestParseUV(t *testing.T) {
	entries, err := parseUV([]byte("plamo-translate v1.0.5\n- plamo-translate\n\nhttpie v3.2.4\n- http\n- https\n- httpie\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []model.CLIEntry{
		{Tool: "plamo-translate", Version: "1.0.5", Source: "uv", Direct: true},
		{Tool: "httpie", Version: "3.2.4", Source: "uv", Direct: true},
	}
	if len(entries) != len(want) {
		t.Fatalf("entries: %+v", entries)
	}
	for i := range want {
		if entries[i] != want[i] {
			t.Fatalf("entry %d: %+v, want %+v", i, entries[i], want[i])
		}
	}
}

func TestParseUVEmpty(t *testing.T) {
	entries, err := parseUV(nil)
	if err != nil || len(entries) != 0 {
		t.Fatalf("entries: %+v, %v", entries, err)
	}
}

func TestParseUVAnnotatedVersion(t *testing.T) {
	entries, err := parseUV([]byte("example v1.2.3rc1 (from git+https://example.com/tool)\r\n- example\r\n"))
	if err != nil || len(entries) != 1 || entries[0].Version != "1.2.3rc1" {
		t.Fatalf("entries: %+v, %v", entries, err)
	}
}

func TestParseUVInvalid(t *testing.T) {
	for _, data := range []string{"unexpected", "tool unknown", "tool v", "tool version"} {
		if _, err := parseUV([]byte(data)); err == nil {
			t.Fatalf("expected error for %q", data)
		}
	}
}
