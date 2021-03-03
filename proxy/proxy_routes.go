package proxy

import (
	"github.com/peramic/utils"
)

//ProxyRoutes all routes concerning date and time
var ProxyRoutes = utils.Routes{
	utils.Route{
		Name:        "GetProxy",
		Method:      "GET",
		Pattern:     "/rest/proxy",
		HandlerFunc: getProxy,
	}, utils.Route{
		Name:        "SetProxy",
		Method:      "PUT",
		Pattern:     "/rest/proxy",
		HandlerFunc: setProxy,
	},
}
