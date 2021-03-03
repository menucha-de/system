/*Package sysinfo Base system information
 * API version: 0.0.1
 * Contact: support@peraMIC.io
 */
package sysinfo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/peramic/logging"
)

var log *logging.Logger = logging.GetLogger("systemd-sysinfo")

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
