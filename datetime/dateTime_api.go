package datetime

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/godbus/dbus/v5"
	"github.com/menucha-de/system/service"
	"github.com/menucha-de/logging"
)

var log *logging.Logger = logging.GetLogger("datetime")

func getDateInfo(w http.ResponseWriter, r *http.Request) {
	var config DateTimeConfig
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")

	}
	xx, err := getDateTime(conn)
	if err != nil {
		log.WithError(err).Error("Failed to get property Time")

	}
	config.Date = xx.Format("2006-01-02")
	config.Time = xx.Format("03:04:05")
	config.LastSync = getLastSync()
	err = json.NewEncoder(w).Encode(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}
func getDateTimeInfo(w http.ResponseWriter, r *http.Request) {
	var config DateTimeConfig
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")

	}
	state, err := conn.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").GetProperty("org.freedesktop.timedate1.NTP")
	if err != nil {
		log.WithError(err).Error("Failed to get property NTP")

	}
	config.NTP = state.Value().(bool)
	//if config.NTP {
	filename := "/etc/systemd/timesyncd.conf"
	unitOptions, err := service.ReadUnitFile(filename)
	if err != nil {
		log.Error("timesyncd file could not be opened")
	}
	for _, u := range unitOptions {
		if u.Section == "Time" && u.Name == "NTP" {
			config.NTPServer = u.Value
		}
	}
	//}
	zone, err := conn.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").GetProperty("org.freedesktop.timedate1.Timezone")
	if err != nil {
		log.WithError(err).Error("Failed to get property Timezone")

	}
	config.TimeZone = zone.Value().(string)
	config.Zones = getTimeZones()
	config.LastSync = getLastSync()
	xx, err := getDateTime(conn)
	if err != nil {
		log.WithError(err).Error("Failed to get property Time")

	}
	config.Date = xx.Format("2006-01-02")
	config.Time = xx.Format("03:04:05")

	err = json.NewEncoder(w).Encode(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}
func setNTP(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	n, err := b.ReadFrom(r.Body)
	if err != nil || n == 0 {
		http.Error(w, "Could not read NTP value", http.StatusBadRequest)
		return
	}
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")
		http.Error(w, "Failed to connect to SystemBus bus", http.StatusBadRequest)
		return
	}
	val, _ := strconv.ParseBool(b.String())
	xx := conn.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").Call("org.freedesktop.timedate1.SetNTP", 0, val, false)

	if xx.Err != nil {
		log.WithError(xx.Err).Error("Failed to set NTP")
		http.Error(w, "Failed to set NTP", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func setNTPServer(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	n, err := b.ReadFrom(r.Body)
	if err != nil || n == 0 {
		http.Error(w, "Could not read NTP Server value", http.StatusBadRequest)
		return
	}
	err = setServer(b.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
func deleteNTPServer(w http.ResponseWriter, r *http.Request) {

	err := setServer("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
func setTimeZone(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	n, err := b.ReadFrom(r.Body)
	if err != nil || n == 0 {
		http.Error(w, "Could not read timezone value", http.StatusBadRequest)
		return
	}
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")
		http.Error(w, "Failed to connect to SystemBus bus", http.StatusBadRequest)
		return
	}

	xx := conn.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").Call("org.freedesktop.timedate1.SetTimezone", 0, b.String(), false)

	if xx.Err != nil {
		log.WithError(xx.Err).Error("Failed to set timezone")
		http.Error(w, "Failed to set timezone", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func setDateTime(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	n, err := b.ReadFrom(r.Body)
	if err != nil || n == 0 {
		http.Error(w, "Could not read time and date value", http.StatusBadRequest)
		return
	}
	t1 := time.Now()
	t, _ := time.ParseInLocation("2006-01-02 03:04:05", b.String(), t1.Location())
	val := t.UTC().UnixNano() / 1000

	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")
		http.Error(w, "Failed to connect to SystemBus bus", http.StatusBadRequest)
		return
	}

	xx := conn.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").Call("org.freedesktop.timedate1.SetTime", 0, int64(val), false, false)

	if xx.Err != nil {
		log.WithError(xx.Err).Error("Failed to set time")
		http.Error(w, "Failed to set time", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getTimeZones() []string {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")

	}
	var s []string
	err = conn.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").Call("org.freedesktop.timedate1.ListTimezones", 0).Store(&s)
	if err != nil {
		return nil
	}
	return s
}

func getDateTime(conn *dbus.Conn) (time.Time, error) {
	t, err := conn.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").GetProperty("org.freedesktop.timedate1.TimeUSec")
	if err != nil {

		return time.Now(), err
	}

	xx := time.Unix(int64(t.Value().(uint64))/1000000, 0)

	return xx, nil
}
func getLastSync() string {

	filename := "/var/lib/systemd/timesync/clock"

	// get last modified time
	file, err := os.Stat(filename)
	if err != nil {
		filename := "/var/lib/systemd/clock"
		file, err = os.Stat(filename)
		if err != nil {
			return ""
		}
	}
	return file.ModTime().Format("2006-01-02 03:04:05 MST")
}
func setServer(val string) error {
	filename := "/etc/systemd/timesyncd.conf"
	unitOptions, err := service.ReadUnitFile(filename)
	if err != nil {
		log.WithError(err).Error("timesyncd file could not be opened")
		return errors.New("timesyncd file could not be opened")
	}
	found := false
	for _, u := range unitOptions {
		if u.Section == "Time" && u.Name == "NTP" {
			u.Value = val
			found = true
		}
	}
	if !found {
		x := unit.UnitOption{
			Name:    "NTP",
			Section: "Time",
			Value:   val,
		}
		unitOptions = append(unitOptions, &x)
	}
	reader := unit.Serialize(unitOptions)
	buffer, err := ioutil.ReadAll(reader)
	if err != nil {
		log.WithError(err).Error("Error while writing file")
		return errors.New("Error setting NTP Server")

	}
	if err := ioutil.WriteFile(filename, buffer, 0644); err != nil {
		log.WithError(err).Error("Error while writing file")
		return errors.New("Error setting NTP Server")

	}
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")
		return errors.New("Settings saved but not applied")
	}
	state, err := conn.Object("org.freedesktop.timedate1", "/org/freedesktop/timedate1").GetProperty("org.freedesktop.timedate1.NTP")
	if err != nil {
		log.WithError(err).Error("Failed to get property NTP")
		return errors.New("Settings saved but not applied")
	}
	ntp := state.Value().(bool)
	if ntp {
		srvErr := service.RestartUnit("systemd-timesyncd.service", "fail")
		if srvErr != nil {
			log.WithError(srvErr).Error("Could not restart service. Settings saved but not applied.")
			return errors.New("Could not restart service. Settings saved but not applied")

		}
	}
	return nil
}
