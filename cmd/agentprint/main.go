package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
	"github.com/SebastianBoehler/agent-cli-utils/internal/printx"
)

type repeatedFlag []string

func (items *repeatedFlag) String() string {
	return fmt.Sprint([]string(*items))
}

func (items *repeatedFlag) Set(value string) error {
	*items = append(*items, value)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	service := printx.NewService()
	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "list":
		runList(service, args)
	case "discover":
		runDiscover(service, args)
	case "ensure":
		runEnsure(service, args)
	case "print":
		runPrint(service, args)
	case "-h", "--help", "help":
		usage()
	default:
		fail(fmt.Errorf("unknown command %q", command))
	}
}

func runList(service *printx.Service, args []string) {
	flags := newFlagSet("list")
	format := flags.String("format", "json", "json, yaml, text, or tsv")
	parseFlags(flags, args)

	result, err := service.List(context.Background())
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return printx.RenderListText(result) }, func() string { return printx.RenderListTSV(result) })
}

func runDiscover(service *printx.Service, args []string) {
	flags := newFlagSet("discover")
	format := flags.String("format", "json", "json, yaml, text, or tsv")
	parseFlags(flags, args)

	result, err := service.Discover(context.Background())
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return printx.RenderDiscoverText(result) }, func() string { return printx.RenderDiscoverTSV(result) })
}

func runEnsure(service *printx.Service, args []string) {
	flags := newFlagSet("ensure")
	printer := flags.String("printer", "", "local CUPS queue name to create or repair")
	uri := flags.String("uri", "", "explicit printer URI")
	match := flags.String("match", "", "discovered printer name or URI fragment")
	makeDefault := flags.Bool("default", false, "set the queue as the CUPS default")
	description := flags.String("description", "", "queue description")
	location := flags.String("location", "", "queue location")
	format := flags.String("format", "json", "json, yaml, text, or tsv")
	parseFlags(flags, args)

	result, err := service.Ensure(context.Background(), printx.EnsureOptions{
		QueueName:   *printer,
		URI:         *uri,
		Match:       *match,
		MakeDefault: *makeDefault,
		Description: *description,
		Location:    *location,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return printx.RenderEnsureText(result) }, func() string { return printx.RenderEnsureTSV(result) })
}

func runPrint(service *printx.Service, args []string) {
	flags := newFlagSet("print")
	var rawOptions repeatedFlag

	printer := flags.String("printer", "", "target CUPS queue")
	input := flags.String("input", "", "local file path or HTTP(S) URL")
	copies := flags.Int("copies", 1, "number of copies")
	duplex := flags.Bool("duplex", false, "enable long-edge duplex")
	media := flags.String("media", "", "media size such as A4 or Letter")
	orientation := flags.String("orientation", "", "portrait, landscape, reverse-landscape, or reverse-portrait")
	position := flags.String("position", "", "raw CUPS position value")
	scalePercent := flags.Int("scale-percent", 0, "scale percentage")
	fitToPage := flags.Bool("fit-to-page", false, "scale content to fit the page")
	fillPage := flags.Bool("fill-page", false, "scale content to fill the page")
	colorMode := flags.String("color-mode", "auto", "auto, color, monochrome, or bi-level")
	jobName := flags.String("job-name", "", "optional lp job title")
	format := flags.String("format", "json", "json, yaml, text, or tsv")
	flags.Var(&rawOptions, "option", "raw CUPS option name=value; repeatable")
	parseFlags(flags, args)

	if *input == "" && flags.NArg() == 1 {
		*input = flags.Arg(0)
	}
	if flags.NArg() > 1 {
		fail(fmt.Errorf("print accepts at most one positional input"))
	}

	result, err := service.Print(context.Background(), printx.PrintOptions{
		Printer:      *printer,
		Source:       *input,
		Copies:       *copies,
		Duplex:       *duplex,
		Media:        *media,
		Orientation:  *orientation,
		Position:     *position,
		ScalePercent: *scalePercent,
		FitToPage:    *fitToPage,
		FillPage:     *fillPage,
		ColorMode:    *colorMode,
		RawOptions:   rawOptions,
		JobName:      *jobName,
	})
	if err != nil {
		fail(err)
	}
	writeOutput(*format, result, func() string { return printx.RenderPrintText(result) }, func() string { return printx.RenderPrintTSV(result) })
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

func writeOutput(format string, value any, renderText func() string, renderTSV func() string) {
	switch format {
	case "json", "yaml":
		if err := output.Write(format, value); err != nil {
			fail(err)
		}
	case "text":
		fmt.Print(renderText())
	case "tsv":
		fmt.Print(renderTSV())
	default:
		fail(fmt.Errorf("unsupported format %q", format))
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: agentprint <list|discover|ensure|print> [flags]")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
