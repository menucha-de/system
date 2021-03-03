package proxy

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/peramic/App.Systemd/service"
	"github.com/peramic/logging"
	"github.com/peramic/utils"
)

var log *logging.Logger = logging.GetLogger("systemd-proxy")

const unitFile string = "/etc/systemd/system.conf.d/10-proxy.conf"

func getProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	proxy := ProxyConfig{}

	unitOptions, err := service.ReadUnitFile(unitFile)

	for _, option := range unitOptions {
		if option.Section == "Manager" && option.Name == "DefaultEnvironment" {
			if strings.Contains(option.Value, "HTTP_PROXY") {
				option.Value = strings.ReplaceAll(option.Value, "HTTP_PROXY=", "")
				option.Value = strings.ReplaceAll(option.Value, "\"", "")
				proxy.HTTPProxy = option.Value
				log.Info(option.Value)
			}
			if strings.Contains(option.Value, "HTTPS_PROXY") {
				option.Value = strings.ReplaceAll(option.Value, "HTTPS_PROXY=", "")
				option.Value = strings.ReplaceAll(option.Value, "\"", "")
				proxy.HTTPSProxy = option.Value
			}
			if strings.Contains(option.Value, "NO_PROXY") {
				option.Value = strings.ReplaceAll(option.Value, "NO_PROXY=", "")
				option.Value = strings.ReplaceAll(option.Value, "\"", "")
				proxy.NoProxy = strings.ReplaceAll(option.Value, "mqtt,169.254.0.1,", "")
			}
		}
	}
	err = json.NewEncoder(w).Encode(proxy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//w.WriteHeader(http.StatusOK)
}

func setProxy(w http.ResponseWriter, r *http.Request) {
	var proxyConf ProxyConfig
	err := utils.DecodeJSONBody(w, r, &proxyConf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	arr := []string{
		"<http_proxy>", proxyConf.HTTPProxy,
		"<https_proxy>", proxyConf.HTTPSProxy,
		"<no_proxy>", proxyConf.NoProxy,
	}
	err = service.WriteUnitFile(unitFile, "proxy", arr...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = service.ReloadDaemon()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = service.RestartUnit("art.service", "fail")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
