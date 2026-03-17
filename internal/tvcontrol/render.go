package tvcontrol

import (
	"fmt"
	"strings"
)

func RenderDiscoverText(result DiscoverResult) string {
	var builder strings.Builder
	if len(result.Devices) == 0 {
		builder.WriteString("no devices found\n")
	} else {
		for _, device := range result.Devices {
			fmt.Fprintf(&builder, "%-8s %-24s host=%s", device.Protocol, device.Name, device.Host)
			if device.Port > 0 {
				fmt.Fprintf(&builder, " port=%d", device.Port)
			}
			if device.Model != "" {
				fmt.Fprintf(&builder, " model=%s", device.Model)
			}
			if device.Manufacturer != "" {
				fmt.Fprintf(&builder, " maker=%s", device.Manufacturer)
			}
			builder.WriteByte('\n')
			if device.ControlURL != "" {
				fmt.Fprintf(&builder, "  control_url: %s\n", device.ControlURL)
			}
			if device.Location != "" {
				fmt.Fprintf(&builder, "  location: %s\n", device.Location)
			}
			if device.MAC != "" {
				fmt.Fprintf(&builder, "  mac: %s\n", device.MAC)
			}
		}
	}
	for _, warning := range result.Warnings {
		fmt.Fprintf(&builder, "warning: %s\n", warning)
	}
	return builder.String()
}

func RenderActionText(result ActionResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "operation: %s\n", result.Operation)
	fmt.Fprintf(&builder, "protocol: %s\n", result.Protocol)
	fmt.Fprintf(&builder, "target: %s\n", result.Target)
	fmt.Fprintf(&builder, "ok: %t\n", result.OK)
	if result.URL != "" {
		fmt.Fprintf(&builder, "url: %s\n", result.URL)
	}
	if result.HTTPStatus > 0 {
		fmt.Fprintf(&builder, "http_status: %d\n", result.HTTPStatus)
	}
	if result.Detail != "" {
		fmt.Fprintf(&builder, "detail: %s\n", result.Detail)
	}
	return builder.String()
}

func RenderProbeText(result ProbeResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "target: %s\n", result.Target)
	if result.Device != nil {
		fmt.Fprintf(&builder, "device: %s\n", result.Device.Name)
	}
	for _, endpoint := range result.Endpoints {
		fmt.Fprintf(&builder, "%s reachable=%t", endpoint.Name, endpoint.Reachable)
		if endpoint.HTTPStatus > 0 {
			fmt.Fprintf(&builder, " http_status=%d", endpoint.HTTPStatus)
		}
		fmt.Fprintf(&builder, " url=%s\n", endpoint.URL)
	}
	for _, hint := range result.Hints {
		fmt.Fprintf(&builder, "hint: %s\n", hint)
	}
	return builder.String()
}

func RenderPairText(result PairResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "protocol: %s\n", result.Protocol)
	fmt.Fprintf(&builder, "target: %s\n", result.Target)
	fmt.Fprintf(&builder, "ok: %t\n", result.OK)
	if result.Detail != "" {
		fmt.Fprintf(&builder, "detail: %s\n", result.Detail)
	}
	if result.Token != "" {
		fmt.Fprintf(&builder, "token: %s\n", result.Token)
	}
	return builder.String()
}

func RenderWakeText(result WakeResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "target: %s\n", result.Target)
	fmt.Fprintf(&builder, "mac: %s\n", result.MAC)
	fmt.Fprintf(&builder, "ok: %t\n", result.OK)
	fmt.Fprintf(&builder, "sent_bytes: %d\n", result.SentBytes)
	return builder.String()
}

func sortDevices(devices []Device) {
	less := func(left Device, right Device) bool {
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		if left.Protocol != right.Protocol {
			return left.Protocol < right.Protocol
		}
		return left.Host < right.Host
	}
	for i := 0; i < len(devices); i++ {
		for j := i + 1; j < len(devices); j++ {
			if less(devices[j], devices[i]) {
				devices[i], devices[j] = devices[j], devices[i]
			}
		}
	}
}
