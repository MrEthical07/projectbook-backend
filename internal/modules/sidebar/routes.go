package sidebar

import (
	"github.com/MrEthical07/superapi/internal/core/httpx"
)

func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		m.BindDependencies(m.runtime.Dependencies())
	}

	return nil
}
