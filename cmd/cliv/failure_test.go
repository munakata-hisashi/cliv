package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/munakata-hisashi/cliv/internal/collector"
	"github.com/munakata-hisashi/cliv/internal/model"
)

type fakeCollector struct {
	name    string
	entries []model.CLIEntry
	err     error
	calls   int
}

func (c *fakeCollector) Name() string { return c.name }
func (c *fakeCollector) Collect(bool) ([]model.CLIEntry, error) {
	c.calls++
	return c.entries, c.err
}

func TestPartialFailure(t *testing.T) {
	for _, jsonOut := range []bool{false, true} {
		var stdout, stderr bytes.Buffer
		failed := &fakeCollector{name: "brew", err: errors.New("metadata unavailable")}
		good := &fakeCollector{name: "uv", entries: []model.CLIEntry{{Tool: "httpie", Version: "3.2.4", Source: "uv", Direct: true}}}
		code := listTools(&stdout, &stderr, []collector.Collector{failed, good}, "", false, jsonOut, false)
		if code != 1 || good.calls != 1 || !strings.Contains(stdout.String(), "httpie") {
			t.Fatalf("partial result: exit=%d, stdout=%q", code, stdout.String())
		}
		if !strings.Contains(stderr.String(), "cliv: brew: metadata unavailable") || !strings.Contains(stderr.String(), "incomplete inventory: collection failed for brew") {
			t.Fatalf("unclear failure: %q", stderr.String())
		}
		if jsonOut {
			var rows []model.CLIEntry
			if err := json.Unmarshal(stdout.Bytes(), &rows); err != nil || len(rows) != 1 {
				t.Fatalf("invalid partial JSON: %q, %v", stdout.String(), err)
			}
		}
	}
}

func TestAllFailedJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	collectors := []collector.Collector{
		&fakeCollector{name: "brew", err: errors.New("failed")},
		&fakeCollector{name: "mise", err: errors.New("failed")},
	}
	if code := listTools(&stdout, &stderr, collectors, "", false, true, false); code != 1 {
		t.Fatalf("exit=%d", code)
	}
	if stdout.String() != "[]\n" || !strings.Contains(stderr.String(), "brew, mise") {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestEmptyInventoryIsSuccess(t *testing.T) {
	// Absent managers and installed managers with no tools both return nil, nil.
	var stdout, stderr bytes.Buffer
	if code := listTools(&stdout, &stderr, []collector.Collector{&fakeCollector{name: "uv"}}, "", false, true, false); code != 0 || stderr.Len() != 0 || stdout.String() != "[]\n" {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestSourceFilterDoesNotReportUnselectedFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	failed := &fakeCollector{name: "brew", err: errors.New("failed")}
	if code := listTools(&stdout, &stderr, []collector.Collector{failed, &fakeCollector{name: "uv"}}, "uv", false, true, false); code != 0 || failed.calls != 0 || stderr.Len() != 0 {
		t.Fatalf("exit=%d calls=%d stderr=%q", code, failed.calls, stderr.String())
	}
}

func TestCollectorPartialRowsPreserved(t *testing.T) {
	var stdout, stderr bytes.Buffer
	c := &fakeCollector{name: "manual", entries: []model.CLIEntry{{Tool: "tool", Version: "unknown", Source: "manual"}}, err: errors.New("version command failed")}
	if code := listTools(&stdout, &stderr, []collector.Collector{c}, "", false, true, false); code != 1 || !strings.Contains(stdout.String(), "unknown") {
		t.Fatalf("exit=%d stdout=%q", code, stdout.String())
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestOutputFailure(t *testing.T) {
	var stderr bytes.Buffer
	if code := listTools(brokenWriter{}, &stderr, nil, "", false, true, false); code != 1 || !strings.Contains(stderr.String(), "cliv: output:") {
		t.Fatalf("exit=%d stderr=%q", code, stderr.String())
	}
}
