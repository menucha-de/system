/*
 * Network configuration
 *
 * Network configuration
 *
 * API info@menucha.deversion: 1.0.0
 * Contact: info@menucha.de
 */
package network

type Mode int

const (
	_ = iota
	Disabled
	DHCP
	Static
	LinkLocalOnly
)

func (mode Mode) String() string {
	return [...]string{"disabled", "dhcp", "static", "linklocal"}[mode]
}

// Overall configuration
type InterfaceConfig struct {
	// Link ID
	ID int `json:"id,omitempty"`
	// Link Name
	Name string `json:"name,omitempty"`
	// Hardware Address (MAC)
	MAC string `json:"mac,omitempty"`
	// DNS Suffix
	DNSSuffix string `json:"dnsSuffix,omitempty"`
	// Hostname
	Hostname string `json:"hostname,omitempty"`
	// Link Type
	InterfaceType string `json:"interfaceType,omitempty"`
	// Ipv4Mode
	Ipv4Mode Mode `json:"ipv4Mode,omitempty"`
	// IPv4 Address
	Ipv4Address string `json:"ipv4Address,omitempty"`
	// Ipv4Gateway
	Ipv4Gateway string `json:"ipv4Gateway,omitempty"`
	// Ipv4Nameserver
	Ipv4Nameserver string `json:"ipv4Nameserver,omitempty"`
	// Ipv6Mode
	Ipv6Mode Mode `json:"ipv6Mode,omitempty"`
	// IPv6 LinkLocal
	Ipv6LL string `json:"ipv6LL,omitempty"`
	// IPv6 ULA
	Ipv6ULA string `json:"ipv6ULA,omitempty"`
	// Gateway
	Ipv6Gateway string `json:"ipv6Gateway,omitempty"`
	// Ipv6Nameserver
	Ipv6Nameserver string `json:"ipv6Nameserver,omitempty"`
}
