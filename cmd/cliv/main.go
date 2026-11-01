// clivのエントリポイントです。
// サブコマンドを持たず、MVPの中心価値である「一覧を見る」操作に寄せています。

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/term"

	"github.com/munakata-hisashi/cliv/internal/collector"
	"github.com/munakata-hisashi/cliv/internal/config"
	"github.com/munakata-hisashi/cliv/internal/model"
	"github.com/munakata-hisashi/cliv/internal/output"
)

// Overridden with -ldflags for tagged release builds.
var version = "dev"

func main() {
	var source string
	var color string
	var all bool
	var jsonOut bool
	var showVersion bool

	// MVPではサブコマンドを持たず、必要な操作だけをflagで受けます。
	flag.StringVar(&source, "source", "", "filter by source (brew, mise, npm, uv, cargo, go, manual)")
	flag.StringVar(&color, "color", "auto", "color output (auto, always, never)")
	flag.BoolVar(&all, "all", false, "include dependency packages when supported")
	flag.BoolVar(&jsonOut, "json", false, "output JSON")
	flag.BoolVar(&showVersion, "version", false, "print cliv version")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "cliv - list CLI tools installed by multiple package managers\n\nUsage:\n  cliv [options]\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Println("cliv", version)
		return
	}
	if flag.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "cliv: subcommands are not supported in MVP")
		os.Exit(2)
	}
	if source != "" && source != "brew" && source != "mise" && source != "npm" && source != "uv" && source != "cargo" && source != "go" && source != "manual" {
		fmt.Fprintf(os.Stderr, "cliv: unknown source %q (brew, mise, npm, uv, cargo, go, manual)\n", source)
		os.Exit(2)
	}

	if color != "auto" && color != "always" && color != "never" {
		fmt.Fprintf(os.Stderr, "cliv: unknown color mode %q (auto, always, never)\n", color)
		os.Exit(2)
	}

	// 設定で有効なcollectorだけを動かし、未導入のmanagerは静かに無視します。
	cfg := config.Load()
	useColor := colorEnabled(color, term.IsTerminal(int(os.Stdout.Fd())), os.Getenv("NO_COLOR"), os.Getenv("TERM"))
	os.Exit(listTools(os.Stdout, os.Stderr, enabledCollectors(cfg), source, all, jsonOut, useColor))
}

// Keep partial results, but make incomplete inventories detectable by scripts.
func listTools(stdout, stderr io.Writer, collectors []collector.Collector, source string, all, jsonOut, useColor bool) int {
	var entries []model.CLIEntry
	var failed []string
	for _, c := range collectors {
		if source != "" && c.Name() != source {
			continue
		}
		got, err := c.Collect(all)
		entries = append(entries, got...)
		if err != nil {
			fmt.Fprintf(stderr, "cliv: %s: %v\n", c.Name(), err)
			failed = append(failed, c.Name())
		}
	}
	if source != "" {
		entries = filterSource(entries, source)
	}
	if entries == nil {
		entries = []model.CLIEntry{}
	}
	// 出力は常に安定した順序にし、差分確認しやすくします。
	sortEntries(entries)

	var err error
	if jsonOut {
		err = output.JSON(stdout, entries)
	} else {
		err = output.ColoredTable(stdout, entries, useColor)
	}
	if len(failed) > 0 {
		fmt.Fprintf(stderr, "cliv: incomplete inventory: collection failed for %s\n", strings.Join(failed, ", "))
	}
	if err != nil {
		fmt.Fprintf(stderr, "cliv: output: %v\n", err)
		return 1
	}
	if len(failed) > 0 {
		return 1
	}
	return 0
}

func colorEnabled(mode string, terminal bool, noColor, termName string) bool {
	if mode == "always" {
		return true
	}
	if mode == "never" {
		return false
	}
	return terminal && noColor == "" && termName != "dumb"
}

func enabledCollectors(cfg config.Config) []collector.Collector {
	var out []collector.Collector
	if cfg.Collectors["brew"] {
		out = append(out, collector.Brew{})
	}
	if cfg.Collectors["mise"] {
		out = append(out, collector.Mise{})
	}
	if cfg.Collectors["npm"] {
		out = append(out, collector.NPM{})
	}
	if cfg.Collectors["uv"] {
		out = append(out, collector.UV{})
	}
	if cfg.Collectors["cargo"] {
		out = append(out, collector.Cargo{})
	}
	if cfg.Collectors["go"] {
		out = append(out, collector.Go{})
	}
	if len(cfg.Tools) > 0 {
		out = append(out, collector.Manual{Tools: cfg.Tools})
	}
	return out
}

func filterSource(entries []model.CLIEntry, source string) []model.CLIEntry {
	out := entries[:0]
	for _, e := range entries {
		if e.Source == source {
			out = append(out, e)
		}
	}
	return out
}

func sortEntries(entries []model.CLIEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Tool != b.Tool {
			return a.Tool < b.Tool
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		return a.Version < b.Version
	})
}
