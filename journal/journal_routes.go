package journal

import (
	"github.com/peramic/utils"
)

//ProxyRoutes all routes concerning date and time
var JournalRoutes = utils.Routes{
	utils.Route{
		Name:        "GetJournal",
		Method:      "GET",
		Pattern:     "/rest/journal",
		HandlerFunc: getJournal,
	}, utils.Route{
		Name:        "GetJournalUnit",
		Method:      "GET",
		Pattern:     "/rest/journal/{unit}",
		HandlerFunc: getJournal,
	},
}
