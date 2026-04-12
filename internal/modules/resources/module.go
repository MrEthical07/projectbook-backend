package resources

import (
	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/core/modulekit"
)

// Module serves resources routes.
type Module struct {
	runtime modulekit.Runtime
	handler *Handler
}

// New constructs resources module.
func New() *Module { return &Module{} }

var _ app.Module = (*Module)(nil)
var _ app.DependencyBinder = (*Module)(nil)

// Name returns module name.
func (m *Module) Name() string { return "resources" }

// BindDependencies wires runtime dependencies.
func (m *Module) BindDependencies(deps *app.Dependencies) {
	m.runtime = modulekit.New(deps)
	repo := NewRepo(m.runtime.RelationalStore(), m.runtime.DocumentStore())
	m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
}
