package tvcontrol

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (service *Service) Probe(ctx context.Context, options ProbeOptions) (ProbeResult, error) {
	ctx, cancel := withTimeout(ctx, options.Timeout)
	defer cancel()

	device, err := service.resolveProbeTarget(ctx, options.Device, options.Host)
	if err != nil {
		return ProbeResult{}, err
	}

	result := ProbeResult{Target: device.Host, Device: &device}
	if device.Host == "" {
		return result, nil
	}

	add := func(name string, url string) {
		status, detail, reachable := service.httpProbe(ctx, url)
		result.Endpoints = append(result.Endpoints, EndpointProbe{
			Name:       name,
			URL:        url,
			Reachable:  reachable,
			HTTPStatus: status,
			Detail:     detail,
		})
	}

	add("airplay", baseURL(device.Host, 7000)+"/server-info")
	add("samsung-api", baseURL(device.Host, 8001)+"/api/v2/")
	add("dial-youtube", baseURL(device.Host, 8080)+"/ws/apps/YouTube")
	if device.ControlURL != "" {
		add("dlna-avtransport", device.ControlURL)
	}

	result.Hints = buildProbeHints(result)
	return result, nil
}

func (service *Service) resolveProbeTarget(ctx context.Context, selector string, host string) (Device, error) {
	if strings.TrimSpace(host) != "" {
		return Device{Name: firstNonEmpty(selector, host), Host: strings.TrimSpace(host), Port: 7000}, nil
	}
	if strings.TrimSpace(selector) == "" {
		return Device{}, fmt.Errorf("provide -device or -host")
	}
	discovered, err := service.Discover(ctx, DiscoverOptions{Timeout: 3 * time.Second})
	if err != nil {
		return Device{}, err
	}
	matches := filterDevices(discovered.Devices, selector, "")
	if len(matches) == 0 {
		return Device{}, fmt.Errorf("no device matched %q", selector)
	}
	return matches[0], nil
}

func (service *Service) httpProbe(ctx context.Context, url string) (int, string, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err.Error(), false
	}
	resp, err := service.client.Do(req)
	if err != nil {
		return 0, err.Error(), false
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	detail := strings.TrimSpace(string(payload))
	if detail == "" {
		detail = resp.Status
	}
	return resp.StatusCode, detail, true
}

func buildProbeHints(result ProbeResult) []string {
	hints := make([]string, 0, 4)
	for _, endpoint := range result.Endpoints {
		switch endpoint.Name {
		case "airplay":
			if endpoint.HTTPStatus == http.StatusForbidden {
				hints = append(hints, "AirPlay is reachable but requires pairing/authentication.")
			}
		case "samsung-api":
			if endpoint.HTTPStatus == http.StatusOK && strings.Contains(endpoint.Detail, "TokenAuthSupport") {
				hints = append(hints, "Samsung native remote API is available and likely requires token pairing.")
			}
		case "dial-youtube":
			if endpoint.HTTPStatus == http.StatusOK {
				hints = append(hints, "DIAL app control is available on this device.")
			}
		case "dlna-avtransport":
			if endpoint.Reachable {
				hints = append(hints, "DLNA AVTransport control endpoint is reachable.")
			}
		}
	}
	return uniqueSorted(hints)
}
