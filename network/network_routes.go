/*Package network Network configuration
 *
 * Network configuration
 *
 * API version: 0.0.1
 * Contact: support@peraMIC.io
 */
package network

import (
	"github.com/peramic/utils"
)

// NetworkRoutes all routes concerning networking
var NetworkRoutes = utils.Routes{
	utils.Route{
		Name:        "GetNetworkInfo",
		Method:      "GET",
		Pattern:     "/rest/network/interfaces",
		HandlerFunc: GetNetworkInfo,
	},
	utils.Route{
		Name:        "GetConfig",
		Method:      "GET",
		Pattern:     "/rest/network/interfaces/{name}",
		HandlerFunc: GetInterface,
	},
	utils.Route{
		Name:        "SetConfig",
		Method:      "PUT",
		Pattern:     "/rest/network/interfaces/{name}",
		HandlerFunc: SetInterface,
	},
}
