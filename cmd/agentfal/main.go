package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/cloudapi"
	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
)

func main() {
	if len(os.Args) < 2 {
		fail(fmt.Errorf("provide a subcommand: submit, status, result, or cancel"))
	}

	switch os.Args[1] {
	case "submit":
		runSubmit(os.Args[2:])
	case "status":
		runStatus(os.Args[2:])
	case "result":
		runResult(os.Args[2:])
	case "cancel":
		runCancel(os.Args[2:])
	default:
		fail(fmt.Errorf("unknown subcommand %q", os.Args[1]))
	}
}

func runSubmit(args []string) {
	flags := flag.NewFlagSet("submit", flag.ExitOnError)
	apiKey := flags.String("api-key", firstEnv("FAL_KEY", "FAL_API_KEY"), "fal API key")
	baseURL := flags.String("base-url", firstEnv("FAL_BASE_URL"), "fal queue API base URL")
	model := flags.String("model", firstEnv("FAL_MODEL"), "fal model path, for example fal-ai/flux/dev")
	inputPath := flags.String("input", "", "read JSON or YAML payload from file, - for stdin")
	inlinePayload := flags.String("payload", "", "inline JSON or YAML payload")
	timeout := flags.Duration("timeout", 60*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	flags.Parse(args)

	payload, err := cloudapi.LoadStructuredInput(*inputPath, *inlinePayload)
	if err != nil {
		fail(err)
	}

	client, err := cloudapi.NewFalClient(cloudapi.NewHTTPClient(*timeout), *apiKey, *model, *baseURL)
	if err != nil {
		fail(err)
	}

	result, err := client.Submit(payload)
	if err != nil {
		fail(err)
	}
	writeResult(result, *format)
}

func runStatus(args []string) {
	flags := flag.NewFlagSet("status", flag.ExitOnError)
	apiKey := flags.String("api-key", firstEnv("FAL_KEY", "FAL_API_KEY"), "fal API key")
	baseURL := flags.String("base-url", firstEnv("FAL_BASE_URL"), "fal queue API base URL")
	model := flags.String("model", firstEnv("FAL_MODEL"), "fal model path, for example fal-ai/flux/dev")
	requestID := flags.String("request", "", "fal request ID")
	includeLogs := flags.Bool("logs", false, "include logs in the status response when supported")
	timeout := flags.Duration("timeout", 30*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	flags.Parse(args)

	client, err := cloudapi.NewFalClient(cloudapi.NewHTTPClient(*timeout), *apiKey, *model, *baseURL)
	if err != nil {
		fail(err)
	}

	result, err := client.Status(*requestID, *includeLogs)
	if err != nil {
		fail(err)
	}
	writeResult(result, *format)
}

func runResult(args []string) {
	flags := flag.NewFlagSet("result", flag.ExitOnError)
	apiKey := flags.String("api-key", firstEnv("FAL_KEY", "FAL_API_KEY"), "fal API key")
	baseURL := flags.String("base-url", firstEnv("FAL_BASE_URL"), "fal queue API base URL")
	model := flags.String("model", firstEnv("FAL_MODEL"), "fal model path, for example fal-ai/flux/dev")
	requestID := flags.String("request", "", "fal request ID")
	timeout := flags.Duration("timeout", 30*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	flags.Parse(args)

	client, err := cloudapi.NewFalClient(cloudapi.NewHTTPClient(*timeout), *apiKey, *model, *baseURL)
	if err != nil {
		fail(err)
	}

	result, err := client.Result(*requestID)
	if err != nil {
		fail(err)
	}
	writeResult(result, *format)
}

func runCancel(args []string) {
	flags := flag.NewFlagSet("cancel", flag.ExitOnError)
	apiKey := flags.String("api-key", firstEnv("FAL_KEY", "FAL_API_KEY"), "fal API key")
	baseURL := flags.String("base-url", firstEnv("FAL_BASE_URL"), "fal queue API base URL")
	model := flags.String("model", firstEnv("FAL_MODEL"), "fal model path, for example fal-ai/flux/dev")
	requestID := flags.String("request", "", "fal request ID")
	timeout := flags.Duration("timeout", 30*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	flags.Parse(args)

	client, err := cloudapi.NewFalClient(cloudapi.NewHTTPClient(*timeout), *apiKey, *model, *baseURL)
	if err != nil {
		fail(err)
	}

	result, err := client.Cancel(*requestID)
	if err != nil {
		fail(err)
	}
	writeResult(result, *format)
}

func writeResult(result cloudapi.Result, format string) {
	switch format {
	case "json", "yaml":
		if err := output.Write(format, result); err != nil {
			fail(err)
		}
	case "text":
		fmt.Print(cloudapi.RenderText(result))
	default:
		fail(fmt.Errorf("unsupported format %q", format))
	}

	if result.HTTPStatus >= 400 {
		os.Exit(1)
	}
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
