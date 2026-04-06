package mattercontrol

import "time"

const (
	ServiceOperational   = "_matter._tcp"
	ServiceCommissioning = "_matterc._udp"
	ServiceExtendedSetup = "_matterd._udp"
)

type Device struct {
	Name                string            `json:"name" yaml:"name"`
	Instance            string            `json:"instance,omitempty" yaml:"instance,omitempty"`
	Service             string            `json:"service" yaml:"service"`
	Host                string            `json:"host" yaml:"host"`
	Port                int               `json:"port,omitempty" yaml:"port,omitempty"`
	Domain              string            `json:"domain,omitempty" yaml:"domain,omitempty"`
	Addresses           []string          `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Discriminator       string            `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	VendorID            string            `json:"vendor_id,omitempty" yaml:"vendor_id,omitempty"`
	ProductID           string            `json:"product_id,omitempty" yaml:"product_id,omitempty"`
	DeviceType          string            `json:"device_type,omitempty" yaml:"device_type,omitempty"`
	CommissioningMode   string            `json:"commissioning_mode,omitempty" yaml:"commissioning_mode,omitempty"`
	PairingHint         string            `json:"pairing_hint,omitempty" yaml:"pairing_hint,omitempty"`
	PairingInstructions string            `json:"pairing_instructions,omitempty" yaml:"pairing_instructions,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type DiscoverOptions struct {
	Timeout time.Duration
}

type DiscoverResult struct {
	Devices  []Device `json:"devices" yaml:"devices"`
	Warnings []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}
