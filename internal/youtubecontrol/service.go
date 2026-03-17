package youtubecontrol

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/tvcontrol"
)

type Service struct {
	client *http.Client
	tv     *tvcontrol.Service
}

func NewService(client *http.Client) *Service {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &Service{client: client, tv: tvcontrol.NewService(client)}
}

func (service *Service) Discover(ctx context.Context, options DiscoverOptions) (DiscoverResult, error) {
	devices, warnings, err := service.lookup(ctx, "", options.Timeout)
	if err != nil {
		return DiscoverResult{}, err
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Name != devices[j].Name {
			return devices[i].Name < devices[j].Name
		}
		return devices[i].Host < devices[j].Host
	})
	return DiscoverResult{Devices: devices, Warnings: warnings}, nil
}

func (service *Service) Status(ctx context.Context, options StatusOptions) (StatusResult, error) {
	device, err := service.resolve(ctx, options.Device, options.Host, options.Timeout)
	if err != nil {
		return StatusResult{}, err
	}
	return StatusResult{Device: device}, nil
}

func (service *Service) Play(ctx context.Context, options PlayOptions) (ActionResult, error) {
	device, err := service.resolve(ctx, options.Device, options.Host, options.Timeout)
	if err != nil {
		return ActionResult{}, err
	}
	videoID, err := normalizeVideo(options.Video)
	if err != nil {
		return ActionResult{}, err
	}
	result := playYouTubeDial(service.client, device, videoID, options.StartOffset)
	if !result.OK {
		return result, nil
	}
	return result, nil
}

func (service *Service) resolve(ctx context.Context, selector string, host string, timeout time.Duration) (Device, error) {
	if strings.TrimSpace(host) != "" {
		device, _, err := fetchYouTubeStatus(service.client, strings.TrimSpace(host))
		if err != nil {
			return Device{}, err
		}
		if device.State == "" && device.Version == "" {
			return Device{}, fmt.Errorf("youtube receiver not found on %s", host)
		}
		return device, nil
	}
	devices, _, err := service.lookup(ctx, selector, timeout)
	if err != nil {
		return Device{}, err
	}
	if len(devices) == 0 {
		return Device{}, fmt.Errorf("no youtube-capable receiver matched %q", selector)
	}
	if len(devices) > 1 {
		return Device{}, fmt.Errorf("multiple youtube-capable receivers matched %q; narrow it with -host", selector)
	}
	return devices[0], nil
}

func (service *Service) lookup(ctx context.Context, selector string, timeout time.Duration) ([]Device, []string, error) {
	discovered, err := service.tv.Discover(ctx, tvcontrol.DiscoverOptions{Timeout: timeout})
	if err != nil {
		return nil, nil, err
	}

	seen := map[string]tvcontrol.Device{}
	for _, device := range discovered.Devices {
		if strings.TrimSpace(device.Host) == "" {
			continue
		}
		if _, ok := seen[device.Host]; !ok {
			seen[device.Host] = device
		}
	}

	var (
		devices  []Device
		warnings []string
	)
	for host, source := range seen {
		device, status, err := fetchYouTubeStatus(service.client, host)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s probe failed: %v", host, err))
			continue
		}
		if status != http.StatusOK {
			continue
		}
		device.Name = source.Name
		device.Manufacturer = source.Manufacturer
		device.Model = source.Model
		if matchesSelector(device, selector) {
			devices = append(devices, device)
		}
	}
	sort.Strings(warnings)
	return devices, warnings, nil
}

func matchesSelector(device Device, selector string) bool {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return true
	}
	return strings.EqualFold(device.Name, selector) ||
		strings.EqualFold(device.Host, selector) ||
		strings.Contains(strings.ToLower(device.Name), strings.ToLower(selector)) ||
		strings.Contains(strings.ToLower(device.Host), strings.ToLower(selector))
}

func normalizeVideo(value string) (string, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "", fmt.Errorf("video id or url is required")
	}
	if !strings.Contains(raw, "://") {
		return raw, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse youtube url: %w", err)
	}
	switch {
	case strings.Contains(parsed.Host, "youtu.be"):
		id := strings.Trim(strings.TrimSpace(parsed.Path), "/")
		if id != "" {
			return id, nil
		}
	case strings.Contains(parsed.Host, "youtube.com"):
		if id := strings.TrimSpace(parsed.Query().Get("v")); id != "" {
			return id, nil
		}
		if parts := strings.Split(strings.Trim(parsed.Path, "/"), "/"); len(parts) >= 2 && parts[0] == "shorts" {
			return parts[1], nil
		}
	}
	return "", fmt.Errorf("unsupported youtube url %q", raw)
}
