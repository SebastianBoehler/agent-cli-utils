package tvcontrol

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/devicestore"
)

type Service struct {
	client *http.Client
}

func NewService(client *http.Client) *Service {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &Service{client: client}
}

func (service *Service) Discover(ctx context.Context, options DiscoverOptions) (DiscoverResult, error) {
	ctx, cancel := withTimeout(ctx, options.Timeout)
	defer cancel()

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		devices  []Device
		warnings []string
	)

	run := func(label string, fn func(context.Context) ([]Device, error)) {
		defer wg.Done()
		found, err := fn(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s discovery failed: %v", label, err))
			return
		}
		devices = append(devices, found...)
	}

	wg.Add(2)
	go run(ProtocolAirPlay, discoverAirPlay)
	go run(ProtocolDLNA, func(ctx context.Context) ([]Device, error) {
		return discoverDLNA(ctx, service.client)
	})
	wg.Wait()

	sortDevices(devices)
	return DiscoverResult{Devices: dedupeDevices(devices), Warnings: uniqueSorted(warnings)}, nil
}

func (service *Service) Play(ctx context.Context, options PlayOptions) (ActionResult, error) {
	mediaURL, err := normalizeMediaURL(options.URL)
	if err != nil {
		return ActionResult{}, err
	}

	ctx, cancel := withTimeout(ctx, options.Timeout)
	defer cancel()

	device, err := service.resolveDevice(ctx, options.Protocol, options.Device, options.Host, options.ControlURL)
	if err != nil {
		return ActionResult{}, err
	}

	switch device.Protocol {
	case ProtocolAirPlay:
		return service.playAirPlayAware(ctx, device, mediaURL, options.StartPosition)
	case ProtocolDLNA:
		return playDLNA(ctx, service.client, device, mediaURL)
	default:
		return ActionResult{}, fmt.Errorf("unsupported protocol %q", device.Protocol)
	}
}

func (service *Service) Stop(ctx context.Context, options StopOptions) (ActionResult, error) {
	ctx, cancel := withTimeout(ctx, options.Timeout)
	defer cancel()

	device, err := service.resolveDevice(ctx, options.Protocol, options.Device, options.Host, options.ControlURL)
	if err != nil {
		return ActionResult{}, err
	}

	switch device.Protocol {
	case ProtocolAirPlay:
		return stopAirPlay(ctx, service.client, device)
	case ProtocolDLNA:
		return stopDLNA(ctx, service.client, device)
	default:
		return ActionResult{}, fmt.Errorf("unsupported protocol %q", device.Protocol)
	}
}

func (service *Service) resolveDevice(ctx context.Context, protocol string, selector string, host string, controlURL string) (Device, error) {
	protocol = normalizeProtocol(protocol)
	if controlURL != "" {
		if protocol == ProtocolAirPlay {
			return Device{}, fmt.Errorf("-control-url is only valid for dlna targets")
		}
		return Device{
			Name:       firstNonEmpty(selector, controlURL),
			Protocol:   ProtocolDLNA,
			Host:       hostFromURL(controlURL),
			Location:   controlURL,
			ControlURL: controlURL,
		}, nil
	}

	if strings.TrimSpace(host) != "" {
		switch protocol {
		case "", ProtocolAirPlay:
			resolvedHost, port, err := normalizeHostPort(host, 7000)
			if err != nil {
				return Device{}, err
			}
			return Device{
				Name:     firstNonEmpty(selector, resolvedHost),
				Protocol: ProtocolAirPlay,
				Host:     resolvedHost,
				Port:     port,
				Location: baseURL(resolvedHost, port),
			}, nil
		case ProtocolDLNA:
			return Device{}, fmt.Errorf("direct dlna playback requires -control-url or a discovered -device name")
		default:
			return Device{}, fmt.Errorf("unsupported protocol %q", protocol)
		}
	}

	if strings.TrimSpace(selector) == "" {
		return Device{}, fmt.Errorf("provide -device, -host, or -control-url")
	}

	discovered, err := service.Discover(ctx, DiscoverOptions{Timeout: 3 * time.Second})
	if err != nil {
		return Device{}, err
	}

	matches := filterDevices(discovered.Devices, selector, protocol)
	if len(matches) == 0 {
		return Device{}, fmt.Errorf("no device matched %q", selector)
	}
	if len(matches) > 1 {
		return Device{}, fmt.Errorf("device %q matched multiple targets; narrow it with -protocol or -host", selector)
	}
	return matches[0], nil
}

func dedupeDevices(devices []Device) []Device {
	seen := make(map[string]struct{}, len(devices))
	out := make([]Device, 0, len(devices))
	for _, device := range devices {
		key := strings.Join([]string{device.Protocol, device.Host, device.ControlURL, device.Location}, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, device)
	}
	return out
}

func filterDevices(devices []Device, selector string, protocol string) []Device {
	selector = strings.TrimSpace(selector)
	matches := make([]Device, 0, 4)
	for _, device := range devices {
		if protocol != "" && device.Protocol != protocol {
			continue
		}
		if strings.EqualFold(device.Name, selector) || strings.EqualFold(device.Host, selector) || strings.EqualFold(device.ID, selector) {
			matches = append(matches, device)
		}
	}
	if len(matches) > 0 {
		return preferAirPlay(matches, protocol)
	}
	for _, device := range devices {
		if protocol != "" && device.Protocol != protocol {
			continue
		}
		if containsFold(device.Name, selector) || containsFold(device.Host, selector) {
			matches = append(matches, device)
		}
	}
	return preferAirPlay(matches, protocol)
}

func preferAirPlay(devices []Device, protocol string) []Device {
	if protocol != "" || len(devices) <= 1 {
		return devices
	}
	airplay := make([]Device, 0, len(devices))
	for _, device := range devices {
		if device.Protocol == ProtocolAirPlay {
			airplay = append(airplay, device)
		}
	}
	if len(airplay) == 1 {
		return airplay
	}
	return devices
}

func normalizeMediaURL(value string) (string, error) {
	parsed, err := neturl.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("parse media url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("media url must be http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("media url must include a host")
	}
	return parsed.String(), nil
}

func hostFromURL(value string) string {
	parsed, err := neturl.Parse(value)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func (service *Service) playAirPlayAware(ctx context.Context, device Device, mediaURL string, startPosition float64) (ActionResult, error) {
	credentials, err := devicestore.LoadAppleCredentials(device.Host)
	if err != nil {
		return ActionResult{}, err
	}
	if credentials != "" {
		return service.playAppleBridge(ctx, device.Host, mediaURL)
	}
	return playAirPlay(ctx, service.client, device, mediaURL, startPosition)
}
