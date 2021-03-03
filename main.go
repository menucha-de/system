/*
 * App.Systemd
 *
 * API version: 0.0.1
 * Contact: support@peraMIC.io
 */

package main

import (
	"net/http"
	"net/rpc"

	"github.com/peramic/logging"

	"github.com/peramic/App.Systemd/datetime"
	"github.com/peramic/App.Systemd/journal"
	"github.com/peramic/App.Systemd/network"
	"github.com/peramic/App.Systemd/proxy"
	"github.com/peramic/App.Systemd/service"
	"github.com/peramic/App.Systemd/sysinfo"
	"github.com/peramic/App.Systemd/systemd"
)

var log *logging.Logger

func main() {
	log = logging.GetLogger("systemd")
	//log.SetLevel(logrus.DebugLevel)

	// systemd.AddRoutes(mount.MountRoutes)
	systemd.AddRoutes(logging.LogRoutes)
	systemd.AddRoutes(network.NetworkRoutes)
	systemd.AddRoutes(datetime.DateTimeRoutes)
	systemd.AddRoutes(proxy.ProxyRoutes)
	systemd.AddRoutes(sysinfo.SysInfoRoutes)
	systemd.AddRoutes(journal.JournalRoutes)
	router := systemd.NewRouter()

	var s = new(service.Service)

	rpcServer := rpc.NewServer()
	rpcServer.Register(s)

	router.Handle(rpc.DefaultRPCPath, rpcServer)

	log.Infof("Server started")

	log.Fatal(http.ListenAndServe("systemd:8080", router))
}
