// 収集済みCLI情報を表示形式に変換します。
// collectorと出力を分け、JSON追加などの変更を局所化します。

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/munakata-hisashi/cliv/internal/model"
)

func JSON(w io.Writer, entries []model.CLIEntry) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func Table(w io.Writer, entries []model.CLIEntry) error {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	if _, err := fmt.Fprintln(tw, "COMMAND\tVERSION\tSOURCE"); err != nil {
		return err
	}
	for _, e := range entries {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Command, e.Version, e.Source); err != nil {
			return err
		}
	}
	return tw.Flush()
}
