package modules

import (
	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/modules/activity"
	"github.com/MrEthical07/superapi/internal/modules/artifacts"
	"github.com/MrEthical07/superapi/internal/modules/auth"
	"github.com/MrEthical07/superapi/internal/modules/calendar"
	"github.com/MrEthical07/superapi/internal/modules/health"
	"github.com/MrEthical07/superapi/internal/modules/home"
	"github.com/MrEthical07/superapi/internal/modules/pages"
	"github.com/MrEthical07/superapi/internal/modules/project"
	"github.com/MrEthical07/superapi/internal/modules/resources"
	"github.com/MrEthical07/superapi/internal/modules/sidebar"
	"github.com/MrEthical07/superapi/internal/modules/system"
	"github.com/MrEthical07/superapi/internal/modules/team"
	// MODULE_IMPORTS
)

// START HERE:
// - This registry controls which modules are loaded at runtime.
// - Module generators update MODULE_IMPORTS and MODULE_LIST markers.
// - Route details live inside each module's routes.go file.

// All returns the complete runtime module list in registration order.
func All() []app.Module {
	return []app.Module{
		auth.New(),
		home.New(),
		project.New(),
		artifacts.New(),
		resources.New(),
		pages.New(),
		calendar.New(),
		sidebar.New(),
		activity.New(),
		team.New(),
		health.New(),
		system.New(),
		// MODULE_LIST
	}
}
