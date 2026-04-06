package mattercontrol

import (
	"context"
	"fmt"
	"sync"

	"github.com/grandcat/zeroconf"
)

type browser interface {
	Browse(context.Context, string, string, chan<- *zeroconf.ServiceEntry) error
}

type resolverFactory func() (browser, error)

type Service struct {
	resolverFactory resolverFactory
}

func NewService() *Service {
	return &Service{
		resolverFactory: func() (browser, error) {
			return zeroconf.NewResolver(nil)
		},
	}
}

func newServiceWithResolverFactory(factory resolverFactory) *Service {
	return &Service{resolverFactory: factory}
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

	run := func(serviceType string) {
		defer wg.Done()

		found, err := service.discoverService(ctx, serviceType)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s discovery failed: %v", serviceType, err))
			return
		}
		devices = append(devices, found...)
	}

	for _, serviceType := range []string{ServiceOperational, ServiceCommissioning, ServiceExtendedSetup} {
		wg.Add(1)
		go run(serviceType)
	}
	wg.Wait()

	sortDevices(devices)
	return DiscoverResult{
		Devices:  dedupeDevices(devices),
		Warnings: uniqueSorted(warnings),
	}, nil
}

func dedupeDevices(devices []Device) []Device {
	seen := make(map[string]struct{}, len(devices))
	out := make([]Device, 0, len(devices))
	for _, device := range devices {
		key := device.Service + "|" + device.Instance + "|" + device.Host + fmt.Sprintf("|%d", device.Port)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, device)
	}
	return out
}
