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

type SearchOptions struct {
	Query      string
	Language   string
	Region     string
	Duration   string
	Caption    string
	Order      string
	SafeSearch string
	MaxResults int
	Timeout    time.Duration
}

type SearchItem struct {
	VideoID      string `json:"video_id" yaml:"video_id"`
	Title        string `json:"title" yaml:"title"`
	ChannelTitle string `json:"channel_title,omitempty" yaml:"channel_title,omitempty"`
	PublishedAt  string `json:"published_at,omitempty" yaml:"published_at,omitempty"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	URL          string `json:"url" yaml:"url"`
}

type SearchResult struct {
	Query      string       `json:"query" yaml:"query"`
	Language   string       `json:"language,omitempty" yaml:"language,omitempty"`
	Region     string       `json:"region,omitempty" yaml:"region,omitempty"`
	Duration   string       `json:"duration,omitempty" yaml:"duration,omitempty"`
	Caption    string       `json:"caption,omitempty" yaml:"caption,omitempty"`
	Order      string       `json:"order,omitempty" yaml:"order,omitempty"`
	SafeSearch string       `json:"safe_search,omitempty" yaml:"safe_search,omitempty"`
	MaxResults int          `json:"max_results" yaml:"max_results"`
	Items      []SearchItem `json:"items" yaml:"items"`
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
