/*
 * Network configuration
 *
 * API info@menucha.deversion: 1.0.0
 * Contact: info@menucha.de
 */
package systemd

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/menucha-de/utils"
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
