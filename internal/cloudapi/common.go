package cloudapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strings"
	"time"

	"github.com/SebastianBoehler/agent-cli-utils/internal/datax"
)

type Result struct {
	Service    string `json:"service" yaml:"service"`
	Operation  string `json:"operation" yaml:"operation"`
	Mode       string `json:"mode,omitempty" yaml:"mode,omitempty"`
	Target     string `json:"target,omitempty" yaml:"target,omitempty"`
	RequestID  string `json:"request_id,omitempty" yaml:"request_id,omitempty"`
	Status     string `json:"status,omitempty" yaml:"status,omitempty"`
	HTTPStatus int    `json:"http_status" yaml:"http_status"`
	URL        string `json:"url" yaml:"url"`
	Data       any    `json:"data,omitempty" yaml:"data,omitempty"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &http.Client{Timeout: timeout}
}

func LoadStructuredInput(inputPath string, inline string) (any, error) {
	switch {
	case inline != "":
		return parseStructured([]byte(inline))
	case inputPath != "":
		payload, err := readInput(inputPath)
		if err != nil {
			return nil, err
		}
		return parseStructured(payload)
	case stdinHasData():
		payload, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		if len(bytes.TrimSpace(payload)) == 0 {
			return nil, nil
		}
		return parseStructured(payload)
	default:
		return nil, nil
	}
}

func DoJSON(client *http.Client, method string, endpoint string, headers map[string]string, body any) (int, any, error) {
	if client == nil {
		client = NewHTTPClient(30 * time.Second)
	}

	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal request body: %w", err)
		}
		requestBody = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, endpoint, requestBody)
	if err != nil {
		return 0, nil, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		if strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("read response: %w", err)
	}

	return resp.StatusCode, parseResponse(payload), nil
}

func RenderText(result Result) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "service: %s\n", result.Service)
	fmt.Fprintf(&builder, "operation: %s\n", result.Operation)
	if result.Mode != "" {
		fmt.Fprintf(&builder, "mode: %s\n", result.Mode)
	}
	if result.Target != "" {
		fmt.Fprintf(&builder, "target: %s\n", result.Target)
	}
	if result.RequestID != "" {
		fmt.Fprintf(&builder, "request_id: %s\n", result.RequestID)
	}
	if result.Status != "" {
		fmt.Fprintf(&builder, "status: %s\n", result.Status)
	}
	fmt.Fprintf(&builder, "http_status: %d\n", result.HTTPStatus)
	fmt.Fprintf(&builder, "url: %s\n", result.URL)
	if result.Error != "" {
		fmt.Fprintf(&builder, "error: %s\n", result.Error)
	}
	if result.Data != nil {
		builder.WriteString("data:\n")
		rendered, err := datax.Render(result.Data, "json")
		if err != nil {
			fmt.Fprintf(&builder, "%v\n", result.Data)
		} else {
			builder.Write(rendered)
		}
	}
	return builder.String()
}

func JoinURL(base string, parts ...string) string {
	out := strings.TrimRight(base, "/")
	for _, part := range parts {
		trimmed := strings.Trim(part, "/")
		if trimmed == "" {
			continue
		}
		out += "/" + trimmed
	}
	return out
}

func AddQuery(endpoint string, values map[string]string) (string, error) {
	if len(values) == 0 {
		return endpoint, nil
	}

	parsed, err := neturl.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse url %q: %w", endpoint, err)
	}
	query := parsed.Query()
	for key, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		query.Set(key, value)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func ExtractString(value any, keys ...string) string {
	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}

	for _, key := range keys {
		item, exists := object[key]
		if !exists {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(item))
		if text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func ExtractError(value any) string {
	if message := ExtractString(value, "error", "message", "detail"); message != "" {
		return message
	}

	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}

	if nested, exists := object["error"]; exists {
		switch typed := nested.(type) {
		case map[string]any:
			if message := ExtractString(typed, "message", "detail", "error"); message != "" {
				return message
			}
		case string:
			return strings.TrimSpace(typed)
		}
	}

	return ""
}

func parseStructured(payload []byte) (any, error) {
	value, _, err := datax.ParseStructured(payload)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func parseResponse(payload []byte) any {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 {
		return nil
	}

	var value any
	if err := json.Unmarshal(trimmed, &value); err == nil {
		return value
	}

	return string(trimmed)
}

func readInput(path string) ([]byte, error) {
	if path == "-" {
		payload, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return payload, nil
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return payload, nil
}

func stdinHasData() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}
