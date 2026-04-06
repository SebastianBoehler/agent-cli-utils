package mattercontrol

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func firstAddress(hostname string, addresses []string) string {
	if host := strings.TrimSuffix(strings.TrimSpace(hostname), "."); host != "" {
		return host
	}
	return firstNonEmpty(addresses...)
}

func entryAddresses(entry *zeroconf.ServiceEntry) []string {
	addresses := make([]string, 0, len(entry.AddrIPv4)+len(entry.AddrIPv6))
	for _, address := range entry.AddrIPv4 {
		addresses = append(addresses, address.String())
	}
	for _, address := range entry.AddrIPv6 {
		addresses = append(addresses, address.String())
	}
	return uniqueSorted(addresses)
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func splitVendorProduct(value string) (string, string) {
	vendor, product, ok := strings.Cut(strings.TrimSpace(value), "+")
	if !ok {
		return strings.TrimSpace(value), ""
	}
	return strings.TrimSpace(vendor), strings.TrimSpace(product)
}

func sortDevices(devices []Device) {
	sort.Slice(devices, func(i int, j int) bool {
		left := devices[i]
		right := devices[j]
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		if left.Service != right.Service {
			return left.Service < right.Service
		}
		if left.Host != right.Host {
			return left.Host < right.Host
		}
		return left.Port < right.Port
	})
}
