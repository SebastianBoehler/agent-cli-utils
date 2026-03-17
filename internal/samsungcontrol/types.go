package samsungcontrol

import "time"

const Protocol = "samsung"

type PairOptions struct {
	Host    string
	Name    string
	Timeout time.Duration
}

type PairResult struct {
	Protocol string `json:"protocol" yaml:"protocol"`
	Target   string `json:"target" yaml:"target"`
	OK       bool   `json:"ok" yaml:"ok"`
	Detail   string `json:"detail,omitempty" yaml:"detail,omitempty"`
	Token    string `json:"token,omitempty" yaml:"token,omitempty"`
}

type RemoteOptions struct {
	Host    string
	Key     string
	Timeout time.Duration
}

type LaunchOptions struct {
	Host    string
	AppID   string
	AppType string
	MetaTag string
	Timeout time.Duration
}

type ActionResult struct {
	Operation string `json:"operation" yaml:"operation"`
	Protocol  string `json:"protocol" yaml:"protocol"`
	Target    string `json:"target" yaml:"target"`
	OK        bool   `json:"ok" yaml:"ok"`
	Detail    string `json:"detail,omitempty" yaml:"detail,omitempty"`
}
