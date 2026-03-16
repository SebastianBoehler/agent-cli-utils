package tvcontrol

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/grandcat/zeroconf"
)

func discoverAirPlay(ctx context.Context) ([]Device, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	browseErr := make(chan error, 1)
	var (
		mu      sync.Mutex
		devices = map[string]Device{}
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-entries:
				if !ok || entry == nil {
					return
				}
				device := airPlayDeviceFromEntry(entry)
				key := device.Protocol + "|" + device.Host + "|" + strconv.Itoa(device.Port)
				mu.Lock()
				devices[key] = device
				mu.Unlock()
			}
		}
	}()

	go func() {
		browseErr <- resolver.Browse(ctx, "_airplay._tcp", "local.", entries)
	}()

	select {
	case err := <-browseErr:
		if err != nil {
			return nil, err
		}
		<-ctx.Done()
	case <-ctx.Done():
	}

	mu.Lock()
	defer mu.Unlock()

	out := make([]Device, 0, len(devices))
	for _, device := range devices {
		out = append(out, device)
	}
	sortDevices(out)
	return out, nil
}

func airPlayDeviceFromEntry(entry *zeroconf.ServiceEntry) Device {
	addresses := make([]string, 0, len(entry.AddrIPv4)+len(entry.AddrIPv6))
	for _, addr := range entry.AddrIPv4 {
		addresses = append(addresses, addr.String())
	}
	for _, addr := range entry.AddrIPv6 {
		addresses = append(addresses, addr.String())
	}

	txt := parseTXT(entry.Text)
	host := firstAddress(entry.HostName, addresses)
	model := txt["model"]
	features := splitFeatureList(txt["features"])
	mac := normalizeMAC(txt["deviceid"])

	return Device{
		ID:           firstNonEmpty(mac, entry.Instance, host),
		Name:         firstNonEmpty(entry.Instance, host),
		Protocol:     ProtocolAirPlay,
		Host:         host,
		Port:         entry.Port,
		Addresses:    uniqueSorted(addresses),
		Manufacturer: "Apple-compatible",
		Model:        model,
		Location:     baseURL(host, entry.Port),
		Features:     features,
		MAC:          mac,
	}
}

func parseTXT(values []string) map[string]string {
	out := make(map[string]string, len(values))
	for _, value := range values {
		key, raw, ok := strings.Cut(value, "=")
		if !ok {
			continue
		}
		out[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(raw)
	}
	return out
}

func splitFeatureList(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	})
	return uniqueSorted(fields)
}

func normalizeMAC(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", ":")
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
