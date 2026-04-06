package mattercontrol

import (
	"context"
	"fmt"
	"sync"

	"github.com/grandcat/zeroconf"
)

func (service *Service) discoverService(ctx context.Context, serviceType string) ([]Device, error) {
	resolver, err := service.resolverFactory()
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	browseErr := make(chan error, 1)
	var (
		mu      sync.Mutex
		devices = make(map[string]Device)
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
				device := deviceFromEntry(serviceType, entry)
				key := device.Service + "|" + device.Instance + "|" + device.Host + fmt.Sprintf("|%d", device.Port)
				mu.Lock()
				devices[key] = device
				mu.Unlock()
			}
		}
	}()

	go func() {
		browseErr <- resolver.Browse(ctx, serviceType, "local.", entries)
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

func deviceFromEntry(serviceType string, entry *zeroconf.ServiceEntry) Device {
	addresses := entryAddresses(entry)
	txt := parseTXT(entry.Text)
	vendorID, productID := splitVendorProduct(txt["vp"])

	return Device{
		Name:                firstNonEmpty(entry.Instance, firstAddress(entry.HostName, addresses)),
		Instance:            firstNonEmpty(entry.Instance),
		Service:             serviceType,
		Host:                firstAddress(entry.HostName, addresses),
		Port:                entry.Port,
		Domain:              firstNonEmpty(entry.Domain),
		Addresses:           addresses,
		Discriminator:       txt["d"],
		VendorID:            vendorID,
		ProductID:           productID,
		DeviceType:          txt["dt"],
		CommissioningMode:   txt["cm"],
		PairingHint:         txt["ph"],
		PairingInstructions: txt["pi"],
		Metadata:            txt,
	}
}
