package youtubecontrol

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type dialService struct {
	Name    string `xml:"name"`
	State   string `xml:"state"`
	Version string `xml:"version"`
	Options struct {
		AllowStop string `xml:"allowStop,attr"`
	} `xml:"options"`
	Links []dialLink `xml:",any"`
}

type dialLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

func fetchYouTubeStatus(client *http.Client, host string) (Device, int, error) {
	appURL := fmt.Sprintf("http://%s:8080/ws/apps/YouTube", host)
	request, err := http.NewRequest(http.MethodGet, appURL, nil)
	if err != nil {
		return Device{}, 0, err
	}
	response, err := client.Do(request)
	if err != nil {
		return Device{}, 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Device{Host: host, AppURL: appURL}, response.StatusCode, nil
	}

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return Device{}, response.StatusCode, err
	}
	var service dialService
	if err := xml.Unmarshal(payload, &service); err != nil {
		return Device{}, response.StatusCode, err
	}

	device := Device{
		Host:      host,
		AppURL:    appURL,
		State:     strings.TrimSpace(service.State),
		Version:   strings.TrimSpace(service.Version),
		AllowStop: strings.EqualFold(strings.TrimSpace(service.Options.AllowStop), "true"),
	}
	for _, link := range service.Links {
		if strings.EqualFold(link.Rel, "run") {
			device.RunURL = resolveAppURL(appURL, link.Href)
			break
		}
	}
	return device, response.StatusCode, nil
}

func playYouTubeDial(client *http.Client, device Device, videoID string, startOffset string) ActionResult {
	values := url.Values{}
	values.Set("launch", "dial")
	if strings.TrimSpace(videoID) != "" {
		values.Set("v", strings.TrimSpace(videoID))
	}
	if strings.TrimSpace(startOffset) != "" {
		values.Set("t", strings.TrimSpace(startOffset))
	}

	request, err := http.NewRequest(http.MethodPost, device.AppURL, strings.NewReader(values.Encode()))
	if err != nil {
		return ActionResult{Operation: "play", Target: device.Host, VideoID: videoID, Detail: err.Error()}
	}
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	request.ContentLength = int64(len(values.Encode()))

	response, err := client.Do(request)
	if err != nil {
		return ActionResult{Operation: "play", Target: device.Host, VideoID: videoID, Detail: err.Error()}
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)

	ok := response.StatusCode >= 200 && response.StatusCode < 300
	detail := response.Status
	if location := strings.TrimSpace(response.Header.Get("Location")); location != "" {
		detail = firstNonEmpty(detail, "launch accepted") + " location=" + location
	}
	return ActionResult{
		Operation:  "play",
		Target:     device.Host,
		VideoID:    videoID,
		OK:         ok,
		HTTPStatus: response.StatusCode,
		Detail:     detail,
	}
}

func resolveAppURL(appURL string, href string) string {
	base, err := url.Parse(appURL)
	if err != nil {
		return href
	}
	relative, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return href
	}
	return base.ResolveReference(relative).String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
