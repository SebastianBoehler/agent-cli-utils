package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/company"
	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
)

type service interface {
	Search(ctx context.Context, selection string, req company.SearchRequest) (company.SearchReport, error)
	Quota(ctx context.Context) (company.QuotaReport, error)
}

type serviceFactory func(cfg company.Config) service

func main() {
	exitCode, err := run(os.Args[1:], os.Stdout, os.Getenv, func(cfg company.Config) service {
		return company.NewService(cfg)
	})
	if err != nil {
		fail(err)
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func run(args []string, stdout io.Writer, getenv func(string) string, factory serviceFactory) (int, error) {
	if len(args) == 0 {
		return 1, fmt.Errorf("provide a subcommand: search or quota")
	}

	switch args[0] {
	case "search":
		return runSearch(args[1:], stdout, getenv, factory)
	case "quota":
		return runQuota(args[1:], stdout, getenv, factory)
	default:
		return 1, fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func runSearch(args []string, stdout io.Writer, getenv func(string) string, factory serviceFactory) (int, error) {
	flags := flag.NewFlagSet("search", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	city := flags.String("city", "", "company seat or city filter")
	country := flags.String("country", "de", "two-letter country code")
	limit := flags.Int("limit", 10, "maximum results to return")
	source := flags.String("source", company.SourceAll, "all, handelsregister, offeneregister, or opencorporates")
	exact := flags.Bool("exact", false, "match exact company names when supported")
	format := flags.String("format", "json", "json, yaml, or text")
	timeout := flags.Duration("timeout", 30*time.Second, "request timeout")
	apiToken := flags.String("opencorporates-api-token", firstEnv(getenv, "OPENCORPORATES_API_TOKEN"), "OpenCorporates API token")
	opencorporatesURL := flags.String("opencorporates-url", "", "override OpenCorporates API base URL")
	handelsregisterURL := flags.String("handelsregister-url", "", "override Handelsregister base URL")
	offeneRegisterURL := flags.String("offeneregister-url", "", "override OffeneRegister discovery URL")
	offeneRegisterDBURL := flags.String("offeneregister-db-url", "", "override OffeneRegister Datasette base URL")
	offeneRegisterSlug := flags.String("offeneregister-slug", firstEnv(getenv, "OFFENEREGISTER_SLUG"), "override OffeneRegister Datasette slug")
	if err := flags.Parse(args); err != nil {
		return 1, err
	}

	query := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if query == "" {
		return 1, fmt.Errorf("provide a company query")
	}

	svc := factory(company.Config{
		Timeout:                *timeout,
		OpenCorporatesAPIToken: *apiToken,
		OpenCorporatesURL:      *opencorporatesURL,
		HandelsregisterURL:     *handelsregisterURL,
		OffeneRegisterURL:      *offeneRegisterURL,
		OffeneRegisterDBURL:    *offeneRegisterDBURL,
		OffeneRegisterSlug:     *offeneRegisterSlug,
	})

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	report, err := svc.Search(ctx, *source, company.SearchRequest{
		Query:   query,
		City:    *city,
		Country: *country,
		Exact:   *exact,
		Limit:   *limit,
	})
	if err != nil {
		return 1, err
	}
	if err := writeSearch(stdout, *format, report); err != nil {
		return 1, err
	}
	if report.AllSourcesFailed() {
		return 1, nil
	}
	return 0, nil
}

func runQuota(args []string, stdout io.Writer, getenv func(string) string, factory serviceFactory) (int, error) {
	flags := flag.NewFlagSet("quota", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	format := flags.String("format", "json", "json, yaml, or text")
	timeout := flags.Duration("timeout", 30*time.Second, "request timeout")
	apiToken := flags.String("opencorporates-api-token", firstEnv(getenv, "OPENCORPORATES_API_TOKEN"), "OpenCorporates API token")
	opencorporatesURL := flags.String("opencorporates-url", "", "override OpenCorporates API base URL")
	if err := flags.Parse(args); err != nil {
		return 1, err
	}

	svc := factory(company.Config{
		Timeout:                *timeout,
		OpenCorporatesAPIToken: *apiToken,
		OpenCorporatesURL:      *opencorporatesURL,
	})

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	report, err := svc.Quota(ctx)
	if err != nil {
		return 1, err
	}
	if err := writeQuota(stdout, *format, report); err != nil {
		return 1, err
	}
	return 0, nil
}

func writeSearch(stdout io.Writer, format string, report company.SearchReport) error {
	switch format {
	case "json", "yaml":
		return output.WriteTo(stdout, format, report)
	case "text":
		_, err := fmt.Fprint(stdout, company.RenderSearchText(report))
		return err
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func writeQuota(stdout io.Writer, format string, report company.QuotaReport) error {
	switch format {
	case "json", "yaml":
		return output.WriteTo(stdout, format, report)
	case "text":
		_, err := fmt.Fprint(stdout, company.RenderQuotaText(report))
		return err
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func firstEnv(getenv func(string) string, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
