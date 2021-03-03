package datetime

type DateTimeConfig struct {
	// Link ID
	NTP bool `json:"ntp"`
	// Link Name
	NTPServer string `json:"ntpserver,omitempty"`
	// Link Type
	TimeZone string   `json:"timezone,omitempty"`
	Zones    []string `json:"zones"`
	Time     string   `json:"time"`
	Date     string   `json:"date"`
	LastSync string   `json:"lastsync"`
}
