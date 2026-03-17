package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
	"github.com/SebastianBoehler/agent-cli-utils/internal/samsungcontrol"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	service := samsungcontrol.NewService()
	switch os.Args[1] {
	case "pair":
		runPair(service, os.Args[2:])
	case "remote":
		runRemote(service, os.Args[2:])
	case "launch":
		runLaunch(service, os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fail(fmt.Errorf("unknown subcommand %q", os.Args[1]))
	}
}

func runPair(service *samsungcontrol.Service, args []string) {
	flags := newFlagSet("pair")
	host := flags.String("host", "", "samsung tv host")
	name := flags.String("name", "agentsamsung", "remote/pairing name")
	timeout := flags.Duration("timeout", 20*time.Second, "pairing timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Pair(context.Background(), samsungcontrol.PairOptions{
		Host:    *host,
		Name:    *name,
		Timeout: *timeout,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return samsungcontrol.RenderPairText(result) })
}

func runRemote(service *samsungcontrol.Service, args []string) {
	flags := newFlagSet("remote")
	host := flags.String("host", "", "samsung tv host")
	key := flags.String("key", "", "Samsung remote key such as KEY_HOME or KEY_VOLUP")
	timeout := flags.Duration("timeout", 8*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Remote(context.Background(), samsungcontrol.RemoteOptions{
		Host:    *host,
		Key:     *key,
		Timeout: *timeout,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return samsungcontrol.RenderActionText(result) })
}

func runLaunch(service *samsungcontrol.Service, args []string) {
	flags := newFlagSet("launch")
	host := flags.String("host", "", "samsung tv host")
	appID := flags.String("app-id", "", "Samsung app id such as 111299001912 for YouTube")
	appType := flags.String("app-type", "DEEP_LINK", "DEEP_LINK or NATIVE_LAUNCH")
	metaTag := flags.String("meta-tag", "", "optional launch metadata")
	timeout := flags.Duration("timeout", 8*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Launch(context.Background(), samsungcontrol.LaunchOptions{
		Host:    *host,
		AppID:   *appID,
		AppType: *appType,
		MetaTag: *metaTag,
		Timeout: *timeout,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return samsungcontrol.RenderActionText(result) })
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
	fmt.Fprintln(os.Stderr, "usage: agentsamsung <pair|remote|launch> [flags]")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
