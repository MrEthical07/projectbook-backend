package artifacts

import (
	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/core/modulekit"
)

// Module serves stories, journeys, problems, ideas, tasks, and feedback routes.
type Module struct {
	runtime modulekit.Runtime
	handler *Handler
}

// New constructs artifacts module.
func New() *Module { return &Module{} }

var _ app.Module = (*Module)(nil)
var _ app.DependencyBinder = (*Module)(nil)

// Name returns module name.
func (m *Module) Name() string { return "artifacts" }

// BindDependencies wires runtime dependencies.
func (m *Module) BindDependencies(deps *app.Dependencies) {
	m.runtime = modulekit.New(deps)
	repo := NewRepo(m.runtime.RelationalStore())
	m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
}
