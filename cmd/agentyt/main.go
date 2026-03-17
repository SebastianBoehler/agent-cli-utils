package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
	"github.com/SebastianBoehler/agent-cli-utils/internal/youtubecontrol"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	service := youtubecontrol.NewService(nil)
	switch os.Args[1] {
	case "discover":
		runDiscover(service, os.Args[2:])
	case "search":
		runSearch(service, os.Args[2:])
	case "status":
		runStatus(service, os.Args[2:])
	case "play":
		runPlay(service, os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fail(fmt.Errorf("unknown subcommand %q", os.Args[1]))
	}
}

func runSearch(service *youtubecontrol.Service, args []string) {
	flags := newFlagSet("search")
	query := flags.String("q", "", "YouTube search query")
	language := flags.String("lang", "", "relevance language, e.g. en or de")
	region := flags.String("region", "", "region code, e.g. US or DE")
	duration := flags.String("duration", "", "any, short, medium, or long")
	caption := flags.String("caption", "", "any, closedCaption, or none")
	order := flags.String("order", "", "relevance, date, rating, title, videoCount, or viewCount")
	safeSearch := flags.String("safe-search", "", "moderate, none, or strict")
	maxResults := flags.Int("max-results", 10, "1-50 results")
	timeout := flags.Duration("timeout", 10*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Search(context.Background(), youtubecontrol.SearchOptions{
		Query:      *query,
		Language:   *language,
		Region:     *region,
		Duration:   *duration,
		Caption:    *caption,
		Order:      *order,
		SafeSearch: *safeSearch,
		MaxResults: *maxResults,
		Timeout:    *timeout,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return youtubecontrol.RenderSearchText(result) })
}

func runDiscover(service *youtubecontrol.Service, args []string) {
	flags := newFlagSet("discover")
	timeout := flags.Duration("timeout", 5*time.Second, "discovery timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Discover(context.Background(), youtubecontrol.DiscoverOptions{Timeout: *timeout})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return youtubecontrol.RenderDiscoverText(result) })
}

func runStatus(service *youtubecontrol.Service, args []string) {
	flags := newFlagSet("status")
	device := flags.String("device", "", "discovered receiver name or host fragment")
	host := flags.String("host", "", "direct receiver host")
	timeout := flags.Duration("timeout", 5*time.Second, "status timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Status(context.Background(), youtubecontrol.StatusOptions{
		Device:  *device,
		Host:    *host,
		Timeout: *timeout,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return youtubecontrol.RenderStatusText(result) })
}

func runPlay(service *youtubecontrol.Service, args []string) {
	flags := newFlagSet("play")
	device := flags.String("device", "", "discovered receiver name or host fragment")
	host := flags.String("host", "", "direct receiver host")
	video := flags.String("video", "", "YouTube video id or URL")
	startOffset := flags.String("start", "", "optional YouTube start offset like 43 or 1m20s")
	timeout := flags.Duration("timeout", 10*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	parseFlags(flags, args)

	result, err := service.Play(context.Background(), youtubecontrol.PlayOptions{
		Device:      *device,
		Host:        *host,
		Video:       *video,
		StartOffset: *startOffset,
		Timeout:     *timeout,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return youtubecontrol.RenderActionText(result) })
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
	fmt.Fprintf(os.Stderr, "usage: agentyt <discover|search|status|play> [flags]\n")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
