package tvcontrol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/SebastianBoehler/agent-cli-utils/internal/devicestore"
)

var appleCredentialPattern = regexp.MustCompile(`You may now use these credentials:\s+([^\s]+)`)

func (service *Service) PairApple(ctx context.Context, options PairOptions) (PairResult, error) {
	host := strings.TrimSpace(options.Host)
	if host == "" {
		return PairResult{}, fmt.Errorf("host is required")
	}
	bin, err := exec.LookPath("atvremote")
	if err != nil {
		return PairResult{}, fmt.Errorf("atvremote is required for experimental apple pairing")
	}
	args := []string{"--scan-hosts", host, "--protocol", "airplay"}
	if strings.TrimSpace(options.PIN) != "" {
		args = append(args, "-p", strings.TrimSpace(options.PIN))
	}
	args = append(args, "pair")
	cmd := exec.CommandContext(ctx, bin, args...)
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return PairResult{}, fmt.Errorf(strings.TrimSpace(combined.String()))
	}
	matches := appleCredentialPattern.FindStringSubmatch(combined.String())
	if len(matches) != 2 {
		return PairResult{}, fmt.Errorf("pairing succeeded but credentials were not found in atvremote output")
	}
	if err := devicestore.SaveAppleCredentials(host, matches[1]); err != nil {
		return PairResult{}, err
	}
	return PairResult{Protocol: ProtocolAirPlay, Target: host, OK: true, Detail: "pairing accepted and credentials stored", Token: matches[1]}, nil
}

func (service *Service) playAppleBridge(ctx context.Context, host string, mediaURL string) (ActionResult, error) {
	python, err := exec.LookPath("python3")
	if err != nil {
		return ActionResult{}, fmt.Errorf("python3 is required for experimental apple playback")
	}
	credentials, err := devicestore.LoadAppleCredentials(host)
	if err != nil {
		return ActionResult{}, err
	}
	if credentials == "" {
		return ActionResult{}, fmt.Errorf("no airplay credentials stored for %s; run pair first", host)
	}

	scriptPath, err := writeAppleHelper()
	if err != nil {
		return ActionResult{}, err
	}
	defer os.Remove(scriptPath)

	cmd := exec.CommandContext(
		ctx,
		python,
		scriptPath,
		"play",
		"--host", host,
		"--credentials", credentials,
		"--url", mediaURL,
		"--hold-seconds", "20",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	var helper struct {
		Detail           string  `json:"detail"`
		HeldSeconds      float64 `json:"held_seconds"`
		Mode             string  `json:"mode"`
		OK               bool    `json:"ok"`
		PlayResponseCode *int    `json:"play_response_code"`
	}
	if parseErr := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &helper); parseErr != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail == "" && err != nil {
			detail = err.Error()
		}
		return ActionResult{}, fmt.Errorf("apple helper failed: %s", firstNonEmpty(detail, "unknown error"))
	}

	if err != nil && strings.TrimSpace(stderr.String()) != "" {
		helper.Detail = strings.TrimSpace(helper.Detail + " " + strings.TrimSpace(stderr.String()))
	}

	return ActionResult{
		Operation: "play",
		Protocol:  ProtocolAirPlay,
		Target:    host,
		URL:       mediaURL,
		OK:        helper.OK,
		Detail:    firstNonEmpty(helper.Detail, "apple helper completed"),
	}, nil
}
