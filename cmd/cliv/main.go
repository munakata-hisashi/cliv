// clivのエントリポイントです。
// サブコマンドを持たず、MVPの中心価値である「一覧を見る」操作に寄せています。

package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/munakata-hisashi/cliv/internal/collector"
	"github.com/munakata-hisashi/cliv/internal/config"
	"github.com/munakata-hisashi/cliv/internal/model"
	"github.com/munakata-hisashi/cliv/internal/output"
)

const version = "0.1.0"

func main() {
	var source string
	var all bool
	var jsonOut bool
	var showVersion bool

	// MVPではサブコマンドを持たず、必要な操作だけをflagで受けます。
	flag.StringVar(&source, "source", "", "filter by source (brew, mise, npm, manual)")
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
	if source != "" && source != "brew" && source != "mise" && source != "npm" && source != "manual" {
		fmt.Fprintf(os.Stderr, "cliv: unknown source %q (brew, mise, npm, manual)\n", source)
		os.Exit(2)
	}

	// 設定で有効なcollectorだけを動かし、未導入のmanagerは静かに無視します。
	cfg := config.Load()
	collectors := enabledCollectors(cfg)
	var entries []model.CLIEntry
	for _, c := range collectors {
		if source != "" && c.Name() != source {
			continue
		}
		got, err := c.Collect(all)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", c.Name(), err)
			continue
		}
		entries = append(entries, got...)
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
		err = output.JSON(os.Stdout, entries)
	} else {
		err = output.Table(os.Stdout, entries)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
		if a.Command != b.Command {
			return a.Command < b.Command
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		return a.Version < b.Version
	})
}
