/*Package network Network configuration
 *
 * Network configuration
 *
 * API info@menucha.deversion: 1.0.0
 * Contact: info@menucha.de
 */
package network

import (
	"github.com/menucha-de/utils"
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
