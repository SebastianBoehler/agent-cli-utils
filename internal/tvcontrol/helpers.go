package tvcontrol

import (
	"context"
	"fmt"
	"net"
	neturl "net/url"
	"sort"
	"strings"
	"time"
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

func normalizeProtocol(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return ""
	case ProtocolAirPlay:
		return ProtocolAirPlay
	case ProtocolDLNA:
		return ProtocolDLNA
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func normalizeHostPort(value string, defaultPort int) (string, int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", 0, fmt.Errorf("host is required")
	}
	if strings.Contains(value, "://") {
		parsed, err := neturl.Parse(value)
		if err != nil {
			return "", 0, fmt.Errorf("parse host %q: %w", value, err)
		}
		host := parsed.Hostname()
		port := defaultPort
		if parsed.Port() != "" {
			parsedPort, err := net.LookupPort("tcp", parsed.Port())
			if err != nil {
				return "", 0, fmt.Errorf("parse port from %q: %w", value, err)
			}
			port = parsedPort
		}
		return host, port, nil
	}

	host := value
	port := defaultPort
	if strings.Contains(value, ":") {
		parsedHost, parsedPort, err := net.SplitHostPort(value)
		if err != nil {
			return "", 0, fmt.Errorf("parse host %q: %w", value, err)
		}
		portValue, err := net.LookupPort("tcp", parsedPort)
		if err != nil {
			return "", 0, fmt.Errorf("parse port from %q: %w", value, err)
		}
		host = parsedHost
		port = portValue
	}
	return host, port, nil
}

func baseURL(host string, port int) string {
	if port <= 0 {
		port = 80
	}
	return (&neturl.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, fmt.Sprintf("%d", port)),
	}).String()
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

func firstAddress(hostname string, addresses []string) string {
	if host := strings.TrimSuffix(strings.TrimSpace(hostname), "."); host != "" {
		return host
	}
	for _, address := range addresses {
		if strings.TrimSpace(address) != "" {
			return address
		}
	}
	return ""
}

func containsFold(haystack string, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}
