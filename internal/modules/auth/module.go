package auth

import (
	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/core/modulekit"
)

// Module serves authentication routes backed by goAuth.
type Module struct {
	runtime modulekit.Runtime
	handler *Handler
}

// New constructs the auth module.
func New() *Module { return &Module{} }

var _ app.Module = (*Module)(nil)
var _ app.DependencyBinder = (*Module)(nil)

// Name returns module registry name.
func (m *Module) Name() string { return "auth" }

// BindDependencies wires runtime-backed service dependencies.
func (m *Module) BindDependencies(deps *app.Dependencies) {
	m.runtime = modulekit.New(deps)
	repo := NewRepo(m.runtime.RelationalStore())
	m.handler = NewHandler(NewService(m.runtime.AuthEngine(), repo))
}
