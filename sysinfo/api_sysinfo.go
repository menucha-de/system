/*Package sysinfo Base system information
 * API info@menucha.deversion: 1.0.0
 * Contact: info@menucha.de
 */
package sysinfo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/menucha-de/logging"
)

var log *logging.Logger = logging.GetLogger("sysinfo")

// GetSysInfo Base system information
func GetSysInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	sysinfo := getSysInfo()
	err := json.NewEncoder(w).Encode(sysinfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//w.WriteHeader(http.StatusOK)
}

// Reboot system
func State(w http.ResponseWriter, r *http.Request) {
	var state StateAction
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	state = StateAction(body)

	result := setState(state)
	if !result {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
