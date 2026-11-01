package collector

import (
	"path/filepath"
	"testing"
)

func TestParseCargoEmptyAndInvalid(t *testing.T) {
	if e, err := parseCargo(nil); err != nil || len(e) != 0 {
		t.Fatalf("empty: %+v %v", e, err)
	}
	for _, s := range []string{"ripgrep 14.1.1:", "ripgrep v14.1.1", "garbage"} {
		if _, err := parseCargo([]byte(s)); err == nil {
			t.Fatalf("expected error for %q", s)
		}
	}
}

func TestParseGoVersionEmptyAndInvalid(t *testing.T) {
	if e, err := parseGoVersion(nil); err != nil || len(e) != 0 {
		t.Fatalf("empty: %+v %v", e, err)
	}
	for _, s := range []string{"garbage", "\tpath\tx"} {
		if _, err := parseGoVersion([]byte(s)); err == nil {
			t.Fatalf("expected error for %q", s)
		}
	}
}

func TestGoBinDir(t *testing.T) {
	sep := string(filepath.ListSeparator)
	cases := map[string]string{
		"/custom/bin\n/home/u/go\n":       "/custom/bin",
		"\n/home/u/go" + sep + "/other\n": filepath.Join("/home/u/go", "bin"),
		"\n\n":                            "",
	}
	for in, want := range cases {
		if got := goBinDir(in); got != want {
			t.Errorf("goBinDir(%q) = %q, want %q", in, got, want)
		}
	}
}
