package collector

import "github.com/munakata-hisashi/cliv/internal/model"

type Collector interface {
	Name() string
	Collect(all bool) ([]model.CLIEntry, error)
}
