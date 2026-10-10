// 収集済みCLI情報を表示形式に変換します。
// collectorと出力を分け、JSON追加などの変更を局所化します。

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/munakata-hisashi/cliv/internal/model"
)

func JSON(w io.Writer, entries []model.CLIEntry) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func Table(w io.Writer, entries []model.CLIEntry) error {
	// Keep the caller's ordering (used by JSON) intact.
	grouped := append([]model.CLIEntry(nil), entries...)
	order := map[string]int{"brew": 0, "mise": 1, "npm": 2, "manual": 3}
	sort.SliceStable(grouped, func(i, j int) bool {
		a, b := grouped[i], grouped[j]
		ai, aKnown := order[a.Source]
		bi, bKnown := order[b.Source]
		if aKnown != bKnown {
			return aKnown
		}
		if aKnown && ai != bi {
			return ai < bi
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		if a.Tool != b.Tool {
			return a.Tool < b.Tool
		}
		return a.Version < b.Version
	})

	for start := 0; start < len(grouped); {
		end := start + 1
		for end < len(grouped) && grouped[end].Source == grouped[start].Source {
			end++
		}
		if start > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, grouped[start].Source); err != nil {
			return err
		}
		// A separate writer keeps column widths local to each source.
		tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
		if _, err := fmt.Fprintln(tw, "TOOL\tVERSION"); err != nil {
			return err
		}
		for _, e := range grouped[start:end] {
			if _, err := fmt.Fprintf(tw, "%s\t%s\n", e.Tool, e.Version); err != nil {
				return err
			}
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		start = end
	}
	return nil
}
