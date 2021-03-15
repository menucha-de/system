/*Package sysinfo System Information
 *
 * Network configuration
 *
 * API info@menucha.deversion: 1.0.0
 * Contact: info@menucha.de
 */
package sysinfo

import (
	"github.com/menucha-de/utils"
)

// SysInfoRoutes all routes concerning sysinfo
var SysInfoRoutes = utils.Routes{
	utils.Route{
		Name:        "GetSysInfo",
		Method:      "GET",
		Pattern:     "/rest/sysinfo",
		HandlerFunc: GetSysInfo,
	},
	utils.Route{
		Name:        "SetState",
		Method:      "PUT",
		Pattern:     "/rest/system/state",
		HandlerFunc: State,
	},
}
