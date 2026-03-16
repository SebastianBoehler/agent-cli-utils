package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
	"github.com/SebastianBoehler/agent-cli-utils/internal/tvcontrol"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	service := tvcontrol.NewService(nil)
	switch os.Args[1] {
	case "discover":
		runDiscover(service, os.Args[2:])
	case "play":
		runPlay(service, os.Args[2:])
	case "stop":
		runStop(service, os.Args[2:])
	case "wake":
		runWake(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fail(fmt.Errorf("unknown subcommand %q", os.Args[1]))
	}
}

func runDiscover(service *tvcontrol.Service, args []string) {
	flags := newFlagSet("discover")
	timeout := flags.Duration("timeout", 4*time.Second, "discovery timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Discover(context.Background(), tvcontrol.DiscoverOptions{Timeout: *timeout})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return tvcontrol.RenderDiscoverText(result) })
}

func runPlay(service *tvcontrol.Service, args []string) {
	flags := newFlagSet("play")
	device := flags.String("device", "", "discovered device name, id, or host fragment")
	host := flags.String("host", "", "direct AirPlay host[:port]")
	controlURL := flags.String("control-url", "", "direct DLNA AVTransport control URL")
	protocol := flags.String("protocol", "auto", "auto, airplay, or dlna")
	url := flags.String("url", "", "HTTP(S) media URL to hand off to the TV")
	startPosition := flags.Float64("start-position", 0, "AirPlay playback start position in seconds fraction, usually 0-1")
	timeout := flags.Duration("timeout", 8*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Play(context.Background(), tvcontrol.PlayOptions{
		Device:        *device,
		Host:          *host,
		ControlURL:    *controlURL,
		Protocol:      *protocol,
		URL:           *url,
		StartPosition: *startPosition,
		Timeout:       *timeout,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return tvcontrol.RenderActionText(result) })
	if !result.OK {
		os.Exit(1)
	}
}

func runStop(service *tvcontrol.Service, args []string) {
	flags := newFlagSet("stop")
	device := flags.String("device", "", "discovered device name, id, or host fragment")
	host := flags.String("host", "", "direct AirPlay host[:port]")
	controlURL := flags.String("control-url", "", "direct DLNA AVTransport control URL")
	protocol := flags.String("protocol", "auto", "auto, airplay, or dlna")
	timeout := flags.Duration("timeout", 8*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Stop(context.Background(), tvcontrol.StopOptions{
		Device:     *device,
		Host:       *host,
		ControlURL: *controlURL,
		Protocol:   *protocol,
		Timeout:    *timeout,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return tvcontrol.RenderActionText(result) })
	if !result.OK {
		os.Exit(1)
	}
}

func runWake(args []string) {
	flags := newFlagSet("wake")
	mac := flags.String("mac", "", "TV MAC address for Wake-on-LAN")
	broadcast := flags.String("broadcast", "255.255.255.255:9", "broadcast host:port for the magic packet")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := tvcontrol.Wake(tvcontrol.WakeOptions{
		MAC:       *mac,
		Broadcast: *broadcast,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return tvcontrol.RenderWakeText(result) })
	if !result.OK {
		os.Exit(1)
	}
}

func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	return flags
}

func parseFlags(flags *flag.FlagSet, args []string) {
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}
}

func writeOutput(format string, value any, renderText func() string) {
	switch format {
	case "json", "yaml":
		if err := output.Write(format, value); err != nil {
			fail(err)
		}
	case "text":
		fmt.Print(renderText())
	default:
		fail(fmt.Errorf("unsupported format %q", format))
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: agenttv <discover|play|stop|wake> [flags]")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
