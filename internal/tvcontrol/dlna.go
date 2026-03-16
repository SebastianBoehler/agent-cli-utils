package tvcontrol

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func playDLNA(ctx context.Context, client *http.Client, device Device, mediaURL string) (ActionResult, error) {
	status, detail, err := dlnaSOAP(ctx, client, device.ControlURL, "SetAVTransportURI", map[string]string{
		"InstanceID":         "0",
		"CurrentURI":         mediaURL,
		"CurrentURIMetaData": "",
	})
	if err != nil {
		return ActionResult{}, err
	}
	if status >= 400 {
		return ActionResult{
			Operation:  "play",
			Protocol:   ProtocolDLNA,
			Target:     device.Name,
			URL:        mediaURL,
			OK:         false,
			HTTPStatus: status,
			Detail:     detail,
			Device:     &device,
		}, nil
	}

	status, detail, err = dlnaSOAP(ctx, client, device.ControlURL, "Play", map[string]string{
		"InstanceID": "0",
		"Speed":      "1",
	})
	if err != nil {
		return ActionResult{}, err
	}

	return ActionResult{
		Operation:  "play",
		Protocol:   ProtocolDLNA,
		Target:     device.Name,
		URL:        mediaURL,
		OK:         status < 400,
		HTTPStatus: status,
		Detail:     detail,
		Device:     &device,
	}, nil
}

func stopDLNA(ctx context.Context, client *http.Client, device Device) (ActionResult, error) {
	status, detail, err := dlnaSOAP(ctx, client, device.ControlURL, "Stop", map[string]string{
		"InstanceID": "0",
	})
	if err != nil {
		return ActionResult{}, err
	}

	return ActionResult{
		Operation:  "stop",
		Protocol:   ProtocolDLNA,
		Target:     device.Name,
		OK:         status < 400,
		HTTPStatus: status,
		Detail:     detail,
		Device:     &device,
	}, nil
}

func dlnaSOAP(ctx context.Context, client *http.Client, endpoint string, action string, args map[string]string) (int, string, error) {
	body := soapEnvelope(action, args)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(body))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", fmt.Sprintf(`"%s#%s"`, avTransportService, action))

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, strings.TrimSpace(string(payload)), nil
}

func soapEnvelope(action string, args map[string]string) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	builder.WriteString(`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body>`)
	builder.WriteString(`<u:` + action + ` xmlns:u="` + avTransportService + `">`)
	keys := []string{"InstanceID", "CurrentURI", "CurrentURIMetaData", "Speed"}
	for _, key := range keys {
		value, ok := args[key]
		if !ok {
			continue
		}
		builder.WriteString("<" + key + ">")
		builder.WriteString(xmlEscape(value))
		builder.WriteString("</" + key + ">")
	}
	builder.WriteString(`</u:` + action + `></s:Body></s:Envelope>`)
	return builder.String()
}

func xmlEscape(value string) string {
	var out bytes.Buffer
	_ = xml.EscapeText(&out, []byte(value))
	return out.String()
}
