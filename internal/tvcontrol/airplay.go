package tvcontrol

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func playAirPlay(ctx context.Context, client *http.Client, device Device, mediaURL string, startPosition float64) (ActionResult, error) {
	body := fmt.Sprintf("Content-Location: %s\r\nStart-Position: %.3f\r\n\r\n", mediaURL, startPosition)
	endpoint := strings.TrimRight(device.Location, "/") + "/play"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(body))
	if err != nil {
		return ActionResult{}, err
	}
	req.Header.Set("Content-Type", "text/parameters")

	resp, err := client.Do(req)
	if err != nil {
		return ActionResult{}, err
	}
	defer resp.Body.Close()

	detail, err := readDetail(resp.Body)
	if err != nil {
		return ActionResult{}, err
	}

	return ActionResult{
		Operation:  "play",
		Protocol:   ProtocolAirPlay,
		Target:     device.Name,
		URL:        mediaURL,
		OK:         resp.StatusCode < 400,
		HTTPStatus: resp.StatusCode,
		Detail:     detail,
		Device:     &device,
	}, nil
}

func stopAirPlay(ctx context.Context, client *http.Client, device Device) (ActionResult, error) {
	endpoint := strings.TrimRight(device.Location, "/") + "/stop"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return ActionResult{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return ActionResult{}, err
	}
	defer resp.Body.Close()

	detail, err := readDetail(resp.Body)
	if err != nil {
		return ActionResult{}, err
	}

	return ActionResult{
		Operation:  "stop",
		Protocol:   ProtocolAirPlay,
		Target:     device.Name,
		OK:         resp.StatusCode < 400,
		HTTPStatus: resp.StatusCode,
		Detail:     detail,
		Device:     &device,
	}, nil
}

func readDetail(reader io.Reader) (string, error) {
	payload, err := io.ReadAll(io.LimitReader(reader, 16<<10))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(payload)), nil
}
