package youtubecontrol

import (
	"fmt"
	"strings"
)

func RenderDiscoverText(result DiscoverResult) string {
	var builder strings.Builder
	if len(result.Devices) == 0 {
		builder.WriteString("no youtube receivers found\n")
	}
	for _, device := range result.Devices {
		fmt.Fprintf(&builder, "%-24s host=%s state=%s", device.Name, device.Host, firstNonEmpty(device.State, "unknown"))
		if device.Version != "" {
			fmt.Fprintf(&builder, " version=%s", device.Version)
		}
		builder.WriteByte('\n')
		fmt.Fprintf(&builder, "  app_url: %s\n", device.AppURL)
		if device.RunURL != "" {
			fmt.Fprintf(&builder, "  run_url: %s\n", device.RunURL)
		}
	}
	for _, warning := range result.Warnings {
		fmt.Fprintf(&builder, "warning: %s\n", warning)
	}
	return builder.String()
}

func RenderStatusText(result StatusResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "name: %s\n", result.Device.Name)
	fmt.Fprintf(&builder, "host: %s\n", result.Device.Host)
	fmt.Fprintf(&builder, "state: %s\n", firstNonEmpty(result.Device.State, "unknown"))
	if result.Device.Version != "" {
		fmt.Fprintf(&builder, "version: %s\n", result.Device.Version)
	}
	fmt.Fprintf(&builder, "app_url: %s\n", result.Device.AppURL)
	if result.Device.RunURL != "" {
		fmt.Fprintf(&builder, "run_url: %s\n", result.Device.RunURL)
	}
	return builder.String()
}

func RenderActionText(result ActionResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "operation: %s\n", result.Operation)
	fmt.Fprintf(&builder, "target: %s\n", result.Target)
	fmt.Fprintf(&builder, "ok: %t\n", result.OK)
	if result.VideoID != "" {
		fmt.Fprintf(&builder, "video_id: %s\n", result.VideoID)
	}
	if result.HTTPStatus > 0 {
		fmt.Fprintf(&builder, "http_status: %d\n", result.HTTPStatus)
	}
	if result.Detail != "" {
		fmt.Fprintf(&builder, "detail: %s\n", result.Detail)
	}
	return builder.String()
}
