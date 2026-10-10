package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/munakata-hisashi/cliv/internal/model"
)

func TestTableGroupsBySource(t *testing.T) {
	entries := []model.CLIEntry{
		{Tool: "zeta", Version: "1", Source: "mise"},
		{Tool: "rg", Version: "15", Source: "brew"},
		{Tool: "alpha", Version: "2", Source: "mise"},
		{Tool: "claude", Version: "3", Source: "manual"},
	}
	var buf bytes.Buffer
	if err := Table(&buf, entries); err != nil {
		t.Fatal(err)
	}
	want := "brew\nTOOL   VERSION\nrg     15\n\nmise\nTOOL    VERSION\nalpha   2\nzeta    1\n\nmanual\nTOOL     VERSION\nclaude   3\n"
	if got := buf.String(); got != want {
		t.Errorf("Table() =\n%q\nwant\n%q", got, want)
	}
	if entries[0].Tool != "zeta" {
		t.Fatal("Table changed input ordering")
	}
}

func TestJSONFields(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, []model.CLIEntry{{Tool: "npm:@openai/codex", Version: "0.162.0", Source: "mise", Direct: true}}); err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0]) != 4 || got[0]["tool"] != "npm:@openai/codex" || got[0]["version"] != "0.162.0" || got[0]["source"] != "mise" || got[0]["direct"] != true {
		t.Fatalf("JSON fields: %s", buf.String())
	}
}

func TestTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := Table(&buf, nil); err != nil || buf.Len() != 0 {
		t.Fatalf("Table(nil) = %q, %v", buf.String(), err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestTableWriteError(t *testing.T) {
	if err := Table(failingWriter{}, []model.CLIEntry{{Tool: "rg", Source: "brew"}}); err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("Table() error = %v", err)
	}
}
