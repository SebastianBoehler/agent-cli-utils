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

func RenderSearchText(result SearchResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "query: %s\n", result.Query)
	if result.Language != "" {
		fmt.Fprintf(&builder, "language: %s\n", result.Language)
	}
	if result.Region != "" {
		fmt.Fprintf(&builder, "region: %s\n", result.Region)
	}
	if result.Duration != "" {
		fmt.Fprintf(&builder, "duration: %s\n", result.Duration)
	}
	if result.Caption != "" {
		fmt.Fprintf(&builder, "caption: %s\n", result.Caption)
	}
	if result.Order != "" {
		fmt.Fprintf(&builder, "order: %s\n", result.Order)
	}
	if result.SafeSearch != "" {
		fmt.Fprintf(&builder, "safe_search: %s\n", result.SafeSearch)
	}
	fmt.Fprintf(&builder, "max_results: %d\n", result.MaxResults)
	for index, item := range result.Items {
		fmt.Fprintf(&builder, "%d. %s\n", index+1, item.Title)
		fmt.Fprintf(&builder, "   video_id: %s\n", item.VideoID)
		if item.ChannelTitle != "" {
			fmt.Fprintf(&builder, "   channel: %s\n", item.ChannelTitle)
		}
		if item.PublishedAt != "" {
			fmt.Fprintf(&builder, "   published_at: %s\n", item.PublishedAt)
		}
		fmt.Fprintf(&builder, "   url: %s\n", item.URL)
	}
	if len(result.Items) == 0 {
		builder.WriteString("no results\n")
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
