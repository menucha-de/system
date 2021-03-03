package datetime

import (
	"github.com/peramic/utils"
)

//DateTimeRoutes all routes concerning date and time
var DateTimeRoutes = utils.Routes{
	utils.Route{
		Name:        "GetDateTimeInfo",
		Method:      "GET",
		Pattern:     "/rest/datetime/info",
		HandlerFunc: getDateTimeInfo,
	}, utils.Route{
		Name:        "SetNTP",
		Method:      "PUT",
		Pattern:     "/rest/datetime/ntp",
		HandlerFunc: setNTP,
	}, utils.Route{
		Name:        "SetTimeZone",
		Method:      "PUT",
		Pattern:     "/rest/datetime/timezone",
		HandlerFunc: setTimeZone,
	}, utils.Route{
		Name:        "GetDateInfo",
		Method:      "GET",
		Pattern:     "/rest/datetime/datetime",
		HandlerFunc: getDateInfo,
	}, utils.Route{
		Name:        "SetDateInfo",
		Method:      "PUT",
		Pattern:     "/rest/datetime/datetime",
		HandlerFunc: setDateTime,
	},
	utils.Route{
		Name:        "SetNTPServer",
		Method:      "PUT",
		Pattern:     "/rest/datetime/ntpserver",
		HandlerFunc: setNTPServer,
	},
	utils.Route{
		Name:        "DeleteNTPServer",
		Method:      "DELETE",
		Pattern:     "/rest/datetime/ntpserver",
		HandlerFunc: deleteNTPServer,
	},
}
