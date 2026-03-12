package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
	"github.com/SebastianBoehler/agent-cli-utils/internal/runwrap"
)

func main() {
	timeout := flag.Duration("timeout", 30*time.Second, "execution timeout")
	maxOutput := flag.Int("max-output", 64*1024, "max captured bytes for stdout and stderr each")
	passStdin := flag.Bool("stdin", false, "forward stdin to the child process")
	workingDir := flag.String("dir", "", "working directory for the command")
	format := flag.String("format", "json", "json, yaml, or text")
	flag.Parse()

	command := flag.Args()
	if len(command) == 0 {
		fail(fmt.Errorf("provide a command after flags"))
	}

	result, err := runwrap.Run(command, runwrap.Options{
		Timeout:        *timeout,
		MaxOutputBytes: *maxOutput,
		PassStdin:      *passStdin,
		WorkingDir:     *workingDir,
	})
	if err != nil {
		fail(err)
	}

	switch *format {
	case "json", "yaml":
		if err := output.Write(*format, result); err != nil {
			fail(err)
		}
	case "text":
		fmt.Print(runwrap.RenderText(result))
	default:
		fail(fmt.Errorf("unsupported format %q", *format))
	}

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
