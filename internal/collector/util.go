// collector間で使う小さな補助関数群です。
// 外部コマンド実行やPATH確認をここに寄せ、各collectorの処理を読みやすくします。

package collector

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func commandPath(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}

func run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func executableFiles(dir string) []string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(ents))
	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}
		// ReadDir entries for symlinks report the link mode, not the target mode.
		info, err := os.Stat(filepath.Join(dir, ent.Name()))
		if err != nil {
			continue
		}
		if info.Mode().IsRegular() && info.Mode()&0111 != 0 {
			out = append(out, ent.Name())
		}
	}
	sort.Strings(out)
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return "unknown"
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

func expandHome(path string) string {
	if path == "~" {
		return homeDir()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir(), path[2:])
	}
	return path
}
