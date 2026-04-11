package sidebar

import (
	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/core/modulekit"
	"github.com/MrEthical07/superapi/internal/modules/artifacts"
	"github.com/MrEthical07/superapi/internal/modules/pages"
)

// Module serves sidebar artifact routes.
type Module struct {
	runtime modulekit.Runtime
	handler *Handler
}

// New constructs sidebar module.
func New() *Module { return &Module{} }

var _ app.Module = (*Module)(nil)
var _ app.DependencyBinder = (*Module)(nil)

// Name returns module name.
func (m *Module) Name() string { return "sidebar" }

// BindDependencies wires runtime dependencies.
func (m *Module) BindDependencies(deps *app.Dependencies) {
	m.runtime = modulekit.New(deps)
	artifactsRepo := artifacts.NewRepo(m.runtime.RelationalStore())
	pagesRepo := pages.NewRepo(m.runtime.RelationalStore())
	pagesSvc := pages.NewService(m.runtime.RelationalStore(), pagesRepo)
	repo := NewRepo(m.runtime.RelationalStore(), artifactsRepo, pagesSvc)
	m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
}
