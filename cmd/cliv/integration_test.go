package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLICollectionExitCodes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake manager uses a POSIX shell; collector unit tests run on all platforms")
	}
	bin := filepath.Join(t.TempDir(), "cliv")
	if output, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %s, %v", output, err)
	}
	for _, c := range []struct {
		name, script string
		wantCode     int
		wantError    bool
	}{
		{"not-installed", "", 0, false},
		{"empty", "#!/bin/sh\nexit 0\n", 0, false},
		{"command-failure", "#!/bin/sh\nprintf 'permission denied\\n' >&2\nexit 1\n", 1, true},
		{"invalid-output", "#!/bin/sh\nprintf 'unexpected output\\n'\n", 1, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if c.script != "" {
				if err := os.WriteFile(filepath.Join(dir, "uv"), []byte(c.script), 0755); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(bin, "--source", "uv", "--json")
			for _, item := range os.Environ() {
				if !strings.HasPrefix(item, "PATH=") && !strings.HasPrefix(item, "HOME=") {
					cmd.Env = append(cmd.Env, item)
				}
			}
			cmd.Env = append(cmd.Env, "PATH="+dir, "HOME="+dir)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			code := 0
			if err != nil {
				if exit, ok := err.(*exec.ExitError); ok {
					code = exit.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			if c.name == "command-failure" && !strings.Contains(stderr.String(), "permission denied") {
				t.Fatalf("lost manager error: %q", stderr.String())
			}
			if code != c.wantCode || stdout.String() != "[]\n" || strings.Contains(stderr.String(), "incomplete inventory") != c.wantError {
				t.Fatalf("exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}
