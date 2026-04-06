package mattercontrol

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/grandcat/zeroconf"
)

type fakeResolver struct {
	entriesByService map[string][]*zeroconf.ServiceEntry
	errorsByService  map[string]error
}

func (resolver *fakeResolver) Browse(ctx context.Context, service string, domain string, entries chan<- *zeroconf.ServiceEntry) error {
	if err := resolver.errorsByService[service]; err != nil {
		return err
	}
	for _, entry := range resolver.entriesByService[service] {
		select {
		case <-ctx.Done():
			return nil
		case entries <- entry:
		}
	}
	<-ctx.Done()
	return nil
}

func TestDeviceFromEntryParsesTXTMetadata(t *testing.T) {
	t.Parallel()

	entry := &zeroconf.ServiceEntry{
		ServiceRecord: zeroconf.ServiceRecord{
			Instance: "DD200C20D25AE5F7",
			Domain:   "local.",
		},
		HostName: "B75AFB458ECD.local.",
		Port:     11111,
		Text:     []string{"D=840", "VP=123+456", "PH=3", "DT=35", "PI=scan-code"},
		AddrIPv6: []net.IP{net.ParseIP("fd00::1")},
	}

	device := deviceFromEntry(ServiceCommissioning, entry)
	if device.Name != "DD200C20D25AE5F7" {
		t.Fatalf("device.Name = %q, want %q", device.Name, "DD200C20D25AE5F7")
	}
	if device.Service != ServiceCommissioning {
		t.Fatalf("device.Service = %q, want %q", device.Service, ServiceCommissioning)
	}
	if device.Host != "B75AFB458ECD.local" {
		t.Fatalf("device.Host = %q, want %q", device.Host, "B75AFB458ECD.local")
	}
	if device.Discriminator != "840" {
		t.Fatalf("device.Discriminator = %q, want %q", device.Discriminator, "840")
	}
	if device.VendorID != "123" || device.ProductID != "456" {
		t.Fatalf("device vendor/product = %q/%q, want %q/%q", device.VendorID, device.ProductID, "123", "456")
	}
	if device.PairingHint != "3" {
		t.Fatalf("device.PairingHint = %q, want %q", device.PairingHint, "3")
	}
	if device.DeviceType != "35" {
		t.Fatalf("device.DeviceType = %q, want %q", device.DeviceType, "35")
	}
	if device.Metadata["pi"] != "scan-code" {
		t.Fatalf("device.Metadata[pi] = %q, want %q", device.Metadata["pi"], "scan-code")
	}
}

func TestDiscoverAggregatesServicesAndWarnings(t *testing.T) {
	t.Parallel()

	service := newServiceWithResolverFactory(func() (browser, error) {
		return &fakeResolver{
			entriesByService: map[string][]*zeroconf.ServiceEntry{
				ServiceOperational: {
					{
						ServiceRecord: zeroconf.ServiceRecord{
							Instance: "2906C908D115D362-8FC7772401CD0696",
							Domain:   "local.",
						},
						HostName: "matter-node.local.",
						Port:     22222,
						AddrIPv4: []net.IP{net.ParseIP("192.168.1.10")},
					},
				},
				ServiceCommissioning: {
					{
						ServiceRecord: zeroconf.ServiceRecord{
							Instance: "DD200C20D25AE5F7",
							Domain:   "local.",
						},
						HostName: "setup-node.local.",
						Port:     11111,
						Text:     []string{"D=840", "VP=123+456"},
						AddrIPv4: []net.IP{net.ParseIP("192.168.1.11")},
					},
					{
						ServiceRecord: zeroconf.ServiceRecord{
							Instance: "DD200C20D25AE5F7",
							Domain:   "local.",
						},
						HostName: "setup-node.local.",
						Port:     11111,
						Text:     []string{"D=840", "VP=123+456"},
						AddrIPv4: []net.IP{net.ParseIP("192.168.1.11")},
					},
				},
			},
			errorsByService: map[string]error{
				ServiceExtendedSetup: context.DeadlineExceeded,
			},
		}, nil
	})

	result, err := service.Discover(context.Background(), DiscoverOptions{Timeout: 20 * time.Millisecond})
	if err != nil {
		t.Fatalf("Discover() unexpected error: %v", err)
	}
	if len(result.Devices) != 2 {
		t.Fatalf("len(result.Devices) = %d, want %d", len(result.Devices), 2)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("len(result.Warnings) = %d, want %d", len(result.Warnings), 1)
	}
	if result.Devices[0].Name != "2906C908D115D362-8FC7772401CD0696" {
		t.Fatalf("first device name = %q", result.Devices[0].Name)
	}
	if result.Devices[1].VendorID != "123" {
		t.Fatalf("second device vendor_id = %q, want %q", result.Devices[1].VendorID, "123")
	}
}
