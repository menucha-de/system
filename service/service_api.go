package service

import (
	"encoding/gob"
	"net/http"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/gorilla/mux"
	"github.com/menucha-de/logging"
)

var log *logging.Logger = logging.GetLogger("service")

// CreateRequest ...
type CreateRequest struct {
	Ns          string
	ID          string
	Name        string
	Description string
}

// StateRequest ...
type StateRequest struct {
	Name    string
	Enabled bool
	Active  bool
}

// UnitRequest ...
type UnitRequest struct {
	Name string
}

// DeleteRequest ...
type DeleteRequest struct {
	Name string
}

// UpgradeRequest ...
type UpgradeRequest struct {
	Hostname     string
	MountOptions string
}

// Response ...
type Response int

// Service ...
type Service int

// CreateService ...
func (s *Service) CreateService(req CreateRequest, resp *Response) error {
	arr := []string{
		"<description>", req.Description,
		"<containername>", req.Name,
		"<namespace>", req.Ns,
	}
	err := WriteUnitFile("/etc/systemd/system/"+req.Name+".service", "service", arr...)
	return err
}

// SwitchServiceState enables/disables the given unit
func (s *Service) SwitchServiceState(req StateRequest, resp *Response) error {
	var err error
	name := req.Name + ".service"
	log.Info("Switch service state "+name, " ", req)

	if req.Enabled {
		err = EnableUnitFile(name)
	} else {
		err = DisableUnitFile(name)
	}
	if req.Active {
		err = RestartUnit(name, "fail")
	} else {
		err = StopUnit(name, "fail")
	}
	return err
}

// GetStatus returns status information of the given unit
func (s *Service) GetStatus(req UnitRequest, status *dbus.UnitStatus) error {
	var err error
	name := req.Name + ".service"
	*status, err = ListUnit(name)
	return err
}

// DeleteUnitFile deletes the given unit file
func (s *Service) DeleteUnitFile(req DeleteRequest, resp *Response) error {

	err := DeleteUnit(req.Name + ".service")
	return err
}

// Upgrade upgrades systemd and containerd
func (s *Service) Upgrade(req UpgradeRequest, resp *Response) error {
	return upgrade(req.Hostname, req.MountOptions)
}

func processService(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	actions, ok := r.URL.Query()["action"]
	if ok && len(actions) > 0 {
		for _, action := range actions {
			switch action {
			case "status":
				result, err := ListUnit(name)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				enc := gob.NewEncoder(w)
				err = enc.Encode(result)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

			case "disable":
				err := DisableUnitFile(name)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case "enable":
				err := EnableUnitFile(name)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case "start":
				err := StartUnit(name, "fail")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case "stop":
				err := StopUnit(name, "fail")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case "delete":
				err := DeleteUnit(name)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}
