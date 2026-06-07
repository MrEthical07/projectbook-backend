package home

import (
	"github.com/MrEthical07/projectbook-backend/internal/core/app"
	"github.com/MrEthical07/projectbook-backend/internal/core/modulekit"
)

// Module serves authenticated home routes.
type Module struct {
	runtime modulekit.Runtime
	handler *Handler
}

// New constructs the home module.
func New() *Module { return &Module{} }

var _ app.Module = (*Module)(nil)
var _ app.DependencyBinder = (*Module)(nil)

// Name returns module registry name.
func (m *Module) Name() string { return "home" }

// BindDependencies wires runtime-backed dependencies.
func (m *Module) BindDependencies(deps *app.Dependencies) {
	m.runtime = modulekit.New(deps)
	repo := NewRepo(m.runtime.RelationalStore())
	m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
}
