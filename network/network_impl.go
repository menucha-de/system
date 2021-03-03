package network

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"

	"github.com/peramic/App.Systemd/service"

	"github.com/peramic/utils"

	"github.com/vishvananda/netlink"
)

func getLinkInfo(lnk netlink.Link) *InterfaceConfig {
	attrs := lnk.Attrs()

	// Try to read information for a configured link
	linkInfo := getNetworkdInfo(attrs.Index)

	result := InterfaceConfig{}
	result.Name = attrs.Name
	result.ID = attrs.Index
	result.MAC = attrs.HardwareAddr.String()
	result.InterfaceType = getInterfaceType(attrs)
	result.Ipv6Gateway = getGateway(lnk, true)
	result.Ipv4Gateway = getGateway(lnk, false)
	getIPAddresses(lnk, &result)
	result.Ipv4Mode = getIpv4Mode(linkInfo, &result)
	result.Ipv6Mode = getIpv6Mode(linkInfo, &result)
	getNameservers(linkInfo, &result)
	result.Hostname = getHostname()
	return &result
}

func getNameservers(linkInfo map[string]string, config *InterfaceConfig) {
	dnss := linkInfo["DNS"]
	dnssArr := strings.Split(dnss, " ")
	for _, dns := range dnssArr {
		if strings.Contains(dns, ":") {
			config.Ipv6Nameserver = dns
		} else {
			config.Ipv4Nameserver = dns
		}
	}
}

func getHostname() string {
	result, _ := os.Hostname()
	return result
}

func getIPAddresses(lnk netlink.Link, config *InterfaceConfig) {
	addrs, err := netlink.AddrList(lnk, 0)
	if err != nil {
		log.Error("Could not read addresses")
		return
	}

	for _, addr := range addrs {
		if addr.IP.DefaultMask() != nil {
			config.Ipv4Address = addr.IPNet.String()
		} else {
			if addr.IP.IsLinkLocalUnicast() {
				config.Ipv6LL = addr.IPNet.String()
			} else {
				config.Ipv6ULA = addr.IPNet.String()
			}
		}
	}
}

func getIpv4Mode(linkInfo map[string]string, config *InterfaceConfig) Mode {
	dhcpAddress := linkInfo["DHCP4_ADDRESS"]
	net := strings.Split(config.Ipv4Address, "/")
	var address string
	if len(net) > 0 {
		address = net[0]
		if address == dhcpAddress {
			return DHCP
		}
	}
	return Static
}

func getIpv6Mode(linkInfo map[string]string, config *InterfaceConfig) Mode {
	var result Mode = Disabled
	if len(config.Ipv6LL) > 0 {
		result = LinkLocalOnly
	}

	if len(config.Ipv6ULA) > 0 {
		result = Static
	}
	return result
}

func readNetworkFile(filename string) []*unit.UnitOption {
	if filename != "" {
		unitOptions, err := service.ReadUnitFile(filename)
		if err != nil {
			log.Error("Network file could not be opened")
			return make([]*unit.UnitOption, 0)
		}
		return unitOptions
	}
	return make([]*unit.UnitOption, 0)
}

func getGateway(lnk netlink.Link, ipv6 bool) string {
	routes, err := netlink.RouteList(lnk, 0)
	if err != nil {
		log.Error("Could not read route list", err)
		return ""
	}

	if len(routes) > 0 {
		for _, route := range routes {
			if route.Dst == nil && route.Gw != nil && route.Gw.DefaultMask() != nil {
				// IPv6
				if ipv6 && route.Gw.DefaultMask() == nil {
					return route.Gw.String()
				}
				//IPv4
				if !ipv6 && route.Gw.DefaultMask() != nil {
					return route.Gw.String()
				}
			}
		}
	}
	return ""
}

func getNetworkdInfo(index int) map[string]string {
	networkdInfo, err := utils.ReadMap(fmt.Sprint("/run/systemd/netif/links/", index))
	if err != nil {
		return nil
	}
	return networkdInfo
}

func getInterfaceType(attrs *netlink.LinkAttrs) string {
	devInfo, err := utils.ReadMap(fmt.Sprint("/sys/class/net/", attrs.Name, "/uevent"))
	if err != nil {
		return attrs.EncapType
	}
	if result := devInfo["DEVTYPE"]; len(result) > 0 {
		return result
	}
	return attrs.EncapType
}

func findDHCPModeAndSetIPMode(current *InterfaceConfig, config *InterfaceConfig) (string, error) {
	// current, err := getInterface(config.Name)
	// if err != nil {
	// 	return "", err
	// }

	if config.Ipv4Mode == 0 {
		config.Ipv4Mode = current.Ipv4Mode
	}
	if config.Ipv6Mode == 0 {
		config.Ipv6Mode = current.Ipv6Mode
	}

	var dhcp string
	if config.Ipv4Mode == DHCP && config.Ipv6Mode == DHCP {
		dhcp = "yes"
	} else if config.Ipv4Mode == DHCP {
		dhcp = "ipv4"
	} else if config.Ipv6Mode == DHCP {
		dhcp = "ipv6"
	}
	return dhcp, nil
}

func setLinkConfig(config InterfaceConfig) error {
	current, err := getInterface(config.Name)
	if err != nil {
		return err
	}

	if len(config.Name) == 0 {
		error := "Interface name must be specified"
		log.Error(error)
		return errors.New(error)
	}

	dhcp, err := findDHCPModeAndSetIPMode(current, &config)
	if err != nil {
		return err
	}

	filename := fmt.Sprint("/etc/systemd/network/00-", config.Name, ".network")
	service.WriteUnitFile(filename, "linkmatch", "<linkname>", config.Name, "<dhcp>", dhcp, "<macaddress>", current.MAC)

	path := fmt.Sprint("/etc/systemd/network/00-", config.Name, ".network.d/")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	if config.Ipv4Mode == Static {
		params := []string{
			"<address>", config.Ipv4Address,
			"<gateway>", config.Ipv4Gateway,
			"<dns>", config.Ipv4Nameserver,
		}
		service.WriteUnitFile(path+"v4.conf", "static4", params...)
	} else {
		os.Remove(path + "v4.conf")
	}

	if config.Ipv6Mode == Static {
		params := []string{
			"<address>", config.Ipv6ULA,
			"<gateway>", config.Ipv6Gateway,
			"<dns>", config.Ipv6Nameserver,
		}
		service.WriteUnitFile(path+"v6.conf", "static6", params...)
	} else {
		os.Remove(path + "v6.conf")
	}
	return nil
}

func getInterface(name string) (*InterfaceConfig, error) {
	lnk, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}
	return getLinkInfo(lnk), nil
}

func setInterface(name string, config InterfaceConfig) error {
	if name != config.Name {
		msg := "Given name and name in config do not match!"
		log.Error(msg)
		return errors.New(msg)
	}
	err := setLinkConfig(config)
	if err == nil {
		srvErr := service.RestartUnit("systemd-networkd.service", "fail")
		if srvErr != nil {
			log.Error("Could not restart service. Settings saved but not applied.", srvErr)
		}
	}
	return nil
}

func getInformation() ([]*InterfaceConfig, error) {
	list, err := netlink.LinkList()

	if err != nil {
		return nil, err
	}
	infos := []*InterfaceConfig{}
	for _, lnk := range list {
		name := lnk.Attrs().Name
		if strings.HasPrefix(name, "en") ||
			strings.HasPrefix(name, "eth") ||
			strings.HasPrefix(name, "wl") ||
			strings.HasPrefix(name, "tun") {
			infos = append(infos, getLinkInfo(lnk))
		}
	}
	return infos, err
}
