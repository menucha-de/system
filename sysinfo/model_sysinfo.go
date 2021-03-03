/*Package sysinfo Base system information
 *
 * API version: 0.0.1
 * Contact: support@peraMIC.io
 */
package sysinfo

// SysInfo Base system information
type SysInfo struct {
	Hostname   string `json:"hostname,omitempty"`
	IPv4       string `json:"ipv4,omitempty"`
	IPv6       string `json:"ipv6,omitempty"`
	DeviceType string `json:"deviceType,omitempty"`
	Serial     string `json:"serial,omitempty"`
	Processor  string `json:"processor,omitempty"`
	MemTotal   uint64 `json:"memTotal,omitempty"`
	MemFree    uint64 `json:"memFree,omitempty"`
	SpaceTotal uint64 `json:"spaceTotal,omitempty"`
	SpaceFree  uint64 `json:"spaceFree,omitempty"`
	OS         string `json:"os,omitempty"`
	Version    string `json:"version,omitempty"`
	Kernel     string `json:"kernel,omitempty"`
}
