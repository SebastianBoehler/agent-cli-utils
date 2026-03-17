package youtubecontrol

import "time"

type Device struct {
	Name         string `json:"name" yaml:"name"`
	Host         string `json:"host" yaml:"host"`
	Manufacturer string `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty" yaml:"model,omitempty"`
	AppURL       string `json:"app_url" yaml:"app_url"`
	RunURL       string `json:"run_url,omitempty" yaml:"run_url,omitempty"`
	State        string `json:"state,omitempty" yaml:"state,omitempty"`
	Version      string `json:"version,omitempty" yaml:"version,omitempty"`
	AllowStop    bool   `json:"allow_stop,omitempty" yaml:"allow_stop,omitempty"`
}

type DiscoverOptions struct {
	Timeout time.Duration
}

type DiscoverResult struct {
	Devices  []Device `json:"devices" yaml:"devices"`
	Warnings []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

type StatusOptions struct {
	Device  string
	Host    string
	Timeout time.Duration
}

type StatusResult struct {
	Device Device `json:"device" yaml:"device"`
}

type PlayOptions struct {
	Device      string
	Host        string
	Video       string
	StartOffset string
	Timeout     time.Duration
}

type ActionResult struct {
	Operation  string `json:"operation" yaml:"operation"`
	Target     string `json:"target" yaml:"target"`
	VideoID    string `json:"video_id,omitempty" yaml:"video_id,omitempty"`
	OK         bool   `json:"ok" yaml:"ok"`
	HTTPStatus int    `json:"http_status,omitempty" yaml:"http_status,omitempty"`
	Detail     string `json:"detail,omitempty" yaml:"detail,omitempty"`
}
