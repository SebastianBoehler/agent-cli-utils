package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/mattercontrol"
	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	service := mattercontrol.NewService()
	switch os.Args[1] {
	case "discover":
		runDiscover(service, os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fail(fmt.Errorf("unknown subcommand %q", os.Args[1]))
	}
}

func runDiscover(service *mattercontrol.Service, args []string) {
	flags := newFlagSet("discover")
	timeout := flags.Duration("timeout", 4*time.Second, "discovery timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Discover(context.Background(), mattercontrol.DiscoverOptions{Timeout: *timeout})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return mattercontrol.RenderDiscoverText(result) })
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
	fmt.Fprintln(os.Stderr, "usage: agentmatter <discover> [flags]")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
