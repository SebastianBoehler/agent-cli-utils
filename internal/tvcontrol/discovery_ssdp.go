package tvcontrol

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ssdpAddress        = "239.255.255.250:1900"
	mediaRendererST    = "urn:schemas-upnp-org:device:MediaRenderer:1"
	avTransportService = "urn:schemas-upnp-org:service:AVTransport:1"
)

type ssdpDescription struct {
	URLBase string          `xml:"URLBase"`
	Device  ssdpDeviceEntry `xml:"device"`
}

type ssdpDeviceEntry struct {
	DeviceType   string            `xml:"deviceType"`
	FriendlyName string            `xml:"friendlyName"`
	Manufacturer string            `xml:"manufacturer"`
	ModelName    string            `xml:"modelName"`
	UDN          string            `xml:"UDN"`
	ServiceList  []ssdpService     `xml:"serviceList>service"`
	DeviceList   []ssdpDeviceEntry `xml:"deviceList>device"`
}

type ssdpService struct {
	ServiceType string `xml:"serviceType"`
	ControlURL  string `xml:"controlURL"`
}

func discoverDLNA(ctx context.Context, client *http.Client) ([]Device, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	request := strings.Join([]string{
		"M-SEARCH * HTTP/1.1",
		"HOST: " + ssdpAddress,
		`MAN: "ssdp:discover"`,
		"MX: 1",
		"ST: " + mediaRendererST,
		"",
		"",
	}, "\r\n")

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(3 * time.Second)
	}
	_ = conn.SetDeadline(deadline)

	addr, err := net.ResolveUDPAddr("udp4", ssdpAddress)
	if err != nil {
		return nil, err
	}
	if _, err := conn.WriteTo([]byte(request), addr); err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	devices := make([]Device, 0, 8)
	buffer := make([]byte, 64*1024)

	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			if isTimeout(err) {
				break
			}
			return nil, err
		}

		headers := parseSSDPHeaders(string(buffer[:n]))
		location := headers["LOCATION"]
		if location == "" {
			continue
		}
		if _, ok := seen[location]; ok {
			continue
		}
		seen[location] = struct{}{}

		device, ok := inspectDLNADevice(ctx, client, location, headers["USN"])
		if ok {
			devices = append(devices, device)
		}
	}

	sortDevices(devices)
	return devices, nil
}

func inspectDLNADevice(ctx context.Context, client *http.Client, location string, usn string) (Device, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return Device{}, false
	}
	resp, err := client.Do(req)
	if err != nil {
		return Device{}, false
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Device{}, false
	}

	var description ssdpDescription
	if err := xml.Unmarshal(payload, &description); err != nil {
		return Device{}, false
	}

	deviceEntry, service, ok := findAVTransport(description.Device)
	if !ok {
		return Device{}, false
	}

	locationURL, err := neturl.Parse(location)
	if err != nil {
		return Device{}, false
	}
	controlURL, err := resolveControlURL(locationURL, description.URLBase, service.ControlURL)
	if err != nil {
		return Device{}, false
	}

	port, _ := strconv.Atoi(locationURL.Port())
	return Device{
		ID:           firstNonEmpty(strings.TrimSpace(deviceEntry.UDN), strings.TrimSpace(usn), location),
		Name:         firstNonEmpty(deviceEntry.FriendlyName, locationURL.Hostname()),
		Protocol:     ProtocolDLNA,
		Host:         locationURL.Hostname(),
		Port:         port,
		Addresses:    uniqueSorted([]string{locationURL.Hostname()}),
		Manufacturer: strings.TrimSpace(deviceEntry.Manufacturer),
		Model:        strings.TrimSpace(deviceEntry.ModelName),
		Location:     location,
		ControlURL:   controlURL,
	}, true
}

func findAVTransport(device ssdpDeviceEntry) (ssdpDeviceEntry, ssdpService, bool) {
	for _, service := range device.ServiceList {
		if strings.TrimSpace(service.ServiceType) == avTransportService {
			return device, service, true
		}
	}
	for _, child := range device.DeviceList {
		if deviceEntry, service, ok := findAVTransport(child); ok {
			return deviceEntry, service, true
		}
	}
	return ssdpDeviceEntry{}, ssdpService{}, false
}

func resolveControlURL(locationURL *neturl.URL, urlBase string, control string) (string, error) {
	base := locationURL
	if strings.TrimSpace(urlBase) != "" {
		parsed, err := neturl.Parse(strings.TrimSpace(urlBase))
		if err != nil {
			return "", fmt.Errorf("parse urlBase %q: %w", urlBase, err)
		}
		base = parsed
	}
	controlURL, err := neturl.Parse(strings.TrimSpace(control))
	if err != nil {
		return "", fmt.Errorf("parse control URL %q: %w", control, err)
	}
	return base.ResolveReference(controlURL).String(), nil
}

func parseSSDPHeaders(payload string) map[string]string {
	headers := make(map[string]string)
	for _, line := range strings.Split(payload, "\r\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		headers[strings.ToUpper(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	return headers
}

func isTimeout(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}
