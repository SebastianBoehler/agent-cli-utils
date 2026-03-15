package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
	"github.com/SebastianBoehler/agent-cli-utils/internal/smtpx"
)

type repeatedFlag []string

func (items *repeatedFlag) String() string {
	return fmt.Sprint([]string(*items))
}

func (items *repeatedFlag) Set(value string) error {
	*items = append(*items, value)
	return nil
}

type commonFlags struct {
	config       string
	provider     string
	host         string
	port         int
	security     string
	auth         string
	username     string
	password     string
	passwordEnv  string
	passwordFile string
	from         string
	timeout      time.Duration
	format       string
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "profile":
		runProfile(args)
	case "test":
		runTest(args)
	case "send":
		runSend(args)
	case "-h", "--help", "help":
		usage()
	default:
		fail(fmt.Errorf("unknown command %q", command))
	}
}

func runProfile(args []string) {
	flags, common := newCommonFlagSet("profile")
	parseFlags(flags, args)

	config := resolveConfig(common)
	profile := smtpx.Inspect(config)
	writeOutput(common.format, profile, func() string { return smtpx.RenderProfileText(profile) })
}

func runTest(args []string) {
	flags, common := newCommonFlagSet("test")
	parseFlags(flags, args)

	config := resolveConfig(common)
	result, err := smtpx.Check(config)
	if err != nil {
		fail(err)
	}
	writeOutput(common.format, result, func() string { return smtpx.RenderResultText(result) })
}

func runSend(args []string) {
	flags, common := newCommonFlagSet("send")
	var to, cc, bcc repeatedFlag

	subject := flags.String("subject", "", "message subject")
	text := flags.String("text", "", "inline plain-text body")
	textFile := flags.String("text-file", "", "load plain-text body from file, or - for stdin")
	flags.Var(&to, "to", "recipient email address; repeatable")
	flags.Var(&cc, "cc", "cc email address; repeatable")
	flags.Var(&bcc, "bcc", "bcc email address; repeatable")
	parseFlags(flags, args)

	body, err := smtpx.ReadBody(*text, *textFile)
	if err != nil {
		fail(err)
	}

	config := resolveConfig(common)
	message := smtpx.Message{
		To:      to,
		Cc:      cc,
		Bcc:     bcc,
		Subject: strings.TrimSpace(*subject),
		Text:    body,
	}

	result, err := smtpx.Send(config, message)
	if err != nil {
		fail(err)
	}
	writeOutput(common.format, result, func() string { return smtpx.RenderResultText(result) })
}

func newCommonFlagSet(name string) (*flag.FlagSet, *commonFlags) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	common := &commonFlags{}
	flags.StringVar(&common.config, "config", "", "load config from YAML or JSON file")
	flags.StringVar(&common.provider, "provider", "", "auto, generic, gmail, or google-workspace")
	flags.StringVar(&common.host, "host", "", "smtp host")
	flags.IntVar(&common.port, "port", 0, "smtp port")
	flags.StringVar(&common.security, "security", "", "starttls, tls, or plain")
	flags.StringVar(&common.auth, "auth", "", "password or none")
	flags.StringVar(&common.username, "username", "", "smtp username")
	flags.StringVar(&common.password, "password", "", "smtp password or app password")
	flags.StringVar(&common.passwordEnv, "password-env", "", "env var holding the smtp password")
	flags.StringVar(&common.passwordFile, "password-file", "", "file containing the smtp password")
	flags.StringVar(&common.from, "from", "", "envelope and header from address")
	flags.DurationVar(&common.timeout, "timeout", 0, "smtp connect timeout, for example 15s")
	flags.StringVar(&common.format, "format", "json", "json, yaml, or text")
	return flags, common
}

func resolveConfig(common *commonFlags) smtpx.Config {
	fileConfig, err := smtpx.LoadConfig(common.config)
	if err != nil {
		fail(err)
	}

	config, err := smtpx.Resolve(smtpx.Merge(fileConfig, smtpx.EnvConfig(), smtpx.Config{
		Provider:     common.provider,
		Host:         common.host,
		Port:         common.port,
		Security:     common.security,
		Auth:         common.auth,
		Username:     common.username,
		Password:     common.password,
		PasswordEnv:  common.passwordEnv,
		PasswordFile: common.passwordFile,
		From:         common.from,
		Timeout:      common.timeout,
	}))
	if err != nil {
		fail(err)
	}
	return config
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
	fmt.Fprintln(os.Stderr, "usage: agentsmtp <profile|test|send> [flags]")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
