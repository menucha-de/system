/*
 * system
 *
 * API version: 1.0.0
 * Contact: info@menucha.de
 */

package main

import (
	"net/http"
	"net/rpc"

	"github.com/menucha-de/logging"

	"github.com/menucha-de/system/datetime"
	"github.com/menucha-de/system/journal"
	"github.com/menucha-de/system/network"
	"github.com/menucha-de/system/proxy"
	"github.com/menucha-de/system/service"
	"github.com/menucha-de/system/sysinfo"
	"github.com/menucha-de/system/systemd"
)

var log *logging.Logger

func main() {
	log = logging.GetLogger("system")
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

	log.Fatal(http.ListenAndServe("system:8080", router))
}
