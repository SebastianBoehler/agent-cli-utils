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
		runStatus("status", os.Args[2:])
	case "result":
		runStatus("result", os.Args[2:])
	case "cancel":
		runCancel(os.Args[2:])
	default:
		fail(fmt.Errorf("unknown subcommand %q", os.Args[1]))
	}
}

func runSubmit(args []string) {
	flags := flag.NewFlagSet("submit", flag.ExitOnError)
	apiKey := flags.String("api-key", firstEnv("RUNPOD_API_KEY"), "RunPod API key")
	baseURL := flags.String("base-url", firstEnv("RUNPOD_BASE_URL"), "RunPod API base URL")
	endpoint := flags.String("endpoint", firstEnv("RUNPOD_ENDPOINT_ID"), "RunPod endpoint ID")
	inputPath := flags.String("input", "", "read JSON or YAML payload from file, - for stdin")
	inlinePayload := flags.String("payload", "", "inline JSON or YAML payload")
	rawBody := flags.Bool("raw-body", false, "send payload as-is instead of wrapping it in {input: ...}")
	syncMode := flags.Bool("sync", false, "use the synchronous /runsync endpoint")
	timeout := flags.Duration("timeout", 60*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	flags.Parse(args)

	payload, err := cloudapi.LoadStructuredInput(*inputPath, *inlinePayload)
	if err != nil {
		fail(err)
	}
	if !*rawBody {
		if payload == nil {
			payload = map[string]any{}
		}
		payload = map[string]any{"input": payload}
	}

	client, err := cloudapi.NewRunPodClient(cloudapi.NewHTTPClient(*timeout), *apiKey, *endpoint, *baseURL)
	if err != nil {
		fail(err)
	}

	result, err := client.Submit(payload, *syncMode)
	if err != nil {
		fail(err)
	}
	writeResult(result, *format)
}

func runStatus(operation string, args []string) {
	flags := flag.NewFlagSet(operation, flag.ExitOnError)
	apiKey := flags.String("api-key", firstEnv("RUNPOD_API_KEY"), "RunPod API key")
	baseURL := flags.String("base-url", firstEnv("RUNPOD_BASE_URL"), "RunPod API base URL")
	endpoint := flags.String("endpoint", firstEnv("RUNPOD_ENDPOINT_ID"), "RunPod endpoint ID")
	requestID := flags.String("request", "", "RunPod job ID")
	timeout := flags.Duration("timeout", 30*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	flags.Parse(args)

	client, err := cloudapi.NewRunPodClient(cloudapi.NewHTTPClient(*timeout), *apiKey, *endpoint, *baseURL)
	if err != nil {
		fail(err)
	}

	var result cloudapi.Result
	if operation == "result" {
		result, err = client.Result(*requestID)
	} else {
		result, err = client.Status(*requestID)
	}
	if err != nil {
		fail(err)
	}
	writeResult(result, *format)
}

func runCancel(args []string) {
	flags := flag.NewFlagSet("cancel", flag.ExitOnError)
	apiKey := flags.String("api-key", firstEnv("RUNPOD_API_KEY"), "RunPod API key")
	baseURL := flags.String("base-url", firstEnv("RUNPOD_BASE_URL"), "RunPod API base URL")
	endpoint := flags.String("endpoint", firstEnv("RUNPOD_ENDPOINT_ID"), "RunPod endpoint ID")
	requestID := flags.String("request", "", "RunPod job ID")
	timeout := flags.Duration("timeout", 30*time.Second, "request timeout")
	format := flags.String("format", "json", "json, yaml, or text")
	flags.Parse(args)

	client, err := cloudapi.NewRunPodClient(cloudapi.NewHTTPClient(*timeout), *apiKey, *endpoint, *baseURL)
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
