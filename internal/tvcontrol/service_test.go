package tvcontrol

import (
	neturl "net/url"
	"testing"
)

func TestFilterDevicesPrefersSingleAirPlayMatch(t *testing.T) {
	devices := []Device{
		{Name: "Living Room TV", Protocol: ProtocolDLNA, Host: "tv-dlna.local"},
		{Name: "Living Room TV", Protocol: ProtocolAirPlay, Host: "tv-airplay.local"},
	}

	matches := filterDevices(devices, "living room", "")
	if len(matches) != 1 {
		t.Fatalf("expected 1 preferred match, got %d", len(matches))
	}
	if matches[0].Protocol != ProtocolAirPlay {
		t.Fatalf("expected airplay match, got %s", matches[0].Protocol)
	}
}

func TestNormalizeMediaURL(t *testing.T) {
	value, err := normalizeMediaURL("https://example.test/video.mp4")
	if err != nil {
		t.Fatalf("normalize media url: %v", err)
	}
	if value != "https://example.test/video.mp4" {
		t.Fatalf("unexpected media url %q", value)
	}
}

func TestBaseURLWrapsIPv6Hosts(t *testing.T) {
	value := baseURL("fe80::1234", 7000)
	if value != "http://[fe80::1234]:7000" {
		t.Fatalf("unexpected base url %q", value)
	}
}

func mustParseURL(t *testing.T, value string) *neturl.URL {
	t.Helper()
	parsed, err := neturl.Parse(value)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return parsed
}
