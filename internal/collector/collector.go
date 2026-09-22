// Collectorは各インストール元を同じ形で扱うための共通口です。
// sourceごとの差分をここで吸収し、後段の出力処理を単純に保ちます。

package collector

import "github.com/munakata-hisashi/cliv/internal/model"

type Collector interface {
	Name() string
	Collect(all bool) ([]model.CLIEntry, error)
}
