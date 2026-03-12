package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/SebastianBoehler/agent-cli-utils/internal/doctor"
	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
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
	var commands repeatedFlag

	profile := flag.String("profile", "", "built-in profile: smb-client, ssh-client, or web-fetch")
	format := flag.String("format", "json", "json, yaml, or text")
	goos := flag.String("os", runtime.GOOS, "target operating system for profile hints")
	strict := flag.Bool("strict", true, "exit non-zero when a required dependency is missing")
	listProfiles := flag.Bool("list-profiles", false, "print available built-in profiles")
	flag.Var(&commands, "cmd", "extra command to require; repeatable")
	flag.Parse()

	if *listProfiles {
		for _, profileName := range doctor.AvailableProfiles() {
			fmt.Println(profileName)
		}
		return
	}

	if *profile == "" && len(commands) == 0 {
		fail(fmt.Errorf("provide -profile, -cmd, or -list-profiles"))
	}

	report, err := doctor.Evaluate(*profile, commands, *goos, exec.LookPath)
	if err != nil {
		fail(err)
	}

	switch *format {
	case "json", "yaml":
		if err := output.Write(*format, report); err != nil {
			fail(err)
		}
	case "text":
		fmt.Print(doctor.RenderText(report))
	default:
		fail(fmt.Errorf("unsupported format %q", *format))
	}

	if *strict && !report.AllRequiredPresent {
		os.Exit(1)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
