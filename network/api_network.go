/*
 * Network configuration
 *
 * Network configuration
 *
 * API version: 0.0.1
 * Contact: support@peraMIC.io
 */
package network

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/peramic/logging"
	"github.com/peramic/utils"
)

var log *logging.Logger = logging.GetLogger("systemd-network")

func GetNetworkInfo(w http.ResponseWriter, r *http.Request) {

	info, err := getInformation()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	//w.WriteHeader(http.StatusOK)
}

func GetInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	intf, err := getInterface(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(intf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func SetInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var config *InterfaceConfig
	err := utils.DecodeJSONBody(w, r, &config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setInterface(name, *config)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
