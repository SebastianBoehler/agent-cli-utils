package mattercontrol

import (
	"fmt"
	"strings"
)

func RenderDiscoverText(result DiscoverResult) string {
	var builder strings.Builder
	if len(result.Devices) == 0 {
		builder.WriteString("no matter services found\n")
	} else {
		for _, device := range result.Devices {
			fmt.Fprintf(&builder, "%-13s %-24s host=%s", device.Service, device.Name, device.Host)
			if device.Port > 0 {
				fmt.Fprintf(&builder, " port=%d", device.Port)
			}
			builder.WriteByte('\n')

			if len(device.Addresses) > 0 {
				fmt.Fprintf(&builder, "  addresses: %s\n", strings.Join(device.Addresses, ", "))
			}
			if device.Discriminator != "" {
				fmt.Fprintf(&builder, "  discriminator: %s\n", device.Discriminator)
			}
			if device.VendorID != "" || device.ProductID != "" {
				fmt.Fprintf(&builder, "  vendor_product: %s", device.VendorID)
				if device.ProductID != "" {
					fmt.Fprintf(&builder, "+%s", device.ProductID)
				}
				builder.WriteByte('\n')
			}
			if device.DeviceType != "" {
				fmt.Fprintf(&builder, "  device_type: %s\n", device.DeviceType)
			}
			if device.CommissioningMode != "" {
				fmt.Fprintf(&builder, "  commissioning_mode: %s\n", device.CommissioningMode)
			}
			if device.PairingHint != "" {
				fmt.Fprintf(&builder, "  pairing_hint: %s\n", device.PairingHint)
			}
			if device.PairingInstructions != "" {
				fmt.Fprintf(&builder, "  pairing_instructions: %s\n", device.PairingInstructions)
			}
		}
	}
	for _, warning := range result.Warnings {
		fmt.Fprintf(&builder, "warning: %s\n", warning)
	}
	return builder.String()
}
