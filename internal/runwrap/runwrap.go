package runwrap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Options struct {
	Timeout        time.Duration
	MaxOutputBytes int
	PassStdin      bool
	WorkingDir     string
}

type Result struct {
	Command         []string `json:"command" yaml:"command"`
	ExitCode        int      `json:"exit_code" yaml:"exit_code"`
	TimedOut        bool     `json:"timed_out" yaml:"timed_out"`
	DurationMS      int64    `json:"duration_ms" yaml:"duration_ms"`
	Stdout          string   `json:"stdout" yaml:"stdout"`
	Stderr          string   `json:"stderr" yaml:"stderr"`
	StdoutTruncated bool     `json:"stdout_truncated" yaml:"stdout_truncated"`
	StderrTruncated bool     `json:"stderr_truncated" yaml:"stderr_truncated"`
	Error           string   `json:"error,omitempty" yaml:"error,omitempty"`
}

type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func Run(command []string, options Options) (Result, error) {
	if len(command) == 0 {
		return Result{}, fmt.Errorf("missing command")
	}

	if options.MaxOutputBytes <= 0 {
		options.MaxOutputBytes = 64 * 1024
	}

	ctx := context.Background()
	cancel := func() {}
	if options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	if options.WorkingDir != "" {
		cmd.Dir = options.WorkingDir
	}
	if options.PassStdin {
		cmd.Stdin = os.Stdin
	}

	stdout := &limitedBuffer{limit: options.MaxOutputBytes}
	stderr := &limitedBuffer{limit: options.MaxOutputBytes}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	runErr := cmd.Run()
	duration := time.Since(start)

	result := Result{
		Command:         append([]string(nil), command...),
		ExitCode:        0,
		TimedOut:        errors.Is(ctx.Err(), context.DeadlineExceeded),
		DurationMS:      duration.Milliseconds(),
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.truncated,
		StderrTruncated: stderr.truncated,
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = runErr.Error()
			if result.ExitCode == 0 {
				result.ExitCode = -1
			}
		}
	}

	if result.TimedOut {
		result.Error = "command timed out"
		if result.ExitCode == -1 || result.ExitCode == 0 {
			result.ExitCode = 124
		}
	}

	return result, nil
}

func RenderText(result Result) string {
	return fmt.Sprintf(
		"command: %v\nexit_code: %d\ntimed_out: %t\nduration_ms: %d\nstdout_truncated: %t\nstderr_truncated: %t\nstdout:\n%s\nstderr:\n%s",
		result.Command,
		result.ExitCode,
		result.TimedOut,
		result.DurationMS,
		result.StdoutTruncated,
		result.StderrTruncated,
		result.Stdout,
		result.Stderr,
	)
}

func (buffer *limitedBuffer) Write(payload []byte) (int, error) {
	if buffer.limit <= 0 {
		buffer.truncated = true
		return len(payload), nil
	}

	remaining := buffer.limit - buffer.buf.Len()
	if remaining <= 0 {
		buffer.truncated = true
		return len(payload), nil
	}

	if len(payload) <= remaining {
		return buffer.buf.Write(payload)
	}

	if _, err := buffer.buf.Write(payload[:remaining]); err != nil {
		return 0, err
	}

	buffer.truncated = true
	return len(payload), nil
}

func (buffer *limitedBuffer) String() string {
	return buffer.buf.String()
}
