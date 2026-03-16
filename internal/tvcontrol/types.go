package tvcontrol

import "time"

const (
	ProtocolAirPlay = "airplay"
	ProtocolDLNA    = "dlna"
)

type Device struct {
	ID           string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name         string   `json:"name" yaml:"name"`
	Protocol     string   `json:"protocol" yaml:"protocol"`
	Host         string   `json:"host" yaml:"host"`
	Port         int      `json:"port,omitempty" yaml:"port,omitempty"`
	Addresses    []string `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
	Model        string   `json:"model,omitempty" yaml:"model,omitempty"`
	Location     string   `json:"location,omitempty" yaml:"location,omitempty"`
	ControlURL   string   `json:"control_url,omitempty" yaml:"control_url,omitempty"`
	Features     []string `json:"features,omitempty" yaml:"features,omitempty"`
	MAC          string   `json:"mac,omitempty" yaml:"mac,omitempty"`
}

type DiscoverOptions struct {
	Timeout time.Duration
}

type DiscoverResult struct {
	Devices  []Device `json:"devices" yaml:"devices"`
	Warnings []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

type PlayOptions struct {
	Device        string
	Host          string
	ControlURL    string
	Protocol      string
	URL           string
	StartPosition float64
	Timeout       time.Duration
}

type StopOptions struct {
	Device     string
	Host       string
	ControlURL string
	Protocol   string
	Timeout    time.Duration
}

type ActionResult struct {
	Operation  string  `json:"operation" yaml:"operation"`
	Protocol   string  `json:"protocol" yaml:"protocol"`
	Target     string  `json:"target" yaml:"target"`
	URL        string  `json:"url,omitempty" yaml:"url,omitempty"`
	OK         bool    `json:"ok" yaml:"ok"`
	HTTPStatus int     `json:"http_status,omitempty" yaml:"http_status,omitempty"`
	Detail     string  `json:"detail,omitempty" yaml:"detail,omitempty"`
	Device     *Device `json:"device,omitempty" yaml:"device,omitempty"`
}

type WakeOptions struct {
	MAC       string
	Broadcast string
}

type WakeResult struct {
	Target    string `json:"target" yaml:"target"`
	MAC       string `json:"mac" yaml:"mac"`
	SentBytes int    `json:"sent_bytes" yaml:"sent_bytes"`
	OK        bool   `json:"ok" yaml:"ok"`
}
