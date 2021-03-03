/*
 * Network configuration
 *
 * API version: 0.0.1
 * Contact: support@peraMIC.io
 */
package systemd

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/peramic/utils"
)

var routes = utils.Routes{}

// AddRoutes adds new routes
func AddRoutes(newRoutes utils.Routes) {
	routes = append(routes, newRoutes...)
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
