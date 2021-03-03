/*Package sysinfo System Information
 *
 * Network configuration
 *
 * API version: 0.0.1
 * Contact: support@peraMIC.io
 */
package sysinfo

import (
	"github.com/peramic/utils"
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
