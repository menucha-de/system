package mount

import (
	"github.com/peramic/utils"
)

// MountRoutes lists the route for testing via REST
var MountRoutes = utils.Routes{
	utils.Route{
		Name:        "ProcessMount",
		Method:      "GET",
		Pattern:     "/rest/mounts/{name}",
		HandlerFunc: processMount,
	},
}
